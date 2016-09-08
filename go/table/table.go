package table

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
)

// NewNamespace Initialise a new Namespace
func NewNamespace(ns *config.SchemaNamespace, tableName string) Namespace {

	if ns != nil {
		tableName = strings.TrimPrefix(tableName, ns.TablePrefix)
		return Namespace{
			SchemaName:    ns.Name,
			ShortName:     ns.ShortName,
			TablePrefix:   ns.TablePrefix,
			TableName:     tableName,
			TableFilename: tableName,
		}
	}
	return Namespace{
		TableName:     tableName,
		TableFilename: tableName,
	}

}

// Namespace Stores the namespacing metadata for the Table
type Namespace struct {
	SchemaName    string
	ShortName     string
	TablePrefix   string
	TableName     string
	TableFilename string
}

// SetExistingFilename Search the files parameter for a file matching the Table
// Will match a file either namespaced, or not.  For example:
//
// For DB table:
// 		animals_dogs
//
// Where SchemaNamespace:
// 		SchemaNamespace {
//			SchemaName:  "Animals",
//			ShortName:   "ani",
//			TablePrefix: "animals",
//			TableName:   "dogs",
//		}
//
// Namespaced:
// <root>/ani/dogs.txt
// Regular:
// <root>/dogs.txt
// Will both match.
func (tn *Namespace) SetExistingFilename(files []string) {

	tnl := strings.ToLower(tn.TableName)
	tnsn := strings.ToLower(tn.ShortName)
	for _, file := range files {
		// check if the file is in a folder
		pathPieces := strings.Split(path.Dir(file), fmt.Sprintf("%c", os.PathSeparator))

		if len(pathPieces) > 1 {
			dir := strings.ToLower(pathPieces[len(pathPieces)-1])

			// If the folder of the file doesn't match the SchemaNamespace short name,
			// ignore it.
			if dir != tnsn {
				continue
			}
			// If the folder is correct, keep checking the file
		}
		// extract the filename without the extension
		f := strings.ToLower(filepath.Base(file))
		fn := strings.TrimSuffix(f, filepath.Ext(f))

		// If the lowercase filename matches the lowercase tablename
		if strings.ToLower(fn) == tnl {
			// set the filename
			tn.TableFilename = strings.TrimSuffix(filepath.Base(file), filepath.Ext(f))
			// stop searching
			break
		}
	}
}

// GenerateFilename Generate a filename for the table given its namespace and a file extension
func (tn Namespace) GenerateFilename(ext string) string {
	path := []string{}
	if tn.ShortName != "" {
		path = append(path, tn.ShortName)
	}
	if tn.TableFilename != "" {
		path = append(path, tn.TableFilename+"."+ext)
	} else {
		path = append(path)
	}
	genpath := filepath.Join(path...)
	return genpath
}

// Tables Helper type for a slice of Table structs
type Tables []Table

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

	Namespace Namespace         `yaml:"-"`
	Filename  string            `yaml:"-"`
	Metadata  metadata.Metadata `yaml:"-"`
}

// SetNamespace Use the path and filename parameters to rename the table
// into an underscore delimited namespace
func (t *Table) SetNamespace(conf config.Config) (err error) {

	var ns *config.SchemaNamespace

	// Search for configured Schema Namespaces
	for _, sns := range conf.Project.Schema.Namespaces {
		if strings.HasPrefix(t.Name, sns.TablePrefix) {
			ns = &sns
			break
		}
	}

	t.Namespace = NewNamespace(ns, t.Name)
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
	tableID := t.Metadata.PropertyID

	// Table
	if tableID == "" {
		tableID = strings.ToLower(t.Metadata.Name)

		if tableID == "" {
			tableID = strings.ToLower(t.Name)
		}

		t.ID = tableID
		t.Metadata.PropertyID = tableID

	} else {
		t.ID = t.Metadata.PropertyID
	}

	// Columns
	for i := 0; i < len(t.Columns); i++ {
		col := &t.Columns[i]
		if col.Metadata.PropertyID == "" {
			colID := col.ID

			if colID == "" {
				colID = strings.ToLower(col.Metadata.Name)
			}

			if colID == "" {
				colID = strings.ToLower(col.Name)
			}

			col.ID = colID
			col.Metadata.PropertyID = colID
			col.Metadata.ParentID = tableID

		} else {
			col.ID = col.Metadata.PropertyID
			col.Metadata.ParentID = tableID
		}
	}

	// Primary Key
	t.PrimaryIndex.ID = "primarykey"
	t.PrimaryIndex.Metadata.PropertyID = "primarykey"
	t.PrimaryIndex.Metadata.ParentID = tableID

	// Indexes
	for i := 0; i < len(t.SecondaryIndexes); i++ {
		ind := &t.SecondaryIndexes[i]
		if ind.Metadata.PropertyID == "" {
			indID := ind.ID
			if indID == "" {
				indID = strings.ToLower(ind.Metadata.Name)
			}

			if indID == "" {
				indID = strings.ToLower(ind.Name)
			}
			ind.ID = indID
			ind.Metadata.PropertyID = indID
			ind.Metadata.ParentID = tableID

		} else {
			ind.ID = ind.Metadata.PropertyID
			ind.Metadata.ParentID = tableID
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
