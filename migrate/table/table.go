package table

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tables []Table

type Column struct {
	Id       string
	Name     string
	Type     string
	Size     int
	Nullable bool

	// Binary      bool
	// AutoInc     bool
	// Unique      bool
	// Unsigned    bool
	// ZeroFilled  bool
}

func (c Column) ToSQL() string {
	isNull := ""

	if !c.Nullable {
		isNull = "NOT NULL"
	}
	return fmt.Sprintf("%s %s(%d) %s COMMENT='%s'", c.Name, c.Type, c.Size, isNull, c.Id)
}

type Index struct {
	Id        string
	Name      string
	Columns   []string
	IsPrimary bool
	IsUnique  bool
}

func (i Index) ToSQL() string {
	name := ""
	columns := ""

	if i.IsPrimary {
		name = "PRIMARY KEY"
	} else {
		isUnique := ""
		if i.IsUnique {
			isUnique = "UNIQUE "
		}
		name = fmt.Sprintf("%sKEY `%s` ", isUnique, i.Name)
	}

	columns = strings.Join(i.Columns, ", ")

	return fmt.Sprintf("%s (%s) COMMENT='%s'", name, columns, i.Id)
}

type Table struct {
	hasId            bool
	Id               string
	Name             string
	Engine           string
	hasAutoInc       bool
	AutoInc          int64
	hasCharSet       bool
	CharSet          string
	Columns          []Column
	PrimaryIndex     Index
	SecondaryIndexes []Index

	namespace []string
}

func (t *Table) SetNamespace(path string, filename string) (err error) {
	wd, err := os.Getwd()

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
