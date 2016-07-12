package mysql

import (
	"testing"

	"github.com/freneticmonkey/migrate/migrate/table"
)

var tblPropertyID = "testtbl"
var tblName = "test"

type ParseTest struct {
	Str        string      // Column definition to parse
	Expected   interface{} // Expected Column defintion
	ExpectFail bool
}

var colTests = []ParseTest{
	// General Tests
	{
		Str: "`name` varchar(64) NOT NULL",
		Expected: table.Column{
			Name:     "name",
			Type:     "varchar",
			Size:     []int{64},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` int(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "int",
			Size:     []int{11},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	// Test type parsing
	{
		Str: "`age` int(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "int",
			Size:     []int{11},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` bigint(20) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "bigint",
			Size:     []int{20},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` char(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "char",
			Size:     []int{11},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` varchar(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "varchar",
			Size:     []int{11},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` decimal(14,4) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "decimal",
			Size:     []int{14, 4},
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` text NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "text",
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` float NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "float",
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` longblob NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "longblob",
			Nullable: false,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` mediumtext",
		Expected: table.Column{
			Name:     "age",
			Type:     "mediumtext",
			Nullable: true,
			AutoInc:  false,
		},
		ExpectFail: false,
	},

	// Test DEFAULT value settings
	{
		Str: "`age` int(11) NOT NULL DEFAULT '1'",
		Expected: table.Column{
			Name:     "age",
			Type:     "int",
			Size:     []int{11},
			Default:  "1",
			Nullable: true,
			AutoInc:  false,
		},
		ExpectFail: false,
	},
	{
		Str: "`age` double DEFAULT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "double",
			Default:  "NULL",
			Nullable: true,
			AutoInc:  false,
		},
		ExpectFail: false,
	},

	// Test malformed sql parse fails
	{
		Str:        "`age` NOT NULL",
		ExpectFail: true,
	},
	{
		Str:        "`age`",
		ExpectFail: true,
	},
	{
		Str:        "`age` nottype",
		ExpectFail: true,
	},
	{
		Str:        "`age` int(sk)",
		ExpectFail: true,
	},
	{
		Str:        "`age` int(11) DEFAULT sdkjf",
		ExpectFail: true,
	},
	{
		Str:        "`age` int(11) DEFAULT ",
		ExpectFail: true,
	},
}

var indexTests = []ParseTest{
	{
		Str: "KEY `idx_id_name` (`id`,`name`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []string{
				"id",
				"name",
			},
		},
		ExpectFail: false,
	},
	{
		Str: "KEY `idx_id_name` (`id`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []string{
				"id",
			},
		},
		ExpectFail: false,
	},
	// Test Fails
	{
		Str: "KEY `idx_id_name` ()",
		Expected: table.Index{
			Name:    "idx_id_name",
			Columns: []string{},
		},
		ExpectFail: true,
	},
	{
		Str: "KEY `` (`id`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []string{
				"id",
			},
		},
		ExpectFail: true,
	},
	{
		Str: "PRIMARY KEY `idx_id_name` (`id`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []string{
				"id",
			},
		},
		ExpectFail: true,
	},
}

var pkTests = []ParseTest{
	{
		Str: "PRIMARY KEY (`id`)",
		Expected: table.Index{
			Name: table.PrimaryKey,
			Columns: []string{
				"id",
			},
		},
		ExpectFail: false,
	},
}

func validateResult(test ParseTest, result interface{}, err error, desc string, t *testing.T) {

	if !test.ExpectFail && err != nil {
		t.Errorf("%s Failed for column: '%s' with Error: '%s'", desc, test.Str, err)

	} else if !test.ExpectFail && err == nil {
		if hasDiff, diff := table.Compare(tblName, "TestColumn", result, test.Expected); hasDiff {
			t.Errorf("%s Failed with Diff: '%s'", desc, diff.Print())
		}
	} else if test.ExpectFail && err == nil {
		t.Errorf("%s Succeeded and it should have FAILED! Test String: '%s'", desc, test.Str)
	}
	// else Successfully failed :)
}

func TestColumnParse(t *testing.T) {
	var err error
	var result table.Column

	for _, test := range colTests {

		result, err = buildColumn(test.Str, tblPropertyID, tblName)
		validateResult(test, result, err, "Column Parse", t)
	}
}

func TestIndexParse(t *testing.T) {
	var err error
	var result table.Index

	for _, test := range indexTests {

		result, err = buildIndex(test.Str, tblPropertyID, tblName)
		validateResult(test, result, err, "Index Parse", t)
	}
}

func TestPKParse(t *testing.T) {
	var err error
	var result table.Index

	for _, test := range pkTests {

		result, err = buildPrimaryKey(test.Str, tblPropertyID, tblName)
		validateResult(test, result, err, "PrimaryKey Parse", t)
	}
}
