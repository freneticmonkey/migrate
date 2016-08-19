package table

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
)

// Tables Helper type for a slice of Table structs
type Tables []Table

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

const (
	PrimaryKey = "PrimaryKey"
)

// IndexColumn Stores the properties of an index field
type IndexColumn struct {
	Name   string
	Length int `yaml:",omitempty"`
}

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

// Table Stores the fields and properties representing a Table parsed from YAML
// or from a MySQL CREATE TABLE statement
type Table struct {
	ID               string `yaml:"id"`
	Name             string
	Engine           string
	AutoInc          int64 `yaml:",omitempty"`
	CharSet          string
	RowFormat        string `yaml:",omitempty"`
	Collation        string `yaml:",omitempty"`
	Columns          []Column
	PrimaryIndex     Index   `yaml:",omitempty"`
	SecondaryIndexes []Index `yaml:",omitempty"`

	namespace []string
	Filename  string            `yaml:"-"`
	Metadata  metadata.Metadata `yaml:"-"`
}

// SetNamespace Use the path and filename parameters to rename the table
// into an underscore delimited namespace
func (t *Table) SetNamespace(path string, filename string) (err error) {
	wd, err := os.Getwd()

	t.Filename = filename

	relativePath, err := filepath.Rel(filepath.Join(wd, path), filename)

	dir, _ := filepath.Split(relativePath)

	var ns []string

	if len(dir) > 0 {
		ns = strings.Split(dir, fmt.Sprintf("%c", os.PathSeparator))
		t.namespace = ns[:len(ns)-1]

		// rewrite tablenames
		t.Name = fmt.Sprintf("%s_%s", strings.Join(t.namespace, "_"), t.Name)

		// Ensure lower
		t.Name = strings.ToLower(t.Name)
	}

	return err
}

// RemoveNamespace Remove the underscore delimited namespace from the table
// to return it to a regular table name
func (t *Table) RemoveNamespace() {
	ns := strings.Split(t.Name, "_")
	t.Name = ns[len(ns)-1]
}

// LoadDBMetadata Populate the Metadata for this table with data from the database
func (t *Table) LoadDBMetadata() (err error) {

	var mds []metadata.Metadata

	mds, err = metadata.LoadAllTableMetadata(t.Name)

	// Check if the error is that there isn't any metadata for this table
	if err == sql.ErrNoRows {
		// which is a valid state, so surpresss the error
		err = nil
	}

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// If there isn't any metadata then this loop won't execute
	for _, md := range mds {
		// Table
		if md.ParentID == "" && md.Type == "Table" {
			t.Metadata = md
		}

		// Columns
		if md.Type == "Column" {
			for i := 0; i < len(t.Columns); i++ {
				if md.Name == t.Columns[i].Name {
					t.Columns[i].Metadata = md
				}
			}
		}

		// Primary Key
		if md.Type == "PrimaryKey" {
			t.PrimaryIndex.Metadata = md
		}

		// Indexes
		if md.Type == "Index" {
			for i := 0; i < len(t.SecondaryIndexes); i++ {
				if md.Name == t.SecondaryIndexes[i].Name {
					t.SecondaryIndexes[i].Metadata = md
				}
			}
		}
	}

	return err
}

// SyncDBMetadata Helper function to insert new Metadata and retrieves existing Metadata from the DB
func (t *Table) SyncDBMetadata() (err error) {

	// If the Table Metadata object doesn't have a Metadata Id
	// then it hasn't been loaded from the DB (it's been built from YAML)

	// Load the DB Metadata state to the table.
	err = t.LoadDBMetadata()
	if err != nil {
		return err
	}

	// Search for Metadata that still doesn't have a Metadata Id.
	// This will mean that they are new Table fields and need to be recorded in
	// the DB so that they will be detected and correctly validated when the
	// migration adding them to the DB is executed.

	// Creating a simple anon function for inserting any new Metadata
	syncDB := func(md *metadata.Metadata) error {
		// Double check that the DB doesn't know about this Metadata
		if md.MDID < 1 {
			// Ensure that a name has been set
			if len(md.Name) == 0 {
				return fmt.Errorf("Cannot create or find Metadata without a name")
			}
			// Update the DB
			return md.Insert()
		}
		// Already exists in the DB
		return nil
	}

	// Process the Table, then Columns, then PrimaryKey, and finally the Indexes
	err = syncDB(&t.Metadata)

	if util.ErrorCheckf(err, "Failed to sync Metadata for Table: [%s]", t.Name) {
		return err
	}

	err = syncDB(&t.PrimaryIndex.Metadata)

	if util.ErrorCheckf(err, "Failed to sync Metadata for Table: [%s] Primary Key", t.Name) {
		return err
	}

	for i := 0; i < len(t.Columns); i++ {
		err = syncDB(&t.Columns[i].Metadata)

		if util.ErrorCheckf(err, "Failed to sync Metadata for Table: [%s] Column: [%s]", t.Name, t.Columns[i].Name) {
			return err
		}
	}

	for i := 0; i < len(t.SecondaryIndexes); i++ {
		err = syncDB(&t.SecondaryIndexes[i].Metadata)

		if util.ErrorCheckf(err, "Failed to sync Metadata for Table: [%s] Index: [%s]", t.Name, t.SecondaryIndexes[i].Name) {
			return err
		}
	}

	return err
}

// GeneratePropertyIDs Generate PropertyIds for all Table Properties that don't have a PropertyID set.
func (t *Table) GeneratePropertyIDs() error {
	tableName := strings.ToLower(t.Metadata.Name)
	// Table
	if t.Metadata.PropertyID == "" {
		t.ID = tableName
		t.Metadata.PropertyID = t.ID
	}

	// Columns
	for i := 0; i < len(t.Columns); i++ {
		col := &t.Columns[i]
		if col.Metadata.PropertyID == "" {
			col.ID = strings.ToLower(col.Metadata.Name)
			col.Metadata.PropertyID = col.ID
			col.Metadata.ParentID = t.ID
		}
	}

	// Primary Key
	if t.PrimaryIndex.Metadata.PropertyID == "" {
		t.PrimaryIndex.ID = "primarykey"
		t.PrimaryIndex.Metadata.PropertyID = t.PrimaryIndex.ID
		t.PrimaryIndex.Metadata.ParentID = t.ID
	}

	// Indexes
	for i := 0; i < len(t.SecondaryIndexes); i++ {
		ind := &t.SecondaryIndexes[i]
		if ind.Metadata.PropertyID == "" {
			ind.ID = strings.ToLower(ind.Metadata.Name)
			ind.Metadata.PropertyID = ind.ID
			ind.Metadata.ParentID = t.ID
		}
	}

	return nil
}

// InsertMetadata Insert all missing Metadata into the Managment Metadata table
func (t *Table) InsertMetadata() (err error) {
	// Table
	err = t.Metadata.OnCreate()
	if util.ErrorCheck(err) {
		return err
	}

	// Columns
	for _, col := range t.Columns {
		err = col.Metadata.OnCreate()
		if util.ErrorCheck(err) {
			return err
		}
	}

	// Primary Key
	err = t.PrimaryIndex.Metadata.OnCreate()
	if util.ErrorCheck(err) {
		return err
	}

	// Indexes
	for _, ind := range t.SecondaryIndexes {
		err = ind.Metadata.OnCreate()
		if util.ErrorCheck(err) {
			return err
		}
	}

	return nil
}
