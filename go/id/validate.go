package id

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

// ValidationItem Stores the specific item detail for a validation error
type ValidationItem struct {
	Context string
	ID      string
	Name    string
	Table   string
	Type    string
	Source  string
}

func (vi ValidationItem) String() string {
	return strings.Join([]string{
		"Context: " + vi.Context,
		"ID: " + vi.ID,
		"Name: " + vi.Name,
		"Table: " + vi.Table,
		"Type: " + vi.Type,
		"Source: " + vi.Source,
	}, "\n")
}

// ValidationError Stores the details of an Error
type ValidationError struct {
	Desc  string
	Items []ValidationItem
}

func (v ValidationError) String() string {
	items := []string{}
	for _, i := range v.Items {
		items = append(items, i.String())
	}
	return fmt.Sprintf("Desc: %s\n%s", v.Desc, strings.Join(items, "\n"))
}

type ValidationErrors struct {
	Errors []*ValidationError
}

func (ve *ValidationErrors) Add(e *ValidationError) {
	ve.Errors = append(ve.Errors, e)
}

func (ve ValidationErrors) Count() int {
	return len(ve.Errors)
}

func (ve ValidationErrors) HasErrors() bool {
	return ve.Count() != 0
}

func (ve ValidationErrors) Log() {
	for _, e := range ve.Errors {
		util.LogError(e.String())
	}
}

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
func (p Properties) Exists(pid, name, tname, ptype, filename string) (e *ValidationError) {
	for i, id := range p.PropertyIds {
		if pid == id {
			util.LogErrorf(idConflictTemplate, pid, tname, name, filename, p.Table[i], p.Name[i], p.Type[i], p.PropertyIds[i], p.Filename[i])

			e = &ValidationError{
				Desc: fmt.Sprintf("PropertyID is already defined: %s", pid),
				Items: []ValidationItem{
					{
						Context: "Duplicate",
						ID:      pid,
						Name:    name,
						Table:   tname,
						Type:    ptype,
						Source:  filename,
					},
					{
						Context: "Existing",
						ID:      p.PropertyIds[i],
						Name:    p.Name[i],
						Table:   p.Table[i],
						Type:    p.Type[i],
						Source:  p.Filename[i],
					},
				},
			}
		}

		if name == p.Name[i] {
			util.LogErrorf(nameConflictTemplate, name, tname, pid, filename, p.Table[i], p.Name[i], p.Type[i], p.PropertyIds[i], p.Filename[i])

			e = &ValidationError{
				Desc: fmt.Sprintf("Name is already defined: %s", name),
				Items: []ValidationItem{
					{
						Context: "Duplicate",
						ID:      pid,
						Name:    name,
						Table:   tname,
						Type:    ptype,
						Source:  filename,
					},
					{
						Context: "Existing",
						ID:      p.PropertyIds[i],
						Name:    p.Name[i],
						Table:   p.Table[i],
						Type:    p.Type[i],
						Source:  p.Filename[i],
					},
				},
			}
		}
	}

	return e
}

// validate Generic validation function which returns 1 for an error and 0 for no error.
func validate(propertyID string, ptype string, name string, tname string, filename string, ids *Properties, vErrors *ValidationErrors) {
	var vErr *ValidationError

	// Check for missing properties
	desc := ""

	if len(propertyID) == 0 {
		desc = "MISSING_ID"
	} else if len(name) == 0 {
		desc = "MISSING_NAME"
	} else if len(ptype) == 0 {
		desc = "MISSING_TYPE"
	}

	if desc != "" {
		vErr = &ValidationError{
			Desc: desc,
			Items: []ValidationItem{
				{
					Context: desc,
					ID:      propertyID,
					Name:    name,
					Table:   tname,
					Type:    ptype,
					Source:  filename,
				},
			},
		}
	} else {
		if vErr = ids.Exists(propertyID, name, tname, ptype, filename); vErr != nil {
			ids.Add(propertyID, ptype, name, tname, filename)
		}
	}

	if vErr != nil {
		vErrors.Add(vErr)
	}
}

// ValidateSchema checks the tables parameter for duplicate names and ids.
// Ids and names cannot be shared between tables and the properties of
// individual tables
func ValidateSchema(tables table.Tables, schemaName string, log bool) (validationErrors ValidationErrors, err error) {
	var tableIds Properties

	// Check each table for unique table ids
	for _, tbl := range tables {
		validate(tbl.Metadata.PropertyID, "Table", tbl.Name, tbl.Name, tbl.Filename, &tableIds, &validationErrors)

		var tablePropertyIds Properties
		// Add table info, so that conflicts with the current table will be detected.
		tablePropertyIds.Add(tbl.Metadata.PropertyID, "Table", tbl.Name, tbl.Name, tbl.Filename)

		// Check Primary Key - if column(s) are defined.
		if len(tbl.PrimaryIndex.Columns) > 0 {
			validate(tbl.PrimaryIndex.Metadata.PropertyID, "Primary Key", "Primary Key", tbl.Name, tbl.Filename, &tablePropertyIds, &validationErrors)
		}

		for _, column := range tbl.Columns {
			validate(column.Metadata.PropertyID, "Column", column.Name, tbl.Name, tbl.Filename, &tablePropertyIds, &validationErrors)
		}

		// Check indexes
		for _, index := range tbl.SecondaryIndexes {
			validate(index.Metadata.PropertyID, "Index", index.Name, tbl.Name, tbl.Filename, &tablePropertyIds, &validationErrors)
		}
	}

	// Display validation output
	if validationErrors.HasErrors() {
		if log {
			validationErrors.Log()
		}
		err = fmt.Errorf("Reading tables from %s failed. %d problems found", schemaName, validationErrors.Count())
	} else {
		util.LogInfof("Validation Successful for Schema: %s. Validated %d tables", schemaName, len(tables))
	}

	return validationErrors, err
}
