package table

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/metadata"
)

const (
	PrimaryKey = "PrimaryKey"
)

// IndexColumn Stores the properties of an index field
type IndexColumn struct {
	Name   string
	Length int `yaml:",omitempty"`
}

// ToSQL Generata a SQL representation
func (i IndexColumn) ToSQL() string {
	if i.Length > 0 {
		return fmt.Sprintf("`%s`(%d)", i.Name, i.Length)
	}
	return fmt.Sprintf("`%s`", i.Name)
}

// Index Stores the properties for a Index field
type Index struct {
	ID        string `yaml:"id"`
	Name      string
	Columns   []IndexColumn
	IsPrimary bool              `yaml:",omitempty"`
	IsUnique  bool              `yaml:",omitempty"`
	Metadata  metadata.Metadata `yaml:"-"`
}

// IsValid Return if the index contains any columns
func (i Index) IsValid() bool {
	return len(i.Columns) > 0
}

// ToSQL Formats the index into its SQL representation
func (i Index) ToSQL() string {

	if len(i.Columns) == 0 {
		return ""
	}
	name := ""

	if i.IsPrimary {
		name = "PRIMARY KEY"
	} else {
		isUnique := ""
		if i.IsUnique {
			isUnique = "UNIQUE"
		}
		name = fmt.Sprintf("%s KEY `%s`", isUnique, i.Name)
	}

	return fmt.Sprintf("%s %s", name, i.ColumnsSQL())
}

// ColumnsSQL Formats the Index columns into the appropriate SQL representation
func (i Index) ColumnsSQL() string {
	columnStr := []string{}

	for _, indCol := range i.Columns {
		columnStr = append(columnStr, indCol.ToSQL())
	}

	return fmt.Sprintf("(%s)", strings.Join(columnStr, ","))
}
