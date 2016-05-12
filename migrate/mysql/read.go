package mysql

import (
	"fmt"
	"strconv"
	"strings"

	// Get a MySQL Database Connection
	"database/sql"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/id"
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

func extractParameter(s string, param string) (result string) {

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
	}

	return result
}

func parseCreateTable(createTable string) (tbl table.Table, err error) {

	// Split by newlines
	lines := strings.Split(createTable, "\n")

	// Process Engine and Charset
	var name string
	var engine string
	var autoinc int64
	var charset string
	var id string

	// extract the name from the first line
	firstLine := lines[0]

	name = firstLine[strings.Index(firstLine, "`")+1 : strings.LastIndex(firstLine, "`")]

	// grab last line
	lastLine := lines[len(lines)-1]

	// trim the cruft of the front of the line
	lastLine = strings.TrimLeft(lastLine, ") ")

	if hasParameter(lastLine, "ENGINE") {
		// extract ENGINE and value
		engine = extractParameter(lastLine, "ENGINE")
	}

	if hasParameter(lastLine, "AUTO_INCREMENT") {
		// extract AUTO_INCREMENT and value
		autoinc, err = strconv.ParseInt(extractParameter(lastLine, "AUTO_INCREMENT"), 10, 64)
		util.ErrorCheckf(err, "Error Parsing AUTO_INCREMENT")
	}

	// extract DEFAULT CHARSET and value
	if hasParameter(lastLine, "DEFAULT CHARSET") {
		// extract DEFAULT CHARSET and value
		charset = extractParameter(lastLine, "DEFAULT CHARSET")
	}

	tbl.Name = name
	tbl.Engine = engine
	tbl.AutoInc = autoinc
	tbl.CharSet = charset
	tbl.Filename = "DB"

	// This code will make some assumptions regarding the create table
	// It will most likely need cleanup at some point
	var pk string
	var primaryKey table.Index
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
	for _, line := range column {
		// process each column entry

		line = strings.TrimRight(line, ",")

		// extract NOT NULL

		// trim whitespace from string after last )
		parameters := strings.TrimSpace(line[strings.LastIndex(line, ")"):])

		// NOT NULL by default
		nullable := false

		// If NOT NULL is not present
		if strings.Index(parameters, "NOT NULL") == -1 {
			nullable = true
		}

		// id, err := extractCommentId(parameters)
		// ErrorCheck(err)

		line = line[:strings.LastIndex(line, ")")]

		// split on whitespace
		lineSplit := strings.Split(strings.TrimSpace(line), " ")

		// extract item[0] = name using ``
		name := strings.Trim(lineSplit[0], "`")

		// split on (
		dt := strings.Split(lineSplit[1], "(")

		// use dt[0] as type
		datatype := dt[0]

		// dt[1][:-1] as size
		var colSize int
		colSize, err = strconv.Atoi(dt[1][:len(dt[1])])
		util.ErrorCheckf(err, "Error Parsing Column Size parameter")

		var column table.Column
		column.PropertyID = id
		column.Name = name
		column.Type = datatype
		column.Size = colSize
		column.Nullable = nullable

		tbl.Columns = append(tbl.Columns, column)
	}

	// Format: PRIMARY KEY (`<COLUMN_1>`, `<COLUMN_2>`) COMMENT='M_ID=<id>'
	// extract substring between brackets
	pk = pk[strings.Index(pk, "(")+1 : strings.Index(pk, ")")]
	// split on ,
	values := strings.Split(pk, ",")
	primaryKey.IsPrimary = true

	for _, column := range values {
		// strip ` and add to primary key array
		primaryKey.Columns = append(primaryKey.Columns, strings.Trim(column, "`"))
	}

	tbl.PrimaryIndex = primaryKey

	// extract any KEY values
	for _, key := range secondaryKeys {
		var index table.Index

		index.PropertyID = id

		// Format: KEY `<NAME>` (`<COLUMN_1>`,`<COLUMN_2>`)

		// Remove KEY Prefix
		key = strings.TrimLeft(key, "KEY ")

		// Separate name from columns
		nv := strings.Split(key, " ")
		index.Name = strings.Trim(nv[0], "`")

		// Process Index Columns
		cvalues := strings.Split(strings.Trim(nv[1], "()"), ",")
		for _, column := range cvalues {
			index.Columns = append(index.Columns, strings.Trim(column, "`"))
		}
		tbl.SecondaryIndexes = append(tbl.SecondaryIndexes, index)

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

				tbl, err = parseCreateTable(create)
				util.ErrorCheck(err)
				Schema = append(Schema, tbl)
			}
		}

		problems := id.ValidateSchema(Schema)
		if problems != 0 {
			err = fmt.Errorf("Reading tables from target database failed. %d problems found during validation", problems)
		}
		util.LogInfof("Successfully read %d tables from target database", len(Schema))
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
