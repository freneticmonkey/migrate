package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// SQLOperation Stores the data associated with an alter operation for each migration step
type SQLOperation struct {
	Statement string
	Op        int
	Name      string
	Metadata  metadata.Metadata
}

// SQLOperations Slice helper type
type SQLOperations []SQLOperation

// Add Add an SQLOperation to the slice
func (s *SQLOperations) Add(op SQLOperation) {
	*s = append(*s, op)
}

// Merge Merge SQLOperations slices
func (s *SQLOperations) Merge(ops SQLOperations) {
	for _, slice := range ops {
		if len(slice.Statement) > 0 {
			s.Add(slice)
		}
	}
}

// StatementBuilder Helper for building SQL ALTER TABLE statements.
// Assists with string concatenation and SQL specific formatting requirements.
type StatementBuilder struct {
	Components []string
}

// Reset Empties the statements components
func (sb *StatementBuilder) Reset() {
	sb.Components = []string{}
}

// Add Add a new statement component
func (sb *StatementBuilder) Add(component string) {
	if component != "" {
		sb.Components = append(sb.Components, component)
	}
}

// AddFormat Add a new formatted statement component
func (sb *StatementBuilder) AddFormat(format string, info ...interface{}) {
	canAdd := true
	for _, i := range info {
		if reflect.TypeOf(i).Kind() == reflect.String {
			if i == "" {
				canAdd = false
				break
			}
		}
	}
	if canAdd {
		sb.Components = append(sb.Components, fmt.Sprintf(format, info...))
	}
}

// AddQuote Add a new component with `quotes`
func (sb *StatementBuilder) AddQuote(component string) {
	if component != "" {
		sb.Components = append(sb.Components, fmt.Sprintf("`%s`", component))
	}
}

// AddType Add a component which correctly formats type size it it's provided
func (sb *StatementBuilder) AddType(typename string, size []int) {
	// Support for decimal places makes the size a little complicated
	if typename != "" {
		colType := ""

		switch len(size) {
		case 2:
			colType = fmt.Sprintf("%s(%d,%d)", typename, size[0], size[1])
		case 1:
			colType = fmt.Sprintf("%s(%d)", typename, size[0])
		default:
			colType = typename
		}
		sb.Add(colType)
	}

}

// Format Produce a formatted String
func (sb StatementBuilder) Format() string {
	return strings.Join(sb.Components, " ") + ";"
}

// generateCreateTable Generate a MySQL CREATE TABLE statement from a
// Table struct
func generateCreateTable(tbl table.Table) (operation SQLOperation) {
	var builder StatementBuilder

	operation.Op = table.Add

	// tableTemplate := "CREATE TABLE `%s` (%s%s) ENGINE=%s%s DEFAULT CHARSET=%s;"

	// Setup Statement
	builder.Add("CREATE TABLE")
	builder.AddQuote(tbl.Name)

	columns := []string{}
	isAutoInc := false

	for _, col := range tbl.Columns {
		columns = append(columns, col.ToSQL())
		if col.AutoInc {
			isAutoInc = true
		}
	}

	indexes := []string{}

	if tbl.PrimaryIndex.IsValid() {
		indexes = append(indexes, tbl.PrimaryIndex.ToSQL())
	}

	for _, ind := range tbl.SecondaryIndexes {
		if ind.IsValid() {
			indexes = append(indexes, ind.ToSQL())
		}
	}

	strIndexes := ""
	if len(indexes) > 0 {
		strIndexes = ", " + strings.Join(indexes, ",")
	}

	// Add Columns and Indexes
	builder.Add(fmt.Sprintf("(%s%s)", strings.Join(columns, ","), strIndexes))

	// Add Table Options
	builder.AddFormat("ENGINE=%s", tbl.Engine)

	// If AUTO_INCREMENT is being used and it has a non-zero value
	if isAutoInc && tbl.AutoInc > 0 {
		builder.AddFormat("AUTO_INCREMENT=%d", tbl.AutoInc)
	}

	builder.AddFormat("DEFAULT CHARSET=%s", tbl.CharSet)

	if len(tbl.RowFormat) > 0 {
		builder.AddFormat("ROW_FORMAT=%s", tbl.RowFormat)
	}

	if len(tbl.Collation) > 0 {
		builder.AddFormat("COLLATE=%s", tbl.Collation)
	}

	operation.Statement = builder.Format()
	operation.Metadata = tbl.Metadata
	return operation
}

