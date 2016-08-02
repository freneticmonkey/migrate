package table

import (
	"fmt"
	"log"
	"reflect"

	"github.com/fatih/color"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
)

// Difference Operations
const (
	Add = iota
	Del = iota
	Mod = iota
)

// FormatOperation Formats the difference in a human readable git style for console output
func FormatOperation(input string, op int) string {
	var prefix string
	switch op {
	case Add:
		prefix = "+++"
		color.Set(color.FgGreen)
	case Del:
		prefix = "---"
		color.Set(color.FgRed)
	case Mod:
		prefix = " M "
		color.Set(color.FgYellow)
	}
	fmtStr := fmt.Sprintf("%s %s", prefix, input)
	log.Println(fmtStr)
	color.Unset()

	return fmtStr
}

// Diff A struct whcih stores the details for an individual difference
type Diff struct {
	Table    string
	Field    string
	Op       int
	Property string
	Value    interface{}
	Metadata metadata.Metadata
}

// Print Generate a human readable string representation of the Diff
func (d Diff) Print() string {
	return FormatOperation(fmt.Sprintf("Table: [%s] Field: [%s] Property: [%s] Value: [%#v]", d.Table, d.Field, d.Property, d.Value), d.Op)
}

// DiffPair Utility struct for grouping diffs
type DiffPair struct {
	From interface{}
	To   interface{}
}

// Differences Utility struct for slices of Diffs
type Differences struct {
	Slice []Diff
}

// Add Add a Diff instance to a Differences slice
func (d *Differences) Add(diff Diff) {
	// Check to make sure that the difference is valid
	// If table name is empty, then the object is most likely empty
	if len(diff.Table) > 0 {
		d.Slice = append(d.Slice, diff)
	} else {
		// TODO: Throw Error here!!
	}
}

// Merge Merge a slice of Diffs packed as Differences
func (d *Differences) Merge(diffs Differences) {
	for _, slice := range diffs.Slice {
		d.Add(slice)
	}
}

// Quick Explaination of the Diff algorithm
//
// A simple recursive loop over ( Table->Columns)
//                              (      ->Indexes)
//
// to determine if there have been any changes between the 'from' (existing) db table(s) and the 'to' (new) db table(s)
//
// Definition of terms:
// Field - An attribute of the containing object
// Property - A value which contains multiple fields.
//
// e.g Table has fields such as Name, but it also contains a Columns property, each of which contains multiple fields.

// Compare Generic object comparison that returns the differences as a Diff struct
func Compare(tableName string, fieldName string, toContainer interface{}, fromContainer interface{}) (hasDiff bool, difference Diff) {

	toField := reflect.ValueOf(toContainer).FieldByName(fieldName)
	fromField := reflect.ValueOf(fromContainer).FieldByName(fieldName)
	switch toField.Kind() {

	case reflect.Bool:
		if toField.Bool() != fromField.Bool() {
			hasDiff = true
		}

	case reflect.String:
		if toField.String() != fromField.String() {
			hasDiff = true
		}

	case reflect.Int:
		if toField.Int() != fromField.Int() {
			hasDiff = true
		}

	case reflect.Int64:
		if toField.Int() != fromField.Int() {
			hasDiff = true
		}

	case reflect.Slice:
		// If they are different lengths, then there's diff
		if toField.Len() != fromField.Len() {
			hasDiff = true
		} else {
			// strings must be in the same order - important for indexes!
			for i := 0; i < toField.Len(); i++ {
				if toField.Index(i).String() != fromField.Index(i).String() {
					hasDiff = true
					break
				}
			}
		}
	}

	if hasDiff {
		difference.Table = tableName
		difference.Property = fieldName
		difference.Op = Mod
		difference.Value = toField.Interface()
	}

	return hasDiff, difference
}

