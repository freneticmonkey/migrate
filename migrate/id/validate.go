package id

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// Properties A helper struct which allows for easy display of validation errors
type Properties struct {
	PropertyIds []string
	Type        []string
	Name        []string
	Table       []string
	Filename    []string
}

// Add Adds the parameters to the list of known properties
func (p *Properties) Add(id string, ptype string, name string, tname string, filename string) {
	p.PropertyIds = append(p.PropertyIds, id)
	p.Type = append(p.Type, ptype)
	p.Name = append(p.Name, name)
	p.Table = append(p.Table, tname)
	p.Filename = append(p.Filename, filename)
}

// Exists Checks if the pid and name parameters exist
func (p Properties) Exists(pid string, name string, tname string, filename string) bool {
	for i, id := range p.PropertyIds {
		if pid == id {
			util.LogErrorf(idConflictTemplate, pid, tname, name, filename, p.Table[i], p.Name[i], p.Type[i], p.Filename[i])
			return true
		}

		if name == p.Name[i] {
			util.LogErrorf(nameConflictTemplate, name, tname, pid, filename, p.Table[i], p.Name[i], p.Type[i], p.Filename[i])
			return true
		}
	}

	return false
}

// validate Generic validation function which returns 1 for an error and 0 for no error.
func validate(propertyID string, ptype string, name string, tname string, filename string, ids *Properties) (result int) {
	result = 0

	if len(propertyID) == 0 {
		// util.LogError(fmt.Sprintf("Missing Id for Property: Name: [%s] Type: [%s] Table: [%s] File: [%s]", name, ptype, tname, filename))
		util.LogError(fmt.Sprintf(missingIDTemplate, propertyID, name, ptype, tname, filename))
		result = 1
	} else {
		if ids.Exists(propertyID, name, tname, filename) {
			result = 1
		} else {
			ids.Add(propertyID, ptype, name, tname, filename)
		}
	}

	return result
}

// ValidateSchema checks the tables parameter for duplicate names and ids.
// Ids and names cannot be shared between tables and the properties of
// individual tables
func ValidateSchema(tables table.Tables, schemaName string) (result int, err error) {

	var tableIds Properties

	// Check each table for unique table ids
	for _, tbl := range tables {
		result += validate(tbl.Metadata.PropertyID, "Table", tbl.Name, tbl.Name, tbl.Filename, &tableIds)

		var tablePropertyIds Properties

		// Check Primary Key
		result += validate(tbl.PrimaryIndex.Metadata.PropertyID, "Primary Key", "Primary Key", tbl.Name, tbl.Filename, &tablePropertyIds)

		for _, column := range tbl.Columns {
			result += validate(column.Metadata.PropertyID, "Column", column.Name, tbl.Name, tbl.Filename, &tablePropertyIds)
		}

		// Check indexes
		for _, index := range tbl.SecondaryIndexes {
			result += validate(index.Metadata.PropertyID, "Index", index.Name, tbl.Name, tbl.Filename, &tablePropertyIds)
		}
	}

	// Display validation output
	if result != 0 {
		err = fmt.Errorf("Reading tables from %s failed. %d problems found", schemaName, result)
	} else {
		util.LogInfof("Successfully read %d tables from %s", len(tables), schemaName)
	}

	return result, err
}
