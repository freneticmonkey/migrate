package id

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

func Validate(tableName string, paramType string, id string, ids *[]string, result *bool) {

	if len(id) == 0 {
		util.LogError(fmt.Sprintf("Table: [%s] Invalid [%s] Id Found: (%s)", tableName, paramType, id))
	}

	if util.StringInArray(id, *ids) {
		util.LogError(fmt.Sprintf("Table: [%s] Duplicate [%s] Id Found: (%s)", tableName, paramType, id))
	} else {
		*ids = append(*ids, id)
		*result = true
	}
}

func ValidateSchema(tables table.Tables) (result bool) {

	var tableIds []string
	result = true

	// Check each table for unique table ids
	for _, table := range tables {
		Validate(table.Name, "Table", table.Id, &tableIds, &result)

		var columnIds []string
		var indexIds []string
		var tablePropIds []string

		// Check Primary Key
		Validate(table.Name, "Primary Key", table.PrimaryIndex.Id, &tablePropIds, &result)

		for _, column := range table.Columns {
			Validate(table.Name, "Column", column.Id, &columnIds, &result)
		}

		// Check indexes
		for _, index := range table.SecondaryIndexes {
			Validate(table.Name, "Index", index.Id, &indexIds, &result)
		}
	}
	return result
}
