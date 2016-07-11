package mysql

import (
	"fmt"
	"strconv"
	"strings"

	// Get a MySQL Database Connection
	"database/sql"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
	// This is apparently how this is included
	_ "github.com/go-sql-driver/mysql"
)

var Schema table.Tables
var alters []string

/*
CREATE TABLE `dogs` (
  `id` int(11) NOT NULL,
  `name` varchar(64) NOT NULL,
  `age` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1


// Dogs Metadata
INSERT INTO management.metadata
(`mdid`,`db`,`property_id`,`parent_id`,`type`,`name`,`exists`)
VALUES
(1,1,"tbl1","","Table","dogs",1),
(2,1,"col1","tbl1","Column","id",1),
(3,1,"col2","tbl1","Column","name",1),
(4,1,"col3","tbl1","Column","age",1),
(5,1,"col4","tbl1","Column","address",1),
(6,1,"pi","tbl1","PrimaryKey","PrimaryKey",1),
(7,1,"sc1","tbl1","Index","idx_id_name",1);

CREATE TABLE `cats` (
  `id` int(11) NOT NULL,
  `name` varchar(64) NOT NULL,
  `age` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1
*/

func hasParameter(s string, param string) (result bool) {
	return strings.Index(s, param) != -1
}

func extractParameter(s string, param string) (result string, err error) {

	param = param + "="

	paramPos := strings.Index(s, param)

	if paramPos != -1 {
		// Trim the front of the string to the end of parameter
		paramString := s[paramPos+len(param):]

		// Check for whitespace - this is not the last parameter in the string
		if strings.Index(paramString, " ") != -1 {
			// Extract the value up to the first whitespace
			result = paramString[:strings.Index(paramString, " ")]
		} else {
			result = paramString
		}
	} else {
		err = fmt.Errorf("Error parsing Parameter: [%s]", param)
	}

	return result, err
}

func parseError(msg string) error {
	return fmt.Errorf("Parse Error MySQL CREATE TABLE: %s", msg)
}

func buildTable(lines []string, tbl *table.Table) (err error) {
	// Process Engine and Charset
	var name string
	var engine string
	var autoinc int64
	var charset string

	var hasMetadata bool
	var md metadata.Metadata

	if len(lines) < 2 {
		return parseError("Invalid table definition")
	}

	// extract the name from the first line
	firstLine := lines[0]

	nameStart := strings.Index(firstLine, "`")
	nameEnd := strings.LastIndex(firstLine, "`")

	if nameStart == -1 || nameStart == nameEnd {
		return parseError("Unable to parse Table Name")
	}
	name = firstLine[nameStart+1 : nameEnd]

	// grab last line
	lastLine := lines[len(lines)-1]

	// trim the cruft of the front of the line
	lastLine = strings.TrimLeft(lastLine, ") ")

	if hasParameter(lastLine, "ENGINE") {
		// extract ENGINE and value
		engine, err = extractParameter(lastLine, "ENGINE")

		if util.ErrorCheckf(err, "Error Parsing ENGINE") {
			return parseError("Malformed Table ENGINE definition")
		}
	}

	if hasParameter(lastLine, "AUTO_INCREMENT") {
		// extract AUTO_INCREMENT and value
		var aip string
		aip, err = extractParameter(lastLine, "AUTO_INCREMENT")
		autoinc, err = strconv.ParseInt(aip, 10, 64)
		if util.ErrorCheckf(err, "Error Parsing AUTO_INCREMENT") {
			return parseError("Malformed AUTO_INCREMENT definition")
		}
	}

	// extract DEFAULT CHARSET and value
	if hasParameter(lastLine, "DEFAULT CHARSET") {
		// extract DEFAULT CHARSET and value
		charset, err = extractParameter(lastLine, "DEFAULT CHARSET")
		if util.ErrorCheckf(err, "Error Parsing DEFAULT CHARSET") {
			return parseError("Malformed DEFAULT CHARSET definition")
		}
	}

	// Get Metadata for the table
	hasMetadata, err = metadata.TableRegistered(name)

	if util.ErrorCheckf(err, "Error accessing Metadata database table") {
		return parseError("Error accessing Metadata database table")
	}

	if hasMetadata {
		md, err = metadata.GetTableByName(name)
		if !util.ErrorCheckf(err, "Problem finding metadata for table: "+name) {
			tbl.Metadata = md
		} else {
			return parseError("Problem finding table in Metadata database")
		}
	} else {
		// New Table so fill out the Metadata
		md.Name = name
		md.Type = "Table"
		md.Exists = true
		tbl.Metadata = md
	}

	tbl.Name = name
	tbl.Engine = engine
	tbl.AutoInc = autoinc
	tbl.CharSet = charset
	tbl.Filename = "DB"

	return err
}

