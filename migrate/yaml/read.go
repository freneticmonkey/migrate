package yaml

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/freneticmonkey/migrate/migrate/id"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

var Schema table.Tables

var schemaList []string

func ReadTables(path string) (err error) {

	// Recursively build a list of YAML schema files
	err = util.ReadDirRelative(path, "yml", &schemaList)

	util.ErrorCheck(err)

	for _, filename := range schemaList {
		var tbl table.Table
		err = ReadFile(filename, &tbl)
		util.ErrorCheck(err)

		// Calculate the table's namespace
		tbl.SetNamespace(path, filename)

		// If the table has an Id, then it can be used.
		// Otherwise ignore it.
		if len(tbl.Id) > 0 {
			// Set the Primary Index to true if it exists
			// FIXME: This is a clunky way of doing this.
			if len(tbl.PrimaryIndex.Columns) > 0 {
				tbl.PrimaryIndex.IsPrimary = true
			}

			Schema = append(Schema, tbl)

			// Validate Table Ids
			if !id.ValidateSchema(Schema) {
				util.LogFatal("One or more problems found.  See log for details.")
			}

		} else {
			color.Set(color.FgYellow, color.Bold)
			util.LogWarn(fmt.Sprintf("Table in file: [%s] is missing a table id and is being ignored.", filename))
			color.Unset()
		}
	}

	return err
}
