package table

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/util"
)

type Tables []Table

type Column struct {
	PropertyID string `yaml:"id"`
	Name       string
	Type       string
	Size       int
	Nullable   bool
	AutoInc    bool

	// Binary      bool
	// Unique      bool
	// Unsigned    bool
	// ZeroFilled  bool
}

func (c Column) ToSQL() string {

	var params util.Params

	if !c.Nullable {
		params.Add("NOT NULL")
	}

	if c.AutoInc {
		params.Add("AUTO_INCREMENT")
	}
	return fmt.Sprintf("%s %s(%d) %s", c.Name, c.Type, c.Size, params.String())
}

type Index struct {
	PropertyID string `yaml:"id"`
	Name       string
	Columns    []string
	IsPrimary  bool
	IsUnique   bool
}

func (i Index) ToSQL() string {

	name := ""
	columns := ""

	if i.IsPrimary {
		name = "PRIMARY KEY"
	} else {
		isUnique := ""
		if i.IsUnique {
			isUnique = "UNIQUE"
		}
		name = fmt.Sprintf("%s KEY `%s` ", isUnique, i.Name)
	}

	columns = strings.Join(i.Columns, ", ")

	return fmt.Sprintf("%s (%s)", name, columns)
}

type Table struct {
	PropertyID       string `yaml:"id"`
	Name             string
	Engine           string
	AutoInc          int64
	CharSet          string
	Columns          []Column
	PrimaryIndex     Index
	SecondaryIndexes []Index

	namespace []string
	file      string
}

func (t *Table) SetNamespace(path string, filename string) (err error) {
	wd, err := os.Getwd()

	t.file = filepath.Join(path, filename)

	relativePath, err := filepath.Rel(filepath.Join(wd, path), filename)

	dir, _ := filepath.Split(relativePath)

	var ns []string

	if len(dir) > 0 {
		// TODO: Cross platform support
		ns = strings.Split(dir, "/")
		t.namespace = ns[:len(ns)-1]

		// rewrite tablenames
		t.Name = fmt.Sprintf("%s_%s", strings.Join(t.namespace, "_"), t.Name)
	}

	return err
}

func (t *Table) RemoveNamespace() {
	ns := strings.Split(t.Name, "_")
	t.Name = ns[len(ns)-1]
}