func buildColumn(line string, tblPropertyID string, tblName string) (column table.Column, err error) {

	var name string
	var hasMetadata bool
	var md metadata.Metadata

	// extract NOT NULL

	// trim whitespace from string after last )
	parameters := strings.TrimSpace(line[strings.LastIndex(line, ")"):])

	// NOT NULL by default
	nullable := false
	autoinc := false

	// If NOT NULL is not present
	if strings.Index(parameters, "NOT NULL") == -1 {
		nullable = true
	}

	// If AUTO_INCREMENT is not present
	if strings.Index(parameters, "AUTO_INCREMENT") == -1 {
		autoinc = false
	}

	// Split components
	bracketClose := strings.LastIndex(line, ")")
	if bracketClose == -1 {
		return column, parseError(fmt.Sprintf("Invalid Column Definition: Missing size: [%s]", line))
	}
	line = line[:bracketClose]

	// split on whitespace
	lineSplit := strings.Split(strings.TrimSpace(line), " ")

	if len(lineSplit) < 2 {
		return column, parseError(fmt.Sprintf("Invalid Column Definition: Invalid number of properties: [%s]", line))
	}

	// Parse Name
	// extract item[0] = name using ``
	name = strings.Trim(lineSplit[0], "`")

	// Parse Datatype and Size
	//

	// split on (
	if strings.Index(lineSplit[1], "(") == -1 {
		return column, parseError(fmt.Sprintf("Invalid Column Definition: Missing size: [%s]", line))
	}
	dt := strings.Split(lineSplit[1], "(")

	// use dt[0] as datatype
	datatype := dt[0]

	// dt[1][:-1] as size
	var colSize int
	sizeStr := strings.Trim(dt[1], ")")
	colSize, err = strconv.Atoi(sizeStr)
	if util.ErrorCheckf(err, "Error Parsing Column Size parameter") {
		return column, parseError(fmt.Sprintf("Invalid Column Definition: Datatype size: Parse failed: [%s]", line))
	}

	// Build Column result
	column.Name = name
	column.Type = datatype
	column.Size = colSize
	column.Nullable = nullable
	column.AutoInc = autoinc

	if hasMetadata {
		// Retrieve Metadata for column
		md, err = metadata.GetByName(name, tblPropertyID)
		if !util.ErrorCheckf(err, "Problem finding metadata for Column: [%s] in Table: [%s]", name, tblName) {
			column.Metadata = md
		} else {
			return column, parseError(fmt.Sprintf("Failed to retrieve Column Metadata: [%s]", line))
		}
	} else {
		md.Name = column.Name
		md.Type = "Column"
		md.Exists = true
		column.Metadata = md
	}

	return column, err
}

func buildPrimaryKey(pk string, tblPropertyID string, tblName string) (primaryKey table.Index, err error) {

	var hasMetadata bool
	var md metadata.Metadata

	// Format: PRIMARY KEY (`<COLUMN_1>`, `<COLUMN_2>`)
	// extract substring between brackets

	firstBracket := strings.Index(pk, "(") + 1
	secondBracket := strings.Index(pk, ")")

	if firstBracket == -1 || secondBracket == -1 || firstBracket == secondBracket {
		return primaryKey, parseError(fmt.Sprintf("Malformed PrimaryKey Columns definition. [%s]", pk))
	}
	pk = pk[firstBracket:secondBracket]

	// split on ,
	columns := strings.Split(pk, ",")

	if len(columns) < 1 {
		return primaryKey, parseError(fmt.Sprintf("No Columns found for PrimaryKey. [%s]", pk))
	}

	primaryKey.IsPrimary = true
	primaryKey.Name = table.PrimaryKey

	if hasMetadata {
		// Retrieve Metadata for Primary Key
		md, err = metadata.GetByName(table.PrimaryKey, tblPropertyID)
		if !util.ErrorCheckf(err, "Problem finding metadata for Primary Key in Table: [%s]", tblName) {
			primaryKey.Metadata = md
		} else {
			return primaryKey, parseError(fmt.Sprintf("Failed to retrieve PrimaryKey Metadata: [%s]", pk))
		}
	} else {
		md.Name = table.PrimaryKey
		md.Type = table.PrimaryKey
		md.Exists = true
		primaryKey.Metadata = md
	}

	for _, column := range columns {
		// strip ` and add to primary key array
		primaryKey.Columns = append(primaryKey.Columns, strings.Trim(column, "`"))
	}

	return primaryKey, err

}

