package yaml

import (
	"path/filepath"

	"github.com/freneticmonkey/migrate/migrate/table"
)

func WriteTables(path string, tables table.Tables) (err error) {
	for _, tbl := range tables {
		filepath := filepath.Join(path, "test.yml")
		WriteFile(filepath, tbl)
	}

	return err
}
