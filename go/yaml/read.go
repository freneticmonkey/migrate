package yaml

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/fatih/color"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

// Schema The parsed from the YAML tables
var Schema table.Tables

var schemaList []string

// ReadYAML Parse the definition string parameter into a table.Table struct
func ReadYAML(definition string, context string) (tbl table.Table, err error) {
	err = yaml.Unmarshal([]byte(definition), &tbl)

	return tbl, err
}

// Postprocess the loaded YAML table for it's Metadata
func processMetadata(t *table.Table) {
	t.Metadata = metadata.Metadata{
		PropertyID: t.ID,
		Name:       t.Name,
		Type:       "Table",
	}

	for i, col := range t.Columns {
		t.Columns[i].Metadata = metadata.Metadata{
			PropertyID: col.ID,
			ParentID:   t.ID,
			Name:       col.Name,
			Type:       "Column",
		}
	}

	pk := t.PrimaryIndex
	t.PrimaryIndex.Metadata = metadata.Metadata{
		PropertyID: pk.ID,
		ParentID:   t.ID,
		Name:       "PrimaryKey",
		Type:       "Index",
	}

	for i, index := range t.SecondaryIndexes {
		t.SecondaryIndexes[i].Metadata = metadata.Metadata{
			PropertyID: index.ID,
			ParentID:   t.ID,
			Name:       index.Name,
			Type:       "Index",
		}
	}
}

// ReadTables Read all of the files at path that have the extension 'yml' and parse them
// into table.Table structs
func ReadTables(path string) (err error) {

	// If the path has been defined as ignore, then immediately return without error.
	// This is intended to be used for unit tests which will manually add Tables to the
	// YAML Schema.
	if path == "ignore" {
		return err
	}

	// Recursively build a list of YAML schema files
	err = util.ReadDirRelative(path, "yml", &schemaList)

	if !util.ErrorCheckf(err, "Error reading YAML files in path: [%s]", path) {
		for _, filename := range schemaList {

			var data []byte
			var tbl table.Table

			data, err = ioutil.ReadFile(filename)

			tbl, err = ReadYAML(string(data), filename)

			if !util.ErrorCheck(err) {
				// Process the table metadata
				processMetadata(&tbl)

				// Calculate the table's namespace
				tbl.SetNamespace(path, filename)

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
	}

	return err
}