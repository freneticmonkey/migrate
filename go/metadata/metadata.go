package metadata

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/test"
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
	if len(m.ParentID) == 0 && m.Type != "Table" {
		return fmt.Errorf("Inserting Table field Metadata with no ParentID. PropertyID: [%s] Name: [%s] ParentID: [%s]", m.PropertyID, m.Name, m.ParentID)
	}
	if err := configured(); err != nil {
		return err
	}
	err := mgmtDb.Insert(m)
	if usingCache {
		if err == nil {
			cache = append(cache, *m)
		}
	}
	return err
}

// Update Update the Metadata in the Management DB
func (m *Metadata) Update() (err error) {
	if err = configured(); err != nil {
		return err
	}
	_, err = mgmtDb.Update(m)

	if err == nil {
		if usingCache {
			for i := range cache {
				if cache[i].MDID == m.MDID {
					cache[i] = *m
				}
			}
			cache = append(cache, *m)
		}
	}
	return err
}

// Delete Remove the Metadata from the database
func (m *Metadata) Delete() (err error) {
	if err := configured(); err != nil {
		return err
	}
	_, err = mgmtDb.Delete(m)
	if usingCache {
		for i := range cache {
			if cache[i].MDID == m.MDID {
				cache = append(cache[:i], cache[i+1:]...)
			}
		}
	}
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

// ToDBRow Used to convert the Metadata into a unit test DBRow
func (m Metadata) ToDBRow() test.DBRow {
	return test.DBRow{
		m.MDID,
		m.DB,
		m.PropertyID,
		m.ParentID,
		m.Type,
		m.Name,
		m.Exists,
	}
}

// Load Uses the valud of MDID to load from the Management DB
func Load(mdid int64) (m *Metadata, err error) {
	var md Metadata

	if usingCache {

		for i := range cache {
			if cache[i].MDID == mdid {
				return &cache[i], nil
			}
		}
	}

	if err = configured(); err != nil {
		return m, err
	}
	query := fmt.Sprintf("SELECT * FROM `metadata` WHERE mdid=%d", mdid)
	err = mgmtDb.SelectOne(&md, query)

	if err == nil {
		m = &md
	}

	// If there was a cache miss, push it into the cache
	if usingCache {
		cache = append(cache, md)
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

// SetTableExists Set all fields on the table to exist
func SetTableExists(name string) (err error) {
	_, err = mgmtDb.Exec(fmt.Sprintf("UPDATE metadata SET `exists` = 1 WHERE name = \"%s\" OR parent_id = \"%s\" AND db = %d", name, name, targetDBID))
	return err
}

// LoadAllTableMetadata Load all of the Metadata rows for a table with the
// matching Property and Database IDs
func LoadAllTableMetadata(name string) (md []Metadata, err error) {
	var tblMd Metadata

	tblMd, err = GetTableByName(name)

	if err != nil {
		return md, err
	}

	// If there isn't any metadata for the table, early exit, all metadata needs to be created
	if tblMd.Name == "" {
		return md, err
	}

	if usingCache {

		for _, meta := range cache {
			if meta.Name == name || meta.ParentID == tblMd.PropertyID {
				md = append(md, meta)
			}
		}

	} else {

		query := fmt.Sprintf("select * from metadata WHERE name = \"%s\" OR parent_id = \"%s\" AND db=%d", tblMd.Name, tblMd.PropertyID, targetDBID)
		_, err = mgmtDb.Select(&md, query)

	}

	util.ErrorCheckf(err, "There was a problem retrieving Metadata for Table with Name: [%s] and PropertyID: [%s]", name, tblMd.PropertyID)
	return md, err
}

// MarkNonExistAllTableMetadata Delete all of a Table's metadata.
func MarkNonExistAllTableMetadata(name string) (err error) {

	_, err = mgmtDb.Exec(fmt.Sprintf("UPDATE `metadata` SET `exists` = 0 WHERE name = \"%s\" OR parent_id = \"%s\" AND db = %d", name, name, targetDBID))

	return err
}

// DeleteAllTableMetadata Delete all of a Table's metadata.
func DeleteAllTableMetadata(name string) (err error) {

	_, err = mgmtDb.Exec(fmt.Sprintf("DELETE FROM metadata WHERE name = \"%s\" OR parent_id = \"%s\" AND db = %d", name, name, targetDBID))

	return err
}

// DeleteAllTargetDBMetadata Delete all of a Table's metadata.  Intended for sandbox use.
func DeleteAllTargetDBMetadata() (err error) {

	query := fmt.Sprintf("DELETE FROM metadata WHERE db = %d", targetDBID)
	_, err = mgmtDb.Exec(query)

	return err
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

	if usingCache {

		for _, md := range cache {
			if md.Name == name && md.ParentID == parentID {
				return md, nil
			}
		}

	} else {
		query := fmt.Sprintf("SELECT * FROM metadata WHERE name=\"%s\"", name)

		if len(parentID) > 0 {
			query += fmt.Sprintf(" AND parent_id=\"%s\"", parentID)
		}

		err = mgmtDb.SelectOne(&md, query)
	}

	return md, err
}

// UpdateCache Build a localstore of the Metadata Management DB for the target DB
func UpdateCache() error {
	query := fmt.Sprintf("SELECT * FROM metadata WHERE db = %d", targetDBID)

	_, err := mgmtDb.Select(&cache, query)
	return err
}
