package table

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// Tables Helper type for a slice of Table structs
type Tables []Table

// Column Stores the properties for a Column field
type Column struct {
	ID       string `yaml:"id"`
	Name     string
	Type     string
	Size     []int
	Default  string `yaml:",omitempty"`
	Nullable bool   `yaml:",omitempty"`
	AutoInc  bool   `yaml:",omitempty"`

	// Binary      bool
	// Unique      bool
	// Unsigned    bool
	// ZeroFilled  bool

	Metadata metadata.Metadata `yaml:"-"`
}

// ToSQL Formats the column into its SQL representation
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

const (
	PrimaryKey = "PrimaryKey"
)

// Index Stores the properties for a Index field
type Index struct {
	ID        string `yaml:"id"`
	Name      string
	Columns   []string
	IsPrimary bool              `yaml:",omitempty"`
	IsUnique  bool              `yaml:",omitempty"`
	Metadata  metadata.Metadata `yaml:"-"`
}

// ToSQL Formats the index into its SQL representation
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

// Table Stores the fields and properties representing a Table parsed from YAML
// or from a MySQL CREATE TABLE statement
type Table struct {
	ID               string `yaml:"id"`
	Name             string
	Engine           string
	AutoInc          int64 `yaml:",omitempty"`
	CharSet          string
	Columns          []Column
	PrimaryIndex     Index
	SecondaryIndexes []Index

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
		// TODO: Cross platform support
		ns = strings.Split(dir, "/")
		t.namespace = ns[:len(ns)-1]

		// rewrite tablenames
		t.Name = fmt.Sprintf("%s_%s", strings.Join(t.namespace, "_"), t.Name)
	}

	return err
}

// RemoveNamespace Remove the underscore delimited namespace from the table
// to return it to a regular table name
func (t *Table) RemoveNamespace() {
	ns := strings.Split(t.Name, "_")
	t.Name = ns[len(ns)-1]
}

// syncMetadata Check if the Metadata exists in the DB and insert it if it doesn't
func syncMetadata(md *metadata.Metadata) {
	// var dbmd metadata.Metadata
	dbmd, err := metadata.GetByName(md.Name, md.ParentID)
	if err != nil {
		err = md.Insert()
		util.ErrorCheckf(err, "Problem inserting %s Metdata for %s for with PropertyID: [%s]", md.Type, md.Name, md.PropertyID)
	} else {
		md.MDID = dbmd.MDID
		md.DB = dbmd.DB
	}
}

// SyncDBMetadata Helper function to insert new Metadata and retrieves existing Metadata from the DB
func (t *Table) SyncDBMetadata() (err error) {

	syncMetadata(&t.Metadata)

	syncMetadata(&t.PrimaryIndex.Metadata)

	for i := 0; i < len(t.Columns); i++ {
		syncMetadata(&t.Columns[i].Metadata)
	}

	for i := 0; i < len(t.SecondaryIndexes); i++ {
		syncMetadata(&t.SecondaryIndexes[i].Metadata)
	}

	return err
}

// GeneratePropertyIDs Generate PropertyIds for all Table Properties that don't have a PropertyID set.
func (t *Table) GeneratePropertyIDs() error {
	// Table
	if t.Metadata.PropertyID == "" {
		t.ID = util.PropertyIDGen(t.Metadata.Name)
		t.Metadata.PropertyID = t.ID
	}

	// Columns
	for i := 0; i < len(t.Columns); i++ {
		col := &t.Columns[i]
		if col.Metadata.PropertyID == "" {
			col.ID = util.PropertyIDGen(col.Name)
			col.Metadata.PropertyID = col.ID
			col.Metadata.ParentID = t.ID
		}
	}

	// Primary Key
	if t.PrimaryIndex.Metadata.PropertyID == "" {
		t.PrimaryIndex.ID = util.PropertyIDGen(t.PrimaryIndex.Name)
		t.PrimaryIndex.Metadata.PropertyID = t.PrimaryIndex.ID
		t.PrimaryIndex.Metadata.ParentID = t.ID
	}

	// Indexes
	for i := 0; i < len(t.SecondaryIndexes); i++ {
		ind := &t.SecondaryIndexes[i]
		if ind.Metadata.PropertyID == "" {
			ind.ID = util.PropertyIDGen(ind.Name)
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
