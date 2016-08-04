package mysql

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	// Get a MySQL Database Connection
	"database/sql"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	// This is apparently how this is included
	_ "github.com/go-sql-driver/mysql"
)

// MySQL Keywords
const (
	NULL            = "NULL"
	DEFAULT         = "DEFAULT"
	NOT_NULL        = "NOT NULL"
	UNSIGNED        = "UNSIGNED"
	AUTO_INCREMENT  = "AUTO_INCREMENT"
	ENGINE          = "ENGINE"
	DEFAULT_CHARSET = "DEFAULT CHARSET"
	ROW_FORMAT      = "ROW_FORMAT"
	COLLATE         = "COLLATE"
	DEFAULT_COLLATE = "DEFAULT COLLATE"
)

var Schema table.Tables
var alters []string

var datatypes = []string{
	"char",
	"varchar",
	"tinytext",
	"text",
	"mediumtext",
	"longtext",
	"tinyblob",
	"blob",
	"mediumblob",
	"longblob",
	"tinyint",
	"smallint",
	"mediumint",
	"int",
	"bigint",
	"real",
	"float",
	"double",
	"decimal",
	"numeric",
	"bit",
	"date",
	"datetime",
	"timestamp",
	"time",
	"enum",
	"set",
	"json",
}

var datatypesSizable = []string{
	"char",
	"varchar",
	"tinyint",
	"smallint",
	"mediumint",
	"int",
	"bigint",
	"real",
	"float",
	"double",
	"decimal",
	"numeric",
	"bit",
}

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
	var rowFormat string
	var collation string

	// var hasMetadata bool
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

	if hasParameter(lastLine, ENGINE) {
		// extract ENGINE and value
		engine, err = extractParameter(lastLine, ENGINE)

		if util.ErrorCheckf(err, "Error Parsing ENGINE") {
			return parseError("Malformed Table ENGINE definition")
		}
	}

	if hasParameter(lastLine, AUTO_INCREMENT) {
		// extract AUTO_INCREMENT and value
		var aip string
		aip, err = extractParameter(lastLine, AUTO_INCREMENT)
		autoinc, err = strconv.ParseInt(aip, 10, 64)
		if util.ErrorCheckf(err, "Error Parsing AUTO_INCREMENT") {
			return parseError("Malformed AUTO_INCREMENT definition")
		}
	}

	// extract DEFAULT CHARSET and value
	if hasParameter(lastLine, DEFAULT_CHARSET) {
		// extract DEFAULT CHARSET and value
		charset, err = extractParameter(lastLine, DEFAULT_CHARSET)
		if util.ErrorCheckf(err, "Error Parsing DEFAULT CHARSET") {
			return parseError("Malformed DEFAULT CHARSET definition")
		}
	}

	// extract ROW_FORMAT and value
	if hasParameter(lastLine, ROW_FORMAT) {
		// extract ROW_FORMAT and value
		rowFormat, err = extractParameter(lastLine, ROW_FORMAT)
		if util.ErrorCheckf(err, "Error Parsing ROW_FORMAT") {
			return parseError("Malformed ROW_FORMAT definition")
		}
	}

	// extract COLLATE and value
	if hasParameter(lastLine, COLLATE) {
		// extract COLLATE and value
		collation, err = extractParameter(lastLine, COLLATE)
		if util.ErrorCheckf(err, "Error Parsing COLLATE") {
			return parseError("Malformed COLLATE definition")
		}
	}

	// Fill out the Metadata details.
	md.Name = name
	md.Type = "Table"
	md.Exists = true
	tbl.Metadata = md

	tbl.Name = name
	tbl.Engine = engine
	tbl.AutoInc = autoinc
	tbl.CharSet = charset
	tbl.RowFormat = rowFormat
	tbl.Collation = collation
	tbl.Filename = "DB"

	return err
}

