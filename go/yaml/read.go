package yaml

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

// ReadTables Read all of the files at path that have the extension 'yml' and parse them
// into table.Table structs
func ReadTables(conf config.Config) (err error) {
	path := strings.ToLower(conf.Project.Name)
	var schemaList []string

	// If the path has been defined as ignore, then immediately return without error.
	// This is intended to be used for unit tests which will manually add Tables to the
	// YAML Schema.
	if path == "ignore" {
		return err
	}

	path = strings.ToLower(path)

	// Recursively build a list of YAML schema files
	err = util.ReadDirRelative(path, "yml", &schemaList)

	if !util.ErrorCheckf(err, "Error reading YAML files in path: [%s]", path) {
		for _, filename := range schemaList {

			var tbl table.Table

			err = ReadFile(filename, &tbl)

			if err != nil {
				return err
			}

			// Process the table metadata
			processMetadata(&tbl)

			// Calculate the table's namespace
			tbl.SetNamespace(conf)

			// If the table has an Id, then it can be used.
			// Otherwise ignore it.
			if len(tbl.Metadata.PropertyID) > 0 {
				// Set the Primary Index to true if it exists
				// FIXME: This is a clunky way of doing this.
				if len(tbl.PrimaryIndex.Columns) > 0 {
					tbl.PrimaryIndex.IsPrimary = true
				}

				Schema = append(Schema, tbl)
			} else {
				color.Set(color.FgYellow, color.Bold)
				util.LogWarn(fmt.Sprintf("Table in file: [%s] is missing a table id and is being ignored.", filename))
				color.Unset()
			}

		}
	}

	return err
}
