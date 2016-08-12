package yaml

import (
	"path/filepath"
	"strings"

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
	tbl.RemoveNamespace()
	filename := strings.ToLower(tbl.Name) + ".yml"
	filepath := filepath.Join(path, filename)

	util.LogInfof("Writing Table YAML to file: %s", filepath)
	err = WriteFile(filepath, tbl)

	return err
}
