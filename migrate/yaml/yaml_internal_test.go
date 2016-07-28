package yaml

import (
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

var tblPropertyID = "testtbl"
var tblName = "test"

type ParseTest struct {
	Str         string      // Column definition to parse
	Expected    interface{} // Expected Column defintion
	ExpectFail  bool
	Description string
}

var yamlTests = []ParseTest{
	// Test Table struct parsing
	{
		Str: `
        name:      test
        charset:   latin1
        engine:    InnoDB
        id:        tbl1
        autoinc:   1234
        rowformat: DYNAMIC
        collation: utf8_bin
        `,
		Expected: table.Table{
			Name:      "test",
			CharSet:   "latin1",
			Engine:    "InnoDB",
			ID:        "tbl1",
			AutoInc:   1234,
			RowFormat: "DYNAMIC",
			Collation: "utf8_bin",
		},
		ExpectFail:  false,
		Description: "YAML Parse: Table Options",
	},
	// Column Parsing
	// Single Column
	{
		Str: `
        columns:
            - name:     id
              type:     int
              size:     [11]
              nullable: No
              id:       col1
        `,
		Expected: table.Table{
			Columns: []table.Column{
				table.Column{
					ID:       "col1",
					Name:     "id",
					Type:     "int",
					Size:     []int{11},
					Nullable: false,
					AutoInc:  false,
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: Basic Column",
	},
	// Single Column
	{
		Str: `
        columns:
            - name:      id
              type:      int
              size:      [11]
              nullable:  No
              id:        col1
              autoinc:   Yes
              collation: utf8_bin
        `,
		Expected: table.Table{
			Columns: []table.Column{
				table.Column{
					ID:        "col1",
					Name:      "id",
					Type:      "int",
					Size:      []int{11},
					Nullable:  false,
					AutoInc:   true,
					Collation: "utf8_bin",
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: Basic Column all options",
	},
	// Multi-Column
	{
		Str: `
        columns:
            - name:     id
              type:     int
              size:     [11]
              nullable: No
              id:       col1

            - name:     name
              type:     varchar
              size:     [64]
              nullable: No
              id:       col2
        `,
		Expected: table.Table{
			Columns: []table.Column{
				table.Column{
					ID:       "col1",
					Name:     "id",
					Type:     "int",
					Size:     []int{11},
					Nullable: false,
					AutoInc:  false,
				},
				table.Column{
					ID:       "col2",
					Name:     "name",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					AutoInc:  false,
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: Multi Column",
	},
	// PrimaryKey Parsing
	{
		Str: `
        primaryindex:
            columns:
                - name: name
            isprimary: Yes
            id:        pi
        `,
		Expected: table.Table{
			PrimaryIndex: table.Index{
				ID:        "pi",
				IsPrimary: true,
				Columns: []table.IndexColumn{
					{
						Name: "name",
					},
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: PrimaryIndex",
	},
	// Index Parsing
	{
		Str: `
        secondaryindexes:
            - name: idx_name
              id:   sc1
              columns:
                  - name: name
        `,
		Expected: table.Table{
			SecondaryIndexes: []table.Index{
				{
					ID:   "sc1",
					Name: "idx_name",
					Columns: []table.IndexColumn{
						{
							Name: "name",
						},
					},
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: Secondary Index Basic",
	},
	{
		Str: `
        secondaryindexes:
            - name: idx_id_name
              id:   sc1
              columns:
                  - name: name
                  - name: id
        `,
		Expected: table.Table{
			SecondaryIndexes: []table.Index{
				{
					ID:   "sc1",
					Name: "idx_id_name",
					Columns: []table.IndexColumn{
						{
							Name: "name",
						},
						{
							Name: "id",
						},
					},
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: SecondaryIndexes multi column",
	},

	{
		Str: `
        secondaryindexes:
            - name: idx_id_name
              id:   sc1
              columns:
                  - name: name
                    length: 20
                  - name: id
        `,
		Expected: table.Table{
			SecondaryIndexes: []table.Index{
				{
					ID:   "sc1",
					Name: "idx_id_name",
					Columns: []table.IndexColumn{
						{
							Name:   "name",
							Length: 20,
						},
						{
							Name: "id",
						},
					},
				},
			},
		},
		ExpectFail:  false,
		Description: "YAML Parse: SecondaryIndexes multi column with partial index",
	},
}

func validateResult(test ParseTest, result interface{}, err error, t *testing.T) {

	if !test.ExpectFail && err != nil {
		t.Errorf("%s Failed for column: '%s' with Error: '%s'", test.Description, test.Str, err)

	} else if !test.ExpectFail && err == nil {

		if !reflect.DeepEqual(result, test.Expected) {
			t.Errorf("%s Failed. Return object differs from expected object.", test.Description)
			util.LogAttentionf("%s Failed. Return object differs from expected object.", test.Description)
			util.DebugDumpDiff(test.Expected, result)
		}
	} else if test.ExpectFail && err == nil {
		t.Errorf("%s Succeeded and it should have FAILED! Test String: '%s'", test.Description, test.Str)
	}
	// else Successfully failed :)
}

func TestYAMLParse(t *testing.T) {

	var err error
	var result table.Table

	for _, test := range yamlTests {

		result, err = ReadYAML(test.Str, "unittest")

		validateResult(test, result, err, t)
	}

}
