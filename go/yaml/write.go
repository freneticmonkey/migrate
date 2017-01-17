package yaml

import (
	"path/filepath"

	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

// WriteTables Write the tables parameter to the path as YAML files
func WriteTables(path string, tables table.Tables) (err error) {
	for _, tbl := range tables {
		WriteTable(path, tbl)
	}

	return err
}

// WriteTable Serialise the Table has YAML and write it to path
func WriteTable(path string, tbl table.Table) (err error) {
	tbl.Namespace.SetTableFilename("<table>")
	filename := tbl.Namespace.GenerateSchemaFilename("yml")
	filepath := filepath.Join(path, filename)
	util.LogInfof("Writing to File PATH: %s", filepath)
	err = WriteFile(filepath, tbl)

	return err
}