func buildIndex(key string, tblPropertyID string, tblName string) (index table.Index, err error) {
	// Format: KEY `<NAME>` (`<COLUMN_1>`,`<COLUMN_2>`)

	var hasMetadata bool
	var md metadata.Metadata

	// Remove KEY Prefix
	key = strings.TrimLeft(key, "KEY ")

	// Separate name from columns
	nv := strings.Split(key, " ")
	if len(nv) < 2 {
		return index, parseError(fmt.Sprintf("Invalid Index Definition: Invalid number of properties: [%s]", key))
	}
	index.Name = strings.Trim(nv[0], "`")

	// Process Index Columns
	cnames := strings.Split(strings.Trim(nv[1], "()"), ",")

	if len(cnames) == 0 {
		return index, parseError(fmt.Sprintf("Invalid Index Definition: No columns defined for index: [%s]", key))
	}

	for _, column := range cnames {
		index.Columns = append(index.Columns, strings.Trim(column, "`"))
	}

	if hasMetadata {
		// Retrieve Metadata for index
		md, err = metadata.GetByName(index.Name, tblPropertyID)
		if !util.ErrorCheckf(err, "Problem finding metadata for Index: [%s] in Table: [%s]", index.Name, tblName) {
			index.Metadata = md
		} else {
			return index, parseError(fmt.Sprintf("Failed to retrieve Index Metadata: [%s]", key))
		}
	} else {
		md.Name = index.Name
		md.Type = "Index"
		md.Exists = true
		index.Metadata = md
	}

	return index, err
}

// ParseCreateTable Parses a MySQL Create Table statement into a table.Table struct
func ParseCreateTable(createTable string) (tbl table.Table, err error) {

	if len(createTable) == 0 {
		return tbl, parseError("Empty CREATE TABLE statement")
	}

	// Split by newlines
	lines := strings.Split(createTable, "\n")

	// Strip any trailing commas
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimRight(lines[i], ",")
	}

	err = buildTable(lines, &tbl)

	// This code will make some assumptions regarding the create table
	// It will most likely need cleanup at some point
	var pk string
	var column []string
	var secondaryKeys []string

	// Process the lines into the appropriate categories
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PRIMARY KEY") {
			pk = line

		} else if strings.HasPrefix(line, "KEY") {
			secondaryKeys = append(secondaryKeys, line)

		} else if strings.HasPrefix(line, "`") {
			column = append(column, line)
		}
	}

	// process table column and keys
	var col table.Column
	for _, line := range column {
		col, err = buildColumn(line, tbl.Metadata.PropertyID, tbl.Name)
		if !util.ErrorCheckf(err, "Failed to parse column from CREATE TABLE") {
			tbl.Columns = append(tbl.Columns, col)
		} else {
			return tbl, err
		}
	}

	// If the table has a Primary Key
	var primaryKey table.Index
	if len(pk) > 0 {
		primaryKey, err = buildPrimaryKey(pk, tbl.Metadata.PropertyID, tbl.Name)
		if !util.ErrorCheckf(err, "Failed to parse PrimaryKey from CREATE TABLE") {
			tbl.PrimaryIndex = primaryKey
		} else {
			return tbl, err
		}
	}

	// extract any KEY values
	var index table.Index
	for _, key := range secondaryKeys {
		index, err = buildIndex(key, tbl.Metadata.PropertyID, tbl.Name)
		if !util.ErrorCheckf(err, "Failed to parse PrimaryKey from CREATE TABLE") {
			tbl.SecondaryIndexes = append(tbl.SecondaryIndexes, index)
		} else {
			return tbl, err
		}

	}

	return tbl, err
}

// ReadTables Reads the database for the project parameter and parses the
// show create table result for each into table.Table structs
func ReadTables(project config.Project) (err error) {
	con, err := sql.Open("mysql", project.DB.ConnectString())

	util.ErrorCheckf(err, "Problem opening connection to target database")

	if con != nil {
		util.LogInfo("DB Connection Success!")

		var rows *sql.Rows
		rows, err = con.Query("show tables")

		util.ErrorCheckf(err, "Problem retrieving tables")

		defer rows.Close()

		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			util.ErrorCheckf(err, "Could not parse name from database tables")

			// Extract the Create Tables
			var row *sql.Rows
			row, err = con.Query("show create table " + name)
			util.ErrorCheckf(err, "Could not execute show create tables for: "+name)

			defer row.Close()

			for row.Next() {
				var create string
				var tbl table.Table

				err = row.Scan(&name, &create)
				util.ErrorCheck(err)

				tbl, err = ParseCreateTable(create)
				if !util.ErrorCheck(err) {
					Schema = append(Schema, tbl)
				}
			}
		}
	}

	util.LogInfo("DB Processing Finished")

	return err
}

// func WriteSQLFile(file string, table Table) (err error) {
//
// 	// writeOp := GenerateCreateTable(table)
// 	//
// 	// createTable := writeOp.Slice[0].Statement
// 	//
// 	// util.LogInfo("Write SQL file: " + file)
// 	//
// 	// err = ioutil.WriteFile(file, []byte(createTable), 0644)
// 	//
// 	// util.ErrorCheck(err)
//
// 	return err
// }
//
// func WriteDBTables(path string, tables []Table) (err error) {
//
// 	rootPath := path
// 	for _, table := range tables {
// 		filepath := filepath.Join(rootPath, "test.sql")
// 		err = WriteSQLFile(filepath, table)
//
// 		util.ErrorCheck(err)
// 	}
//
// 	return err
// }
