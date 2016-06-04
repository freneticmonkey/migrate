package yaml

import (
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
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
	filename := strings.ToLower(tbl.Name) + ".yml"
	filepath := filepath.Join(path, filename)

	util.LogErrorf("Writing Table YAML to file: %s", filepath)
	WriteFile(filepath, tbl)

	return err
}