// generateAlterColumn Generate a MySQL ALTER COLUMN statement from a
// Table struct
func generateAlterColumn(diff table.Diff) (ops SQLOperations) {
	var operation SQLOperation
	var builder StatementBuilder

	operation.Op = diff.Op
	operation.Metadata = diff.Metadata
	operation.Name = diff.Metadata.Name

	builder.Add("ALTER TABLE")

	switch diff.Op {

	case table.Add:
		builder.AddQuote(diff.Table)
		builder.Add("COLUMN")
		builder.AddQuote(diff.Property)

		column, ok := diff.Value.(table.Column)
		if ok {
			builder.AddType(column.Type, column.Size)

			if !column.Nullable {
				builder.Add("NOT NULL")
			}
		}

	case table.Del:
		builder.AddQuote(diff.Table)
		builder.Add("DROP COLUMN")
		builder.AddQuote(diff.Property)

	case table.Mod:
		// Process modification by type

		diffPair := diff.Value.(table.DiffPair)
		toColumn := diffPair.To.(table.Column)
		fromColumn := diffPair.From.(table.Column)

		builder.AddQuote(diff.Table)

		// Name needs special handling because it requires a different number of components
		// Assuming that we are modifying the column definition by default
		if diff.Property == "Name" {
			builder.Add("CHANGE COLUMN")
			builder.AddFormat("`%s` `%s`", fromColumn.Name, toColumn.Name)
			operation.Name = toColumn.Name

		} else {
			builder.Add("MODIFY COLUMN")
			builder.AddQuote(fromColumn.Name)
		}

		// Support for decimal places makes the size a little complicated
		builder.AddType(toColumn.Type, toColumn.Size)

		// if Nullable is T or F
		if !toColumn.Nullable {
			builder.Add("NOT NULL")
		}

		// if AutoInc is T or F
		if toColumn.AutoInc {
			builder.Add("AUTO_INCREMENT")
		}

		// if a Default value is defined
		if len(toColumn.Default) > 0 {
			if toColumn.Default == NULL {
				builder.Add("DEFAULT NULL")
			} else {
				builder.AddFormat("DEFAULT '%s'", toColumn.Default)
			}
		}
	}

	operation.Statement = builder.Format()

	ops.Add(operation)
	return ops
}

// generateAlterIndex Generate a MySQL ALTER INDEX statement from a
// Table struct
func generateAlterIndex(diff table.Diff) (ops SQLOperations) {

	var builder StatementBuilder

	// Obtain Index Object
	diffPair := diff.Value.(table.DiffPair)
	toIndex, ok := diffPair.To.(table.Index)

	if ok {
		indexName := ""

		if diff.Field == "PrimaryIndex" {
			indexName = "PRIMARY KEY"

		} else if diff.Field == "SecondaryIndexes" {
			indexName = fmt.Sprintf("%s", toIndex.Name)
		}

		// Drop
		builder.Add("DROP INDEX")
		builder.AddQuote(indexName)
		builder.Add("ON")
		builder.AddQuote(diff.Table)

		removeOp := SQLOperation{
			Statement: builder.Format(),
			Op:        table.Del,
			Metadata:  diff.Metadata,
		}

		builder.Reset()
		builder.Add("CREATE INDEX")
		builder.AddQuote(indexName)
		builder.Add("ON")
		builder.AddQuote(diff.Table)
		builder.Add(toIndex.ColumnsSQL())

		addOp := SQLOperation{
			Statement: builder.Format(),
			Op:        table.Add,
			Metadata:  diff.Metadata,
		}

		switch diff.Op {

		case table.Add:
			ops.Add(addOp)

		case table.Del:
			ops.Add(removeOp)

		case table.Mod:
			// Process modification by type
			if diff.Property == "Name" {

				fromIndex, ok := diffPair.From.(table.Index)
				if ok {
					builder.Reset()
					builder.Add("ALTER TABLE")
					builder.AddQuote(diff.Table)
					builder.Add("RENAME")
					builder.AddFormat("%s %s", fromIndex.Name, toIndex.Name)

					ops.Add(SQLOperation{
						Statement: builder.Format(),
						Op:        table.Mod,
						Metadata:  diff.Metadata,
					})
				} else {
					util.LogError("Gen SQL: ALTER INDEX: MOD: Could not obtain from index")
				}
			} else {
				// if anything other than a rename, we need to drop the index and re-add
				ops.Add(removeOp)
				ops.Add(addOp)
			}
		}
	} else {
		util.LogError("Obtaining Index FAILED")
	}
	return ops
}

