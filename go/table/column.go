package table

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
)

// Column Stores the properties for a Column field
type Column struct {
	ID        string `yaml:"id"`
	Name      string
	Type      string
	Size      []int  `yaml:",flow"`
	Default   string `yaml:",omitempty"`
	Nullable  bool   `yaml:",omitempty"`
	AutoInc   bool   `yaml:",omitempty"`
	Unsigned  bool   `yaml:",omitempty"`
	Collation string `yaml:",omitempty"`

	// Binary      bool
	// Unique      bool
	// ZeroFilled  bool

	Metadata metadata.Metadata `yaml:"-"`
}

// ToSQL Formats the column into its SQL representation
func (c Column) ToSQL() string {

	var params util.Params

	if c.Unsigned {
		params.Add("unsigned")
	}

	if !c.Nullable {
		params.Add("NOT NULL")
	}

	if c.AutoInc {
		params.Add("AUTO_INCREMENT")
	}

	if len(c.Default) > 0 {
		value := c.Default
		// Throw quotes around it if the value is not NULL
		if value != "NULL" {
			value = fmt.Sprintf("'%s'", value)
		}
		params.Add(fmt.Sprintf("DEFAULT %s", value))
	}

	if len(c.Collation) > 0 {
		params.Add(fmt.Sprintf("COLLATE %s", c.Collation))
	}

	size := ""

	switch len(c.Size) {
	case 1:
		size = fmt.Sprintf("(%d)", c.Size[0])
	case 2:
		size = fmt.Sprintf("(%d,%d)", c.Size[0], c.Size[1])
	case 0:
		size = ""
	}
	sql := fmt.Sprintf("`%s` %s%s", c.Name, c.Type, size)
	if len(params.Values) > 0 {
		sql += fmt.Sprintf(" %s", params.String())
	}
	return sql
}
