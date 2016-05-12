package metadata

import "github.com/freneticmonkey/migrate/migrate/util"

// Metadata This struct stores the identification information for each table
// and table field in the target database.  This data is used to match the
// schema of the target database to the YAML schema
type Metadata struct {
	MDID       int64  `db:"mdid, autoincrement, primarykey"`
	PropertyID string `db:"property_id"`
	ParentID   string `db:"parent_id"`
	Type       string `db:"type"`
	Name       string `db:"name"`
}

// Insert Insert the Metadata into the Management DB
func (m *Metadata) Insert() {
	mgmtDb.Insert(m)
}

// Update Update the Metadata in the Management DB
func (m *Metadata) Update() {
	mgmtDb.Update(m)
}

// GetTableByName Get a Table metadata object from the database by name
func GetTableByName(name string) (md Metadata, err error) {

	// Find a Table (Tables don't have a ParentID)
	return GetByName(name, "")
}

// GetByName Get a metadata object from the database with name
func GetByName(name string, parentID string) (md Metadata, err error) {
	err = mgmtDb.SelectOne(&md, "SELECT * FROM metadata WHERE name=? AND parent_id=?", name, parentID)
	util.ErrorCheckf(err, "Failed to find Property with name: [%s] in ParentID: [%s]", name, parentID)

	return md, err
}