// Check each property of the generic object
func diffProperties(tableName string, fieldName string, propertyNames []string, toProperties []interface{}, fromProperties []interface{}) (diff Differences) {

	// Init possibly modified field storage
	var existingProperties []DiffPair

	// Check for new fields
	for _, to := range toProperties {
		found := false
		toMD := reflect.ValueOf(to).FieldByName("Metadata").Interface().(metadata.Metadata)

		// If an Id is defined
		if len(toMD.PropertyID) > 0 {
			for _, from := range fromProperties {
				fromMD := reflect.ValueOf(from).FieldByName("Metadata").Interface().(metadata.Metadata)
				if toMD.PropertyID == fromMD.PropertyID {
					existingProperties = append(existingProperties, DiffPair{from, to})
					found = true
					continue
				}
			}
			if !found {
				// Pack the entire object as a diff

				// get a reflect.Value for the method,
				// turn that into an interface{},
				// turn that into a function that has the expected signature,
				// call it
				// fields := reflect.ValueOf(to).MethodByName("AsDiff").Interface().(func(bool) PropertyDiff)(true)
				name := reflect.ValueOf(to).FieldByName("Name").String()
				diff.Add(Diff{
					Table:    tableName,
					Field:    fieldName,
					Op:       Add,
					Property: name,
					Value:    reflect.ValueOf(to).Interface(),
					Metadata: toMD,
				})
			}
		}
	}

	// Check for deleted fields
	for _, from := range fromProperties {
		found := false
		fromMD := reflect.ValueOf(from).FieldByName("Metadata").Interface().(metadata.Metadata)
		for _, to := range toProperties {
			toMD := reflect.ValueOf(to).FieldByName("Metadata").Interface().(metadata.Metadata)
			if toMD.PropertyID == fromMD.PropertyID {
				found = true
				continue
			}
		}
		if !found {
			// Pack the entire object as a diff

			// get a reflect.Value for the method,
			// turn that into an interface{},
			// turn that into a function that has the expected signature,
			// call it
			name := reflect.ValueOf(from).FieldByName("Name").String()
			diff.Add(Diff{
				Table:    tableName,
				Field:    fieldName,
				Op:       Del,
				Property: name,
				Value:    reflect.ValueOf(from).Interface(),
				Metadata: fromMD,
			})
		}
	}

	// Check for differences in existing fields
	for _, existingProperty := range existingProperties {
		// For each field
		for _, field := range propertyNames {

			if diffFound, difference := Compare(tableName, field, existingProperty.To, existingProperty.From); diffFound {

				difference.Field = fieldName
				difference.Value = reflect.ValueOf(existingProperty).Interface()
				difference.Metadata = reflect.ValueOf(existingProperty.To).FieldByName("Metadata").Interface().(metadata.Metadata)
				diff.Add(difference)
			}
		}
	}

	return diff
}

func diffColumns(toTable Table, fromTable Table) (hasDiff bool, differences Differences) {
	// Ugly, but it works?
	toColumns := make([]interface{}, len(toTable.Columns))
	for i, v := range toTable.Columns {
		toColumns[i] = v
	}

	fromColumns := make([]interface{}, len(fromTable.Columns))
	for i, v := range fromTable.Columns {
		fromColumns[i] = v
	}

	// Column Properties
	fieldNames := []string{"Name", "Type", "Size", "Nullable", "AutoInc", "Default", "Collation"}
	if differentColumns := diffProperties(toTable.Name, "Columns", fieldNames, toColumns, fromColumns); len(differentColumns.Slice) > 0 {
		hasDiff = true

		differences.Merge(differentColumns)
	}

	return hasDiff, differences
}

