package id

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
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
	Errors []ValidationError
}

func (ve *ValidationErrors) Add(e ValidationError) {
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
func (p Properties) Exists(pid, name, tname, ptype, filename string) (exists bool, e ValidationError) {
	for i, id := range p.PropertyIds {
		if pid == id {
			// util.LogErrorf(idConflictTemplate, pid, tname, name, filename, p.Table[i], p.Name[i], p.Type[i], p.PropertyIds[i], p.Filename[i])

			e = ValidationError{
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
			exists = true
		}

		if name == p.Name[i] {
			// util.LogErrorf(nameConflictTemplate, name, tname, pid, filename, p.Table[i], p.Name[i], p.Type[i], p.PropertyIds[i], p.Filename[i])

			e = ValidationError{
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
			exists = true
		}
	}

	return exists, e
}

// validate Generic validation function which returns 1 for an error and 0 for no error.
func validate(propertyID string, ptype string, name string, tname string, filename string, ids *Properties, vErrors *ValidationErrors) {
	var vErr ValidationError
	var err bool

	// Check for missing properties
	desc := ""

	if len(propertyID) == 0 {
		desc = "MISSING_ID"
	} else if len(name) == 0 {
		desc = "MISSING_NAME"
	} else if len(ptype) == 0 {
		desc = "MISSING_TYPE"
	}

	if ptype == "PrimaryKey" {
		if propertyID != "primarykey" {
			desc = "INVALID_PK_ID"
		} else if name != "PrimaryKey" {
			desc = "INVALID_PK_NAME"
		}
	}

	if desc != "" {
		vErr = ValidationError{
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
		err = true
	} else {
		if err, vErr = ids.Exists(propertyID, name, tname, ptype, filename); !err {
			ids.Add(propertyID, ptype, name, tname, filename)
		}
	}

	if err {
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
		validate(tbl.Metadata.PropertyID, tbl.Metadata.Type, tbl.Name, tbl.Name, tbl.Filename, &tableIds, &validationErrors)

		var tablePropertyIds Properties
		// Add table info, so that conflicts with the current table will be detected.
		tablePropertyIds.Add(tbl.Metadata.PropertyID, tbl.Metadata.Type, tbl.Name, tbl.Name, tbl.Filename)

		// Check Primary Key - if column(s) are defined.
		if len(tbl.PrimaryIndex.Columns) > 0 {
			validate(tbl.PrimaryIndex.Metadata.PropertyID, tbl.PrimaryIndex.Metadata.Type, tbl.PrimaryIndex.Name, tbl.Name, tbl.Filename, &tablePropertyIds, &validationErrors)
		}

		for _, column := range tbl.Columns {
			validate(column.Metadata.PropertyID, column.Metadata.Type, column.Name, tbl.Name, tbl.Filename, &tablePropertyIds, &validationErrors)
		}

		// Check indexes
		for _, index := range tbl.SecondaryIndexes {
			validate(index.Metadata.PropertyID, index.Metadata.Type, index.Name, tbl.Name, tbl.Filename, &tablePropertyIds, &validationErrors)
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

// ValidatePropertyIDs Compare the tables between the YAML and MySQL, and check if property ids have been changed
// without any change to the table and field.  This would indicate a YAML data error and the user should be
// notified.
func ValidatePropertyIDs(yamlSchema []table.Table, mysqlSchema []table.Table, log bool) (validationErrors ValidationErrors, err error) {

	// Match YAML and MySQL tables using names, validate property ids matches.  If name matches
	// but the PropertyIDs are different, then create an error.
	// NOTE: This creates a limitiation with the tool in that Tables, Columns, PrimaryKeys,
	// and Indexes cannot be deleted and recreated in the same migration, however I can't envisage a safe
	// scenario in which this would legitimately occurr.
	for _, yTable := range yamlSchema {
		for _, msTable := range mysqlSchema {
			if yTable.Name == msTable.Name {

				// Check Table
				if yTable.Metadata.PropertyID != msTable.Metadata.PropertyID {
					validationErrors.Add(ValidationError{
						Desc: fmt.Sprintf("YAML PropertyID change detected. MySQL ID: [%s]", msTable.Metadata.PropertyID),
						Items: []ValidationItem{
							{
								Context: "CHANGED_ID",
								ID:      yTable.Metadata.PropertyID,
								Name:    yTable.Name,
								Table:   yTable.Name,
								Type:    "Table",
								Source:  yTable.Filename,
							},
						},
					})
				}

				// Check the Table Fields

				// Check Primary Key - if column(s) are defined.
				if len(yTable.PrimaryIndex.Columns) > 0 {
					if yTable.PrimaryIndex.Metadata.PropertyID != "primarykey" {
						validationErrors.Add(ValidationError{
							Desc: "Invalid PropertyID for Primary Key",
							Items: []ValidationItem{
								{
									Context: "CHANGED_ID",
									ID:      yTable.PrimaryIndex.Metadata.PropertyID,
									Name:    "PrimaryKey",
									Table:   yTable.Name,
									Type:    "PrimaryKey",
									Source:  yTable.Filename,
								},
							},
						})
					}
				}

				// Check the YAML Columns
				for _, yColumn := range yTable.Columns {
					for _, msColumn := range msTable.Columns {
						if yColumn.Name == msColumn.Name {
							if yColumn.Metadata.PropertyID != msColumn.Metadata.PropertyID {
								validationErrors.Add(ValidationError{
									Desc: fmt.Sprintf("YAML PropertyID change detected. MySQL ID: [%s]", msColumn.Metadata.PropertyID),
									Items: []ValidationItem{
										{
											Context: "CHANGED_ID",
											ID:      yColumn.Metadata.PropertyID,
											Name:    yColumn.Name,
											Table:   yTable.Name,
											Type:    "Column",
											Source:  yTable.Filename,
										},
									},
								})
							}
						}
					}
				}

				// Check the YAML Indexes
				for _, yIndex := range yTable.SecondaryIndexes {
					for _, msIndex := range msTable.SecondaryIndexes {
						if yIndex.Name == msIndex.Name {
							if yIndex.Metadata.PropertyID != msIndex.Metadata.PropertyID {
								validationErrors.Add(ValidationError{
									Desc: fmt.Sprintf("YAML PropertyID change detected. MySQL ID: [%s]", msIndex.Metadata.PropertyID),
									Items: []ValidationItem{
										{
											Context: "CHANGED_ID",
											ID:      yIndex.Metadata.PropertyID,
											Name:    yIndex.Name,
											Table:   yTable.Name,
											Type:    "Index",
											Source:  yTable.Filename,
										},
									},
								})
							}
						}
					}
				}
			}
		}
	}

	if validationErrors.HasErrors() {
		if log {
			validationErrors.Log()
		}
		err = fmt.Errorf("YAML PropertyID Validation FAILED. %d problems detected", validationErrors.Count())
	} else {
		util.LogInfof("Successful YAML PropertyID Validation. Checked %d tables", len(yaml.Schema))
	}

	return validationErrors, err
}
