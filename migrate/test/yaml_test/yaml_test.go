package yaml_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/migrate/yaml"
)

// Successful Read
func TestRead(t *testing.T) {

	context := "unittest"
	definition := `
    name:     test
    charset:  latin1
    engine:   InnoDB
    id:       tbl1
    columns:
        - name:     id
          type:     int
          size:		[11]
          nullable: No
          id:       col1

        - name:     name
          type:     varchar
          size:		[64]
          nullable: No
          id:       col2

		- name:     age
		  type:     decimal
		  size:		[14,4]
		  nullable: No
		  id:       col3

		- name:     address
		  type:    	varchar
		  size:		[64]
		  nullable: No
		  id:       col4
	      default:  "not supplied"

    primaryindex:
        columns:
          - name
        isprimary: Yes
        id: pi

    secondaryindexes:
      - name: idx_id_name
        id: sc1
        columns:
            - id
            - name
    `

	tbl, err := yaml.ReadYAML(definition, context)

	if err != nil {
		t.Error(fmt.Sprintf("Parse Error: %v", err))
	}

	// Validate table properties
	if tbl.Name != "test" {
		t.Error(fmt.Sprintf("Table Name incorrect: Expected: 'test' Found: '%s'", tbl.Name))
	}

	if tbl.CharSet != "latin1" {
		t.Error(fmt.Sprintf("Table CharSet incorrect: Expected: 'latin1' Found: '%s'", tbl.CharSet))
	}

	if tbl.Engine != "InnoDB" {
		t.Error(fmt.Sprintf("Table Engine incorrect: Expected: 'InnoDB' Found: '%s'", tbl.Engine))
	}

	if tbl.ID != "tbl1" {
		t.Error(fmt.Sprintf("Table ID incorrect: Expected: 'tbl1' Found: '%s'", tbl.ID))
	}

	numCol := len(tbl.Columns)
	if numCol != 2 {
		t.Error(fmt.Sprintf("Table has invalid number of columns: Expected: 2 Found: %d", numCol))
	}

	// Validate Column Properties
	col := tbl.Columns[0]

	if col.Name != "id" {
		t.Error(fmt.Sprintf("Column Name incorrect: Expected: 'id' Found: '%s'", col.Name))
	}

	if col.Type != "int" {
		t.Error(fmt.Sprintf("Column Type incorrect: Expected: 'int' Found: '%s'", col.Type))
	}

	if col.Size[0] != 11 {
		t.Error(fmt.Sprintf("Column Size incorrect: Expected: '11' Found: '%d'", col.Size))
	}

	if col.Nullable != false {
		t.Error(fmt.Sprintf("Column Size incorrect: Expected: 'No' Found: '%t'", col.Nullable))
	}

	if col.ID != "col1" {
		t.Error(fmt.Sprintf("Column ID incorrect: Expected: 'col1' Found: '%s'", col.ID))
	}

	// Validate PrimaryKey Properties
	pi := tbl.PrimaryIndex

	if pi.Name != "" {
		t.Error(fmt.Sprintf("PrimaryKey Name incorrect: Expected: '' Found: '%s'", pi.Name))
	}

	if pi.ID != "pi" {
		t.Error(fmt.Sprintf("PrimaryKey ID incorrect: Expected: 'pi' Found: '%s'", pi.ID))
	}

	cols := []string{"name"}
	if !reflect.DeepEqual(pi.Columns, cols) {
		t.Error(fmt.Sprintf("PrimaryKey Columns incorrect: Expected: '%v' Found: '%v'", cols, pi.Columns))
	}

	numInd := len(tbl.SecondaryIndexes)
	if numInd != 1 {
		t.Error(fmt.Sprintf("Table has invalid number of indexes: Expected: 1 Found: %d", numInd))
	}

	si := tbl.SecondaryIndexes[0]

	if si.Name != "idx_id_name" {
		t.Error(fmt.Sprintf("Secondary Index Name incorrect: Expected: 'idx_id_name' Found: '%s'", si.Name))
	}

	if si.ID != "sc1" {
		t.Error(fmt.Sprintf("Secondary Index ID incorrect: Expected: 'sc1' Found: '%s'", si.ID))
	}

	cols = []string{"id", "name"}
	if !reflect.DeepEqual(si.Columns, cols) {
		t.Error(fmt.Sprintf("SecondaryIndexe Columns incorrect: Expected: '%v' Found: '%v'", cols, pi.Columns))
	}
}

// Unsuccessful Read
