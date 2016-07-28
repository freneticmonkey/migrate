package yaml_test

import (
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
)

var tblPropertyID = "testtbl"
var tblName = "test"

type ParseTest struct {
	Str        string      // Column definition to parse
	Expected   interface{} // Expected Column defintion
	ExpectFail bool
}

var yamlTests = []ParseTest{
	// Test Table struct parsing
	{
		Str: `
        name:     test
        charset:  latin1
        engine:   InnoDB
        id:       tbl1
        autoinc:  1234
        `,
		Expected: table.Table{
			Name:    "test",
			CharSet: "latin1",
			Engine:  "InnoDB",
			ID:      "tbl1",
			AutoInc: 1234,
		},
		ExpectFail: false,
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
		ExpectFail: false,
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
		ExpectFail: false,
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
		ExpectFail: false,
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
		ExpectFail: false,
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
		ExpectFail: false,
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
		ExpectFail: false,
	},
}

func validateResult(test ParseTest, result interface{}, err error, desc string, t *testing.T) {

	if !test.ExpectFail && err != nil {
		t.Errorf("%s Failed for column: '%s' with Error: '%s'", desc, test.Str, err)

	} else if !test.ExpectFail && err == nil {

		if !reflect.DeepEqual(result, test.Expected) {
			t.Errorf("%s Failed. Return object differs from expected object.", desc)
			util.LogAttentionf("%s Failed. Return object differs from expected object.", desc)
			util.LogWarn("Expected")
			util.DebugDump(test.Expected)
			util.LogWarn("Result")
			util.DebugDump(result)
		}
	} else if test.ExpectFail && err == nil {
		t.Errorf("%s Succeeded and it should have FAILED! Test String: '%s'", desc, test.Str)
	}
	// else Successfully failed :)
}

func TestYAMLParse(t *testing.T) {

	var err error
	var result table.Table

	for _, test := range yamlTests {

		result, err = yaml.ReadYAML(test.Str, "unittest")

		validateResult(test, result, err, "YAML Column Parse", t)
	}

}