func diffIndexes(toTable Table, fromTable Table) (hasDiff bool, differences Differences) {
	// Ugly, but it works?
	toIndexes := make([]interface{}, 1)
	toIndexes[0] = toTable.PrimaryIndex
	fromIndexes := make([]interface{}, 1)
	fromIndexes[0] = fromTable.PrimaryIndex

	// Primary Index Properties
	fieldNames := []string{"Columns", "IsPrimary", "PropertyID"}

	if primaryIndex := diffProperties(toTable.Name, "PrimaryIndex", fieldNames, toIndexes, fromIndexes); len(primaryIndex.Slice) > 0 {
		hasDiff = true
		differences.Merge(primaryIndex)
	}

	// Ugly, but it works?
	toIndexes = make([]interface{}, len(toTable.SecondaryIndexes))
	for i, v := range toTable.SecondaryIndexes {
		toIndexes[i] = v
	}

	fromIndexes = make([]interface{}, len(fromTable.SecondaryIndexes))
	for i, v := range fromTable.SecondaryIndexes {
		fromIndexes[i] = v
	}

	// Index Properties
	fieldNames = []string{"Name", "Columns", "IsPrimary", "IsUnique"}

	if differentIndexes := diffProperties(toTable.Name, "SecondaryIndexes", fieldNames, toIndexes, fromIndexes); len(differentIndexes.Slice) > 0 {
		hasDiff = true

		differences.Merge(differentIndexes)
	}

	return hasDiff, differences
}

func diffTable(toTable Table, fromTable Table) (hasDiff bool, differences Differences) {
	hasDiff = false

	// Table Fields
	fieldNames := []string{"Name", "Engine", "CharSet", "AutoInc", "RowFormat", "Collation"}

	for _, field := range fieldNames {
		if diffFound, fieldsDiff := Compare(fromTable.Name, field, toTable, fromTable); diffFound {
			hasDiff = diffFound
			fieldsDiff.Metadata = fromTable.Metadata
			differences.Add(fieldsDiff)
		}
	}

	// Table Columns
	if diffFound, columnsDiff := diffColumns(toTable, fromTable); diffFound {
		hasDiff = diffFound
		differences.Merge(columnsDiff)
	}

	// Table Indexes
	if diffFound, indexesDiff := diffIndexes(toTable, fromTable); diffFound {
		hasDiff = diffFound
		differences.Merge(indexesDiff)
	}

	return hasDiff, differences
}

// DiffTables Compare the toTables and fromTables Slices of Table structs and
// return a Differences Slice containing all of the differences between the tables.
func DiffTables(toTables []Table, fromTables []Table) (tableDiffs Differences, err error) {
	util.LogInfo("Starting Diff")

	// Search through the input tables
	for i := 0; i < len(toTables); i++ {

		toTable := toTables[i]

		found := false

		// Sync the metadata for the table and it's fields to the DB so that it can be
		// detected by the Migration when it executes
		err = toTable.SyncDBMetadata()
		if util.ErrorCheckf(err, "Problem syncing Metadata with DB for Table: [%s]", toTable.Name) {
			return tableDiffs, err
		}

		// match against mysql tables
		for _, fromTable := range fromTables {

			if toTable.Metadata.PropertyID == fromTable.Metadata.PropertyID {
				found = true
				if hasDiff, diff := diffTable(toTable, fromTable); hasDiff {
					tableDiffs.Merge(diff)
				}
			}
		}
		if !found {
			// The table is a new table
			tableDiffs.Add(Diff{
				Table:    toTable.Name,
				Field:    "*",
				Op:       Add,
				Property: "*",
				Value:    toTable,
				Metadata: toTable.Metadata,
			})
		}
	}

	// Search through the existing tables for dropped tables
	for _, fromTable := range fromTables {
		found := false
		// match against the new tables
		for _, toTable := range toTables {
			if toTable.Metadata.PropertyID == fromTable.Metadata.PropertyID {
				found = true
				break
			}
		}
		if !found {
			// The table doesn't exist in the set of new tables and so it needs to be deleted
			tableDiffs.Add(Diff{
				Table:    fromTable.Name,
				Field:    "*",
				Op:       Del,
				Property: "*",
				Value:    fromTable,
				Metadata: fromTable.Metadata,
			})
		}
	}

	return tableDiffs, err
}