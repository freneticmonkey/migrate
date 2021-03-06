package yaml

import (
	"os"
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

	// If the path has been defined as ignore, then immediately return without error.
	// This is intended to be used for unit tests which will manually add Tables to the
	// YAML Schema.
	if path == "ignore" {
		return err
	}

	path = strings.ToLower(path)

	// Read path under the project name
	err = readPath(path, false, conf)

	// Read any Schema namespaces
	for _, ns := range conf.Project.Schema.Namespaces {

		nsPath := ""

		if ns.SchemaPath != "" {
			nsPath = ns.SchemaPath//filepath.Join(path, ns.Path)
			err = readPath(nsPath, true, conf)

		} else {
			err = fmt.Errorf("SchemaPath value missing for Schema Namespace: %s", ns.Name)
		}

		if err != nil {
			if os.IsNotExist(err) {
				// Warn about missing schema
				util.LogWarnf("Schema Namespace path doesn't exist: %s", nsPath)
				// Disable the error
				err = nil
			} else {
				// Otherwise let it hit the fan
				util.ErrorCheckf(err, "Error reading YAML files in path: [%s]", path)
				return err
			}
		}
	}

	return err
}

func readPath(path string, recursive bool, conf config.Config) (err error) {
	var schemaList []string

	// Recursively build a list of YAML schema files
	err = util.ReadDirRelative(path, "yml", recursive, &schemaList)

	if err == nil {
		for _, filename := range schemaList {

			var tbl table.Table
			err = ReadFile(filename, &tbl)
			util.LogInfof("Reading YAML Table: %s", filename)
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
					tbl.PrimaryIndex.Name = "PrimaryKey"
					tbl.PrimaryIndex.ID = "primarykey"
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