func buildColumn(line string, tblPropertyID string, tblName string) (column table.Column, err error) {

	var name string
	var md metadata.Metadata

	// Trim whitespace from the ends of the statement
	line = strings.TrimSpace(line)

	// Split on whitespace.
	// This will result in:
	// [0] Name
	// [1] Size definition
	// [2] 1st clause
	// ...
	// [x] X clause
	components := strings.Split(line, " ")

	if len(components) < 2 {
		return column, parseError(fmt.Sprintf("Invalid Column Definition: Unparsable due to malformed or missing name or datatype: [%s]", line))
	}

	// Parse Name
	// extract item[0] = name using ``
	name = strings.Trim(components[0], "`")

	// Parse Datatype and Size
	dts := components[1]

	var datatype string
	var sizeStr string
	var colSizes []int

	// Check if the datatype supports an optional size

	// Detect size brackets
	if strings.Index(dts, "(") == -1 {
		// If not found check if the datatype matches a known type
		if !util.StringInArray(dts, datatypes) {
			// it's not a valid datatype
			return column, parseError(fmt.Sprintf("Invalid Column Definition: Unsupported datatype: [%s]", line))
		}
		datatype = dts

	} else {
		dtComp := strings.Split(dts, "(")

		if len(dtComp) != 2 {
			return column, parseError(fmt.Sprintf("Invalid Column Definition: Unparsable datatype defintion: [%s]", line))
		}
		// Extract datatype
		datatype = dtComp[0]

		if len(datatype) == 0 {
			return column, parseError(fmt.Sprintf("Invalid Column Definition: Malformed column name: [%s]", line))
		}

		// Extract size
		sizeStr = dtComp[1]
		// Clean trailing bracket
		sizeStr = strings.TrimRight(sizeStr, ")")

		// Check that the datatype can have a size set
		if !util.StringInArray(datatype, datatypesSizable) {
			// it's not a valid sizable datatype
			return column, parseError(fmt.Sprintf("Invalid Column Definition: Size defined for unsizable datatype: [%s]", line))
		}

		// Parse column size value(s)
		var colSize int

		// If the size type contains a length and decimal length
		if strings.Index(sizeStr, ",") != -1 {
			sizeSlice := strings.Split(sizeStr, ",")
			for _, size := range sizeSlice {
				colSize, err = strconv.Atoi(size)
				if err != nil {
					return column, parseError(fmt.Sprintf("Invalid Column Definition: Malformed length or decimal lengthk for datatype size: [%s]", line))
				}
				colSizes = append(colSizes, colSize)
			}
		} else {
			colSize, err = strconv.Atoi(sizeStr)
			if err != nil {
				return column, parseError(fmt.Sprintf("Invalid Column Definition: Datatype size: Parse failed: [%s]", line))
			}

			colSizes = []int{colSize}
		}
	}

	// Calculate the column clauses / parameters offset.

	// offset = name + space + datatype(size)
	paramOffset := len(components[0]) + 1 + len(components[1])

	parameters := line[paramOffset:]

	// Convert the parameters to Upper
	parameters = strings.ToUpper(parameters)

	// NULL by default
	nullable := true
	unsigned := false
	autoinc := false
	defaultValue := ""
	collationValue := ""

	// If unsigned is present
	if strings.Index(parameters, UNSIGNED) != -1 {
		unsigned = true
	}

	// If NOT NULL is present
	if strings.Index(parameters, NOT_NULL) != -1 {
		nullable = false
	}

	// If AUTO_INCREMENT is present
	if strings.Index(parameters, AUTO_INCREMENT) != -1 {
		autoinc = true
	}

	// if DEFAULT is present
	defaultPos := strings.Index(parameters, DEFAULT)
	if defaultPos != -1 {
		// Grab the string after DEFAULT, trim it, and split on whitespace.

		// Check that the line contains a DEFAULT value
		if len(parameters) < defaultPos+len(DEFAULT)+1 {
			return column, parseError(fmt.Sprintf("Invalid Column Definition: Default value missing: [%s]", line))
		}

		// Now extract value of the parameter from the original line (non-ToUpper())
		lineEnd := line[paramOffset+defaultPos+len(DEFAULT):]
		defaultStr := strings.TrimSpace(lineEnd)

		// if single quotes are detected
		quotePos := strings.Index(defaultStr, "'")
		if quotePos != -1 {
			// extract the contents of the single quotes
			qEnd := strings.LastIndex(defaultStr, "'")

			if quotePos != qEnd {
				defaultValue = defaultStr[quotePos+1 : qEnd]
			} else {
				return column, parseError(fmt.Sprintf("Invalid Column Definition: DEFAULT value is empty: [%s]", line))
			}
		} else {

			// Check for NULL default value if there aren't any quotes.  Can only be NULL
			if len(defaultStr) >= 4 {
				if defaultStr != NULL {
					dCmp := strings.Split(defaultStr, " ")
					if len(dCmp) > 0 && dCmp[0] == NULL {
						defaultValue = NULL
					}
				} else {
					defaultValue = defaultStr
				}
			}
			// If the default value is shorter than NULL, then there's a problem

			// If we're unable to parse the default value
			if defaultValue == "" {
				return column, parseError(fmt.Sprintf("Invalid Column Definition: Unable to parse DEFAULT value: [%s]", line))
			}
		}
	}

	// if COLLATE is present
	collatePos := strings.Index(parameters, COLLATE)
	if collatePos != -1 {
		// Grab the string after COLLATE, trim it, and split on whitespace.
		// Check that the line contains a COLLATION value
		if len(parameters) < collatePos+len(COLLATE)+1 {
			return column, parseError(fmt.Sprintf("Invalid Column Definition: COLLATE type missing: [%s]", line))
		}

		// Now extract value of the parameter from the original line (non-ToUpper())
		lineEnd := line[paramOffset+collatePos+len(COLLATE):]
		collateStr := strings.TrimSpace(lineEnd)

		cCmp := strings.Split(collateStr, " ")
		if len(cCmp) > 0 && cCmp[0] != "" {
			collationValue = cCmp[0]
		} else {
			return column, parseError(fmt.Sprintf("Invalid Column Definition: Couldn't extract COLLATE type: [%s]", line))
		}
	}

	// Build Column result
	column.Name = name
	column.Type = datatype
	column.Size = colSizes
	column.Unsigned = unsigned
	column.Default = defaultValue
	column.Nullable = nullable
	column.AutoInc = autoinc
	column.Collation = collationValue

	md.Name = column.Name
	md.Type = "Column"
	md.Exists = true

	column.Metadata = md

	return column, err
}