// generateAlterTable Generate a MySQL CREATE TABLE, DROP TABLE or ALTER TABLE statement from a
// Table struct
func generateAlterTable(diff table.Diff) (ops SQLOperations) {

	if diff.Field == "*" && diff.Property == "*" {
		switch diff.Op {
		case table.Add:
			// generate the create table
			tbl, ok := diff.Value.(table.Table)
			if ok {
				ops.Add(generateCreateTable(tbl))
			} else {
				util.LogError("ISSUES obtaining table object: " + diff.Table)
			}
		case table.Del:
			// generate the drop table
			ops.Add(SQLOperation{
				Statement: fmt.Sprintf("DROP TABLE `%s`;", diff.Table),
				Op:        table.Del,
				Name:      diff.Table,
				Metadata:  diff.Metadata,
			})
		}

	} else {
		switch diff.Property {

		case "Name":
			newTableName, ok := diff.Value.(string)
			if ok {
				ops.Add(SQLOperation{
					Statement: fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`;", diff.Table, newTableName),
					Op:        table.Mod,
					Name:      newTableName,
					Metadata:  diff.Metadata,
				})
			} else {
				util.LogError("ISSUES obtaining table name for rename: " + diff.Table)
			}

		case "AutoInc":
			ops.Add(SQLOperation{
				Statement: fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT=%d;", diff.Table, diff.Value),
				Op:        table.Mod,
				Name:      diff.Table,
				Metadata:  diff.Metadata,
			})

		case "Engine":
			ops.Add(SQLOperation{
				Statement: fmt.Sprintf("ALTER TABLE `%s` ENGINE=%s;", diff.Table, diff.Value),
				Op:        table.Mod,
				Name:      diff.Table,
				Metadata:  diff.Metadata,
			})

		case "CharSet":
			ops.Add(SQLOperation{
				Statement: fmt.Sprintf("ALTER TABLE `%s` DEFAULT CHARACTER SET `%s`;", diff.Table, diff.Value),
				Op:        table.Mod,
				Name:      diff.Table,
				Metadata:  diff.Metadata,
			})
		}
	}

	return ops

}

// GenerateAlters Generate MySQL ALTER TABLE statements from the Differences
// between Table structs
func GenerateAlters(differences table.Differences) (operations SQLOperations) {

	for _, diff := range differences.Slice {
		var alter SQLOperations

		// Check if the Diff is for a table
		if diff.Field == "Columns" {
			// It's a column change.
			alter = generateAlterColumn(diff)

		} else if diff.Field == "PrimaryIndex" || diff.Field == "SecondaryIndexes" {
			// It's an index change.
			alter = generateAlterIndex(diff)

		} else {
			alter = generateAlterTable(diff)
		}
		operations.Merge(alter)
	}

	util.LogInfo("Generated MySQL")
	for _, ops := range operations {
		table.FormatOperation(ops.Statement, ops.Op)
	}

	return operations
}
