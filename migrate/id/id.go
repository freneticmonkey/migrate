package id

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

var idConflictTemplate = `
Duplicate id found for:
Name: [%s]
ID: [%s]
File: [%s]
-------------
ID already assigned to:
Name: [%s]
Type: [%s]
File: [%s]
=============
`

var nameConflictTemplate = `
Duplicate name found for:
Name: [%s]
ID: [%s]
File: [%s]
-------------
Name already assigned to:
Name: [%s]
Type: [%s]
File: [%s]
=============
`

// Properties A helper struct which allows for easy display of validation errors
type Properties struct {
	PropertyIds []string
	Type        []string
	Name        []string
	Filename    []string
}

// Add Adds the parameters to the list of known properties
func (p *Properties) Add(id string, ptype string, name string, filename string) {
	p.PropertyIds = append(p.PropertyIds, id)
	p.Type = append(p.Type, ptype)
	p.Name = append(p.Name, name)
	p.Filename = append(p.Filename, filename)
}

// Exists Checks if the pid and name parameters exist
func (p Properties) Exists(pid string, name string, filename string) bool {
	for i, id := range p.PropertyIds {
		if pid == id {
			util.LogErrorf(idConflictTemplate, name, pid, filename, p.Name[i], p.Type[i], p.Filename[i])
			return true
		}

		if name == p.Name[i] {
			util.LogErrorf(nameConflictTemplate, name, pid, filename, p.Name[i], p.Type[i], p.Filename[i])
			return true
		}
	}

	return false
}

// validate Generic validation function which returns 1 for an error and 0 for no error.
func validate(propertyID string, ptype string, name string, filename string, ids *Properties) int {

	if len(propertyID) == 0 {
		util.LogError(fmt.Sprintf("Invalid Id Found for Property: Name: [%s] Type: [%s] File: [%s]", name, ptype, filename))
	}

	if ids.Exists(propertyID, name, filename) {
		return 1
	}

	ids.Add(propertyID, ptype, name, filename)
	return 0
}

// ValidateSchema checks the tables parameter for duplicate names and ids.
// Ids and names cannot be shared between tables and the properties of
// individual tables
func ValidateSchema(tables table.Tables) (result int) {

	var tableIds Properties

	// Check each table for unique table ids
	for _, table := range tables {
		result += validate(table.PropertyID, "Table", table.Name, table.Filename, &tableIds)

		var tablePropertyIds Properties

		// Check Primary Key
		result += validate(table.PrimaryIndex.PropertyID, "Primary Key", "Primary Key", table.Filename, &tablePropertyIds)

		for _, column := range table.Columns {
			result += validate(column.PropertyID, "Column", column.Name, table.Filename, &tablePropertyIds)
		}

		// Check indexes
		for _, index := range table.SecondaryIndexes {
			result += validate(index.PropertyID, "Index", index.Name, table.Filename, &tablePropertyIds)
		}
	}
	return result
}