func buildIndexColumns(key string) (indexColumns []table.IndexColumn, err error) {
	// Find Column brackets
	lb := strings.Index(key, "(")
	rb := strings.LastIndex(key, ")")

	columnsStr := key[lb+1 : rb]

	if len(columnsStr) == 0 {
		return indexColumns, parseError(fmt.Sprintf("Invalid Index Definition: No columns defined for index: [%s]", key))
	}
	columns := strings.Split(columnsStr, ",")

	for _, colStr := range columns {
		ic := table.IndexColumn{}
		// Check for length definition
		if strings.ContainsAny(colStr, "()") {
			clb := strings.Index(colStr, "(")
			crb := strings.LastIndex(colStr, ")")
			ic.Length, err = strconv.Atoi(colStr[clb+1 : crb])
			if err != nil {
				return indexColumns, parseError(fmt.Sprintf("Invalid Index Definition: Invalid partial index value: [%s]", key))
			}
			colStr = colStr[:clb]
		}
		// Strip quotes from name
		ic.Name = strings.Trim(colStr, "`")
		indexColumns = append(indexColumns, ic)
	}

	return
}

func buildPrimaryKey(pk string, tblPropertyID string, tblName string) (primaryKey table.Index, err error) {

	var md metadata.Metadata

	// Format: PRIMARY KEY (`<COLUMN_1>`[(<size)], `<COLUMN_2>`[(<size)])
	// extract substring between brackets

	// remove whitespace
	pk = strings.TrimSpace(pk)

	if !strings.HasPrefix(pk, "PRIMARY KEY") {
		return primaryKey, parseError(fmt.Sprintf("Invalid Primary Key Definition: Invalid PRIMARY KEY type: [%s]", pk))
	}

	pk = strings.TrimPrefix(pk, "PRIMARY KEY")

	primaryKey.Columns, err = buildIndexColumns(pk)
	primaryKey.IsPrimary = true
	primaryKey.Name = table.PrimaryKey

	md.Name = primaryKey.Name
	md.Type = "PrimaryKey"
	md.Exists = true
	primaryKey.Metadata = md

	return primaryKey, err

}

func buildIndex(key string, tblPropertyID string, tblName string) (index table.Index, err error) {
	// Format: [UNIQUE] KEY `<NAME>` (`<COLUMN_1>`[(<size)],`<COLUMN_2>`[(<size)])

	var md metadata.Metadata

	if !strings.HasPrefix(key, "KEY") && !strings.HasPrefix(key, "UNIQUE KEY") {
		return index, parseError(fmt.Sprintf("Invalid Index Definition: Invalid KEY type: [%s]", key))
	}

	if strings.HasPrefix(key, "UNIQUE KEY") {
		index.IsUnique = true
		// Remove UNIQUE Prefix
		key = strings.TrimLeft(key, "UNIQUE ")
	}

	// Remove KEY Prefix
	key = strings.TrimLeft(key, "KEY ")

	// Extract Name, stripping whitespace and backticks
	index.Name = strings.Trim(key[:strings.Index(key, "(")], " `")

	if len(index.Name) == 0 {
		return index, parseError(fmt.Sprintf("Invalid Index Definition: No name defined: [%s]", key))
	}

	// Extract Columns
	index.Columns, err = buildIndexColumns(key)

	md.Name = index.Name
	md.Type = "Index"
	md.Exists = true
	index.Metadata = md

	return index, err
}

