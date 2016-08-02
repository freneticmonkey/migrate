package mysql

import (
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

var tblPropertyID = "testtbl"
var tblName = "test"

type ParseTest struct {
	Str         string      // Column definition to parse
	Expected    interface{} // Expected Column defintion
	ExpectFail  bool
	Description string
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
			Metadata: metadata.Metadata{
				Name:   "name",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: varchar not null",
	},
	{
		Str: "`age` int(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "int",
			Size:     []int{11},
			Nullable: false,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: int not null",
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
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
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
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: bigint not null",
	},
	{
		Str: "`age` char(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "char",
			Size:     []int{11},
			Nullable: false,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: char not null",
	},
	{
		Str: "`age` decimal(14,4) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "decimal",
			Size:     []int{14, 4},
			Nullable: false,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
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
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: text not null",
	},
	{
		Str: "`age` float NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "float",
			Nullable: false,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: float not null",
	},
	{
		Str: "`age` longblob NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "longblob",
			Nullable: false,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: longblob not null",
	},
	{
		Str: "`age` mediumtext",
		Expected: table.Column{
			Name:     "age",
			Type:     "mediumtext",
			Nullable: true,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: mediumtext not null",
	},

	// Test DEFAULT value settings
	{
		Str: "`age` int(11) NOT NULL DEFAULT '1'",
		Expected: table.Column{
			Name:     "age",
			Type:     "int",
			Size:     []int{11},
			Default:  "1",
			Nullable: false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: int not null default",
	},
	{
		Str: "`age` double DEFAULT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "double",
			Default:  "NULL",
			Nullable: true,
			AutoInc:  false,
			Metadata: metadata.Metadata{
				Name:   "age",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: double default null",
	},

	{
		Str: "`address` text COLLATE utf8_bin",
		Expected: table.Column{
			Name:      "address",
			Type:      "text",
			Nullable:  true,
			Collation: "utf8_bin",
			Metadata: metadata.Metadata{
				Name:   "address",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: text with utf8_bin collation",
	},

	{
		Str: "`count` int(11) AUTO_INCREMENT",
		Expected: table.Column{
			Name:     "count",
			Type:     "int",
			Size:     []int{11},
			Nullable: true,
			AutoInc:  true,
			Metadata: metadata.Metadata{
				Name:   "count",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: int auto increment",
	},

	{
		Str: "`count` int(11) DEFAULT '1'",
		Expected: table.Column{
			Name:     "count",
			Type:     "int",
			Size:     []int{11},
			Nullable: true,
			Default:  "1",
			Metadata: metadata.Metadata{
				Name:   "count",
				Type:   "Column",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Column: int default '1'",
	},

	// Test malformed sql parse fails
	{
		Str:         "`age` NOT NULL",
		ExpectFail:  true,
		Description: "Parse Column: Test FAIL missing type",
	},
	{
		Str:         "`age`",
		ExpectFail:  true,
		Description: "Parse Column: Test FAIL missing everything",
	},
	{
		Str:         "`age` nottype",
		ExpectFail:  true,
		Description: "Parse Column: Test FAIL invalid type",
	},
	{
		Str:         "`age` int(sk)",
		ExpectFail:  true,
		Description: "Parse Column: Test FAIL invalid size",
	},
	{
		Str:         "`age` int(11) DEFAULT sdkjf",
		ExpectFail:  true,
		Description: "Parse Column: Test FAIL invalid default value format",
	},
	{
		Str:         "`age` int(11) DEFAULT ",
		ExpectFail:  true,
		Description: "Parse Column: Test FAIL missing default",
	},
}

var indexTests = []ParseTest{
	{
		Str: "KEY `idx_id_name` (`id`,`name`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
				{
					Name: "name",
				},
			},
			Metadata: metadata.Metadata{
				Name:   "idx_id_name",
				Type:   "Index",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Index: Basic Index multi column",
	},
	{
		Str: "KEY `idx_id_name` (`id`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				Name:   "idx_id_name",
				Type:   "Index",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Index: Basic Index single column",
	},
	{
		Str: "KEY `idx_id_name` (id)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				Name:   "idx_id_name",
				Type:   "Index",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Index: Basic Index single column no quotes",
	},
	// Test Fails
	{
		Str: "KEY `idx_id_name` ()",
		Expected: table.Index{
			Name:    "idx_id_name",
			Columns: []table.IndexColumn{},
		},
		ExpectFail:  true,
		Description: "Parse Index: Test Fail no columns",
	},
	{
		Str: "KEY `` (`id`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
		},
		ExpectFail:  true,
		Description: "Parse Index: Test Fail no name",
	},
	{
		Str: "PRIMARY KEY `idx_id_name` (`id`)",
		Expected: table.Index{
			Name: "idx_id_name",
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
		},
		ExpectFail:  true,
		Description: "Parse Index: Test Fail name on Primary Key",
	},

	{
		Str: "KEY `id name` (`id`,`name`)",
		Expected: table.Index{
			Name: "id name",
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
				{
					Name: "name",
				},
			},
			Metadata: metadata.Metadata{
				Name:   "id name",
				Type:   "Index",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Index: Test Name with spaces",
	},

	{
		Str: "KEY `idx_name` (`name`(20))",
		Expected: table.Index{
			Name: "idx_name",
			Columns: []table.IndexColumn{
				{
					Name:   "name",
					Length: 20,
				},
			},
			Metadata: metadata.Metadata{
				Name:   "idx_name",
				Type:   "Index",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Index: Test partial column",
	},
}

var pkTests = []ParseTest{
	{
		Str: "PRIMARY KEY (`id`)",
		Expected: table.Index{
			Name: table.PrimaryKey,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			IsPrimary: true,
			Metadata: metadata.Metadata{
				Name:   "PrimaryKey",
				Type:   "PrimaryKey",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Primary Key: Test single column",
	},
	{
		Str: "PRIMARY KEY (`id`,`name`)",
		Expected: table.Index{
			Name: table.PrimaryKey,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
				{
					Name: "name",
				},
			},
			IsPrimary: true,
			Metadata: metadata.Metadata{
				Name:   "PrimaryKey",
				Type:   "PrimaryKey",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Primary Key: Test multiple columns",
	},
	{
		Str: "PRIMARY KEY (id,name)",
		Expected: table.Index{
			Name: table.PrimaryKey,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
				{
					Name: "name",
				},
			},
			IsPrimary: true,
			Metadata: metadata.Metadata{
				Name:   "PrimaryKey",
				Type:   "PrimaryKey",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Parse Primary Key: Test multiple columns no quotes",
	},
	// Test failures
	{
		Str: "KEY (`id`)",
		Expected: table.Index{
			Name: table.PrimaryKey,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
		},
		ExpectFail:  true,
		Description: "Parse Primary Key: Test malformed Primary Key definition",
	},
	{
		Str: "PRIMARY KEY ()",
		Expected: table.Index{
			Name:    table.PrimaryKey,
			Columns: []table.IndexColumn{},
		},
		ExpectFail:  true,
		Description: "Parse Primary Key: Test malformed Primary Key column definition",
	},
}

func validateResult(test ParseTest, result interface{}, err error, t *testing.T) {

	if !test.ExpectFail && err != nil {
		t.Errorf("%s FAILED for column: '%s' with Error: '%s'", test.Description, test.Str, err)

	} else if !test.ExpectFail && err == nil {
		if hasDiff, diff := table.Compare(tblName, "TestColumn", result, test.Expected); hasDiff {
			t.Errorf("%s FAILED with Diff: '%s'", test.Description, diff.Print())
		} else {
			if !reflect.DeepEqual(result, test.Expected) {
				t.Errorf("%s FAILED. Return object differs from expected object.", test.Description)
				util.LogAttentionf("%s FAILED. Return object differs from expected object.", test.Description)
				util.DebugDumpDiff(test.Expected, result)
			}
		}
	} else if test.ExpectFail && err == nil {
		t.Errorf("%s Succeeded and it should have FAILED! Test String: '%s'", test.Description, test.Str)
	}
	// else Successfully failed :)
}

func TestColumnParse(t *testing.T) {
	var err error
	var result table.Column

	// mgmtDb, mgmtMock, err := test.CreateMockDB()
	//
	// if err != nil {
	// 	t.Errorf("TestDiffSchema: Setup Project DB Failed with Error: %v", err)
	// }
	//
	// // Configure metadata
	// metadata.Setup(mgmtDb, 1)

	for _, tst := range colTests {

		// // Search Metadata for `dogs` table query - MySQL
		// query := test.DBQueryMock{
		// 	Type: test.QueryCmd,
		//
		// 	Columns: []string{
		// 		"mdid",
		// 		"db",
		// 		"property_id",
		// 		"parent_id",
		// 		"type",
		// 		"name",
		// 		"exists",
		// 	},
		// 	Rows: [][]driver.Value{
		// 		{1, 1, "tbl1", "", "Table", "dogs", 1},
		// 	},
		// }
		//
		// query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", tst.Expected.(table.Column).Name, "testtbl")
		// test.ExpectDB(mgmtMock, query)

		result, err = buildColumn(tst.Str, tblPropertyID, tblName)
		validateResult(tst, result, err, t)
	}
}

func TestIndexParse(t *testing.T) {
	var err error
	var result table.Index

	for _, test := range indexTests {

		result, err = buildIndex(test.Str, tblPropertyID, tblName)
		validateResult(test, result, err, t)
	}
}

func TestPKParse(t *testing.T) {
	var err error
	var result table.Index

	for _, test := range pkTests {

		result, err = buildPrimaryKey(test.Str, tblPropertyID, tblName)
		validateResult(test, result, err, t)
	}
}
