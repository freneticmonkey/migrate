package metadata

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/util"
)

// Metadata This struct stores the identification information for each table
// and table field in the target database.  This data is used to match the
// schema of the target database to the YAML schema
type Metadata struct {
	MDID       int64  `db:"mdid, autoincrement, primarykey"`
	DB         int    `db:"db"`
	PropertyID string `db:"property_id"`
	ParentID   string `db:"parent_id"`
	Type       string `db:"type"`
	Name       string `db:"name"`
	Exists     bool   `db:"exists"`
}

// Load Uses the valud of MDID to load from the Management DB
func Load(mdid int64) (m *Metadata, err error) {
	md, err := mgmtDb.Get(Metadata{}, mdid)
	if md != nil {
		m = md.(*Metadata)
	}
	return m, err
}

// Insert Insert the Metadata into the Management DB
func (m *Metadata) Insert() error {
	m.DB = targetDBID
	if len(m.PropertyID) == 0 {
		util.LogError("Inserting empty Metadata")
	}
	return mgmtDb.Insert(m)
}

// Update Update the Metadata in the Management DB
func (m *Metadata) Update() (err error) {
	_, err = mgmtDb.Update(m)
	return err
}

// Delete Remove the Metadata from the database
func (m *Metadata) Delete() (err error) {
	_, err = mgmtDb.Delete(m)
	return err
}

// IsTable Returns if there is a value for ParentID. If empty the property is a table.
func (m *Metadata) IsTable() bool {
	return len(m.ParentID) == 0
}

// OnCreate Check if the Metadata is known to the database yet.  Depends on MDID being set != 0
func (m *Metadata) OnCreate() {
	// If this Metadata hasn't been inserted into the database yet, insert it
	if m.MDID == 0 {
		m.Insert()
	}
}

// TableRegistered Returns a boolean indicating that the Table named 'name' is
// registered in the Metadata table
func TableRegistered(name string) (reg bool, err error) {
	query := fmt.Sprintf("SELECT count(*) from metadata WHERE name=\"%s\" and type=\"Table\"", name)
	count, err := mgmtDb.SelectInt(query)
	return count > 0, err
}

// GetTableByName Get a Table metadata object from the database by name
func GetTableByName(name string) (md Metadata, err error) {

	// Find a Table (Tables don't have a ParentID)
	return GetByName(name, "")
}

// GetByName Get a metadata object from the database with name
func GetByName(name string, parentID string) (md Metadata, err error) {
	errString := fmt.Sprintf("Failed to find Property with name: [%s]", name)
	query := fmt.Sprintf("SELECT * FROM metadata WHERE name=\"%s\"", name)

	if len(parentID) > 0 {
		query += fmt.Sprintf(" AND parent_id=\"%s\"", parentID)
		errString += fmt.Sprintf(" in ParentID: [%s]", parentID)
	}

	err = mgmtDb.SelectOne(&md, query)
	util.ErrorCheckf(err, errString)

	return md, err
}
