package mysql

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

type SQLOperation struct {
	Statement string
	Op        int
	Metadata  metadata.Metadata
}

type SQLOperations []SQLOperation

func (s *SQLOperations) Add(op SQLOperation) {
	*s = append(*s, op)
}

func (s *SQLOperations) Merge(ops SQLOperations) {
	for _, slice := range ops {
		if len(slice.Statement) > 0 {
			s.Add(slice)
		}
	}
}

// generateCreateTable Generate a MySQL CREATE TABLE statement from a
// Table struct
func generateCreateTable(tbl table.Table) (ops SQLOperations) {

	var operation SQLOperation
	operation.Op = table.Add

	tableTemplate := "CREATE TABLE\t`%s` (\n\t%s,%s \n) ENGINE=%s DEFAULT CHARSET=%s;"

	columns := []string{}

	for _, col := range tbl.Columns {
		columns = append(columns, col.ToSQL())
	}

	indexes := []string{}
	indexes = append(indexes, tbl.PrimaryIndex.ToSQL())

	for _, ind := range tbl.SecondaryIndexes {
		indexes = append(indexes, ind.ToSQL())
	}
	operation.Statement = fmt.Sprintf(tableTemplate, tbl.Name, strings.Join(columns, ",\n\t"), strings.Join(indexes, ",\n\t"), tbl.Engine, tbl.CharSet)
	operation.Metadata = tbl.Metadata
	ops.Add(operation)
	return ops
}

// generateAlterColumn Generate a MySQL ALTER COLUMN statement from a
// Table struct
func generateAlterColumn(diff table.Diff) (ops SQLOperations) {
	var operation SQLOperation
	operation.Op = diff.Op

	dropTemplate := "ALTER  TABLE `%s` DROP %s;"
	addTemplate := "ALTER  TABLE `%s` ADD %s `%s` %s;"
	addColumnTemplate := "%s(%d) %s"

	modTemplate := "ALTER  TABLE     `%s` %s;"

	switch diff.Op {

	case table.Add:
		definition := ""

		column, ok := diff.Value.(table.Column)
		if ok {
			nullable := ""
			if !column.Nullable {
				nullable = "NOT NULL"
			}
			definition = fmt.Sprintf(addColumnTemplate, column.Type, column.Size, nullable)
		}

		operation.Statement = fmt.Sprintf(addTemplate, diff.Table, "COLUMN", diff.Property, definition)

	case table.Del:
		operation.Statement = fmt.Sprintf(dropTemplate, diff.Table, fmt.Sprintf("%s `%s`", "COLUMN", diff.Property))

	case table.Mod:
		// Process modification by type
		modStatement := ""

		diffPair := diff.Value.(table.DiffPair)
		toColumn := diffPair.To.(table.Column)
		fromColumn := diffPair.From.(table.Column)

		// Assuming that we are modifying the column definition by default
		columnOperation := "MODIFY COLUMN"

		name := fromColumn.Name
		colType := fmt.Sprintf(" %s(%d) ", fromColumn.Type, fromColumn.Size)
		isNull := ""
		if !fromColumn.Nullable {
			isNull = "NOT NULL"
		}

		switch diff.Property {
		case "Name":
			// if rename
			name = fmt.Sprintf("%s %s", name, toColumn.Name)
			// Use the correct MySQL Operation when renaming
			columnOperation = "CHANGE COLUMN"

		case "Type", "Size":
			// if changed type or size
			colType = fmt.Sprintf(" %s(%d) ", toColumn.Type, toColumn.Size)

		case "Nullable":
			// if Nullable is T or F
			if !toColumn.Nullable {
				isNull = "NOT NULL"
			}
		}

		modStatement = fmt.Sprintf("%s %s %s %s", columnOperation, name, colType, isNull)

		operation.Statement = fmt.Sprintf(modTemplate, diff.Table, modStatement)

	}
	ops.Add(operation)
	return ops
}

// generateAlterIndex Generate a MySQL ALTER INDEX statement from a
// Table struct
func generateAlterIndex(diff table.Diff) (ops SQLOperations) {
	dropTemplate := "ALTER  TABLE `%s` DROP %s;"
	addTemplate := "ALTER  TABLE `%s` ADD %s `%s` %s;"
	addIndexTemplate := "%s %s"
	renameIndexTemplate := "ALTER  TABLE `%s` RENAME %s "

	// Obtain Index Object
	diffPair := diff.Value.(table.DiffPair)
	toIndex, ok := diffPair.To.(table.Index)

	if ok {
		indexName := ""
		columns := fmt.Sprintf("(%s)", strings.Join(toIndex.Columns, ", "))

		if diff.Field == "PrimaryIndex" {
			indexName = "PRIMARY KEY"

		} else if diff.Field == "SecondaryIndexes" {
			indexName = fmt.Sprintf("`%s`", toIndex.Name)
		}

		indexDefinition := fmt.Sprintf(addIndexTemplate, indexName, columns)

		removeOp := SQLOperation{
			Statement: fmt.Sprintf(dropTemplate, diff.Table, indexName),
			Op:        table.Del,
			Metadata:  diff.Metadata,
		}
		addOp := SQLOperation{
			Statement: fmt.Sprintf(addTemplate, diff.Table, indexName, diff.Property, indexDefinition),
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
					renameStatement := fmt.Sprintf(renameIndexTemplate, diff.Table, fmt.Sprintf("%s %s", fromIndex.Name, toIndex.Name))
					ops.Add(SQLOperation{
						Statement: renameStatement,
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

// GenerateAlters Generate MySQL ALTER TABLE statements from the Differences
// between Table structs
func GenerateAlters(differences table.Differences) (operations SQLOperations) {

	for _, diff := range differences.Slice {
		var alter SQLOperations

		// Check if the Diff is for a table
		if diff.Field == "*" && diff.Property == "*" {
			switch diff.Op {
			case table.Add:
				// generate the create table
				tbl, ok := diff.Value.(table.Table)
				if ok {
					alter = generateCreateTable(tbl)
				} else {
					util.LogError("ISSUES obtaining table object: " + diff.Table)
				}
			case table.Del:
				// generate the drop table
				alter.Add(SQLOperation{
					Statement: fmt.Sprintf("DROP TABLE `%s`;", diff.Table),
					Op:        table.Del,
					Metadata:  diff.Metadata,
				})
			}

		} else if diff.Property == "Name" {
			tableName, ok := diff.Value.(string)
			if ok {
				alter.Add(SQLOperation{
					Statement: fmt.Sprintf("ALTER  TABLE `%s` RENAME TO `%s`;", diff.Table, tableName),
					Op:        table.Mod,
					Metadata:  diff.Metadata,
				})
			} else {
				util.LogError("ISSUES obtaining table name for rename: " + diff.Table)
			}

		} else if diff.Field == "Columns" {
			// It's a column change.
			alter = generateAlterColumn(diff)

		} else if diff.Field == "PrimaryIndex" || diff.Field == "SecondaryIndexes" {
			// It's an index change.
			alter = generateAlterIndex(diff)

		}
		operations.Merge(alter)
	}

	util.LogInfo("Generated MySQL")
	for _, ops := range operations {
		table.FormatOperation(ops.Statement, ops.Op)
	}

	return operations
}