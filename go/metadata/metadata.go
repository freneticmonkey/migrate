package metadata

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/util"
)

// Metadata This struct stores the identification information for each table
// and table field in the target database.  This data is used to match the
// schema of the target database to the YAML schema
type Metadata struct {
	MDID       int64  `db:"mdid, autoincrement, primarykey" json:"mdid"`
	DB         int    `db:"db" json:"db"`
	PropertyID string `db:"property_id" json:"property_id"`
	ParentID   string `db:"parent_id" json:"parent_id"`
	Type       string `db:"type" json:"type"`
	Name       string `db:"name" json:"name"`
	Exists     bool   `db:"exists" json:"exists"`
}

// Insert Insert the Metadata into the Management DB
func (m *Metadata) Insert() error {
	m.DB = targetDBID
	if len(m.PropertyID) == 0 {
		return fmt.Errorf("Inserting empty Metadata")
	}
	if err := configured(); err != nil {
		return err
	}
	return mgmtDb.Insert(m)
}

// Update Update the Metadata in the Management DB
func (m *Metadata) Update() (err error) {
	if err := configured(); err != nil {
		return err
	}
	_, err = mgmtDb.Update(m)
	return err
}

// Delete Remove the Metadata from the database
func (m *Metadata) Delete() (err error) {
	if err := configured(); err != nil {
		return err
	}
	_, err = mgmtDb.Delete(m)
	return err
}

// IsTable Returns if there is a value for ParentID. If empty the property is a table.
func (m *Metadata) IsTable() bool {
	return len(m.ParentID) == 0
}

// OnCreate Check if the Metadata is known to the database yet.  Depends on MDID being set != 0
func (m *Metadata) OnCreate() error {
	// If this Metadata hasn't been inserted into the database yet, insert it
	if m.MDID == 0 {

		if err := configured(); err != nil {
			return err
		}

		return m.Insert()
	}
	return nil
}

// Load Uses the valud of MDID to load from the Management DB
func Load(mdid int64) (m *Metadata, err error) {
	var md Metadata
	if err = configured(); err != nil {
		return m, err
	}
	query := fmt.Sprintf("SELECT * FROM `metadata` WHERE mdid=%d", mdid)
	err = mgmtDb.SelectOne(&md, query)

	if err == nil {
		m = &md
	}
	return m, err
}

// TableRegistered Returns a boolean indicating that the Table named 'name' is
// registered in the Metadata table
func TableRegistered(name string) (reg bool, err error) {
	if err = configured(); err != nil {
		return false, err
	}

	query := fmt.Sprintf("SELECT count(*) from metadata WHERE name=\"%s\" and type=\"Table\"", name)
	count, err := mgmtDb.SelectInt(query)
	return count > 0, err
}

// LoadAllTableMetadata Load all of the Metadata rows for a table with the
// matching Property and Database IDs
func LoadAllTableMetadata(name string) (md []Metadata, err error) {
	var tblMd Metadata

	tblMd, err = GetTableByName(name)
	if err != nil {
		return md, err
	}

	query := fmt.Sprintf("select * from metadata WHERE name = \"%s\" OR parent_id = \"%s\" AND db=%d", tblMd.PropertyID, tblMd.PropertyID, targetDBID)
	_, err = mgmtDb.Select(&md, query)

	util.ErrorCheckf(err, "There was a problem retrieving Metadata for Table with Name: [%s] and PropertyID: [%s]", name, tblMd.PropertyID)
	return md, err
}

// GetTableByName Get a Table metadata object from the database by name
func GetTableByName(name string) (md Metadata, err error) {

	// Find a Table (Tables don't have a ParentID)
	return GetByName(name, "")
}

// GetByName Get a metadata object from the database with name
func GetByName(name string, parentID string) (md Metadata, err error) {
	if err = configured(); err != nil {
		return md, err
	}

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