// retrieveTableMetadata Retrieve the metadata for the Table and all of its properties
func retrieveTableMetadata(tbl *table.Table) (err error) {
	var mds []metadata.Metadata

	mds, err = metadata.LoadAllTableMetadata(tbl.Name)

	for _, md := range mds {
		// Table
		if md.ParentID == "" && md.Type == "Table" {
			tbl.Metadata = md
		}

		// Columns
		if md.Type == "Column" {
			for i := 0; i < len(tbl.Columns); i++ {
				if md.Name == tbl.Columns[i].Name {
					tbl.Columns[i].Metadata = md
				}
			}
		}

		// Primary Key
		if md.Type == "PrimaryKey" {
			tbl.PrimaryIndex.Metadata = md
		}

		// Indexes
		if md.Type == "Index" {
			for i := 0; i < len(tbl.SecondaryIndexes); i++ {
				if md.Name == tbl.SecondaryIndexes[i].Name {
					tbl.SecondaryIndexes[i].Metadata = md
				}
			}
		}
	}

	return err
}

// ParseCreateTable Parses a MySQL Create Table statement into a table.Table struct
func ParseCreateTable(createTable string) (tbl table.Table, err error) {

	if len(createTable) == 0 {
		return tbl, parseError("Empty CREATE TABLE statement")
	}

	// Split by newlines
	lines := strings.Split(createTable, "\n")

	// Strip any trailing commas or semi-colons
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimRight(lines[i], ",")
		lines[i] = strings.TrimRight(lines[i], ";")
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

		} else if strings.HasPrefix(line, "KEY") || strings.HasPrefix(line, "UNIQUE KEY") {
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

	// Retrieve any Metadata from the Management DB
	retrieveTableMetadata(&tbl)

	return tbl, err
}

// ReadTables Reads the database for the project parameter and parses the
// show create table result for each into table.Table structs
func ReadTables() (err error) {

	type CreateTable struct {
		Name            string
		CreateStatement string
	}

	var rows *sql.Rows
	var pdb *sql.DB
	var tables []CreateTable
	var tbl table.Table

	// Connect to the Project database
	pdb, err = connectProjectDB()

	// Ensure that the connection is cleaned up
	defer pdb.Close()
	if util.ErrorCheckf(err, "Problem opening connection to target database") {
		return err
	}

	// If the Database connection exists
	if pdb != nil {
		rows, err = pdb.Query("show tables")

		if util.ErrorCheckf(err, "Problem retrieving tables") {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if util.ErrorCheckf(err, "Could not parse name from database tables") {
				return err
			}
			tables = append(tables, CreateTable{name, ""})
		}

		// Extract the Create Table Statements
		for i, ct := range tables {
			rows, err = pdb.Query("show create table " + ct.Name)
			if util.ErrorCheckf(err, "Could not execute show create tables for: "+ct.Name) {
				return err
			}

			for rows.Next() {
				var name string
				var create string

				err = rows.Scan(&name, &create)
				if util.ErrorCheck(err) {
					return err
				}
				tables[i].CreateStatement = create

			}
		}

		// Process the Create Table Statements into Tables
		for _, ct := range tables {
			tbl, err = ParseCreateTable(ct.CreateStatement)
			if util.ErrorCheck(err) {
				return err
			}
			Schema = append(Schema, tbl)
		}
	}

	return err
}

// ReadDump Read a MySQL Dump file as a source of MySQL Schema and return the
// CREATE TABLE statements as an array of strings
func ReadDump(filename string) (statements []string, err error) {

	var dump []byte
	// Read the Dump file.
	dump, err = ioutil.ReadFile(filename)

	lines := strings.Split(string(dump), "\n")

	// Extract CREATE TABLE statements
	type DumpTable struct {
		lines []string
	}
	dumpTables := []*DumpTable{}

	var currentTable *DumpTable

	for _, line := range lines {
		// Ignore comments
		if len(line) == 0 || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "DROP") {
			continue
		}

		if strings.HasPrefix(line, "CREATE TABLE") {
			if currentTable != nil {
				dumpTables = append(dumpTables, currentTable)
			}
			currentTable = &DumpTable{}
		}

		if currentTable != nil {
			currentTable.lines = append(currentTable.lines, line)
		}
	}

	if currentTable != nil {
		dumpTables = append(dumpTables, currentTable)
	}

	for _, dt := range dumpTables {

		// Build CREATE TABLE string
		statements = append(statements, strings.Join(dt.lines, "\n"))
	}

	return statements, err

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
