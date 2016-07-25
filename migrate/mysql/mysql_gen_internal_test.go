package mysql

import (
	"testing"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

type SQLCTTest struct {
	Table       table.Table
	Statement   string
	ExpectFail  bool
	Description string
}

var createTableTests = []SQLCTTest{

	{
		Table: table.Table{
			Name:    "TestTable",
			Engine:  "InnoDB",
			CharSet: "latin1",
			Columns: []table.Column{
				table.Column{
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
				table.Column{
					Name:     "address",
					Type:     "varchar",
					Size:     []int{128},
					Nullable: false,
					AutoInc:  false,
					Metadata: metadata.Metadata{
						Name:   "address",
						Type:   "Column",
						Exists: true,
					},
				},
			},
			PrimaryIndex: table.Index{
				IsPrimary: true,
				Columns: []string{
					"name",
					"address",
				},
			},
			SecondaryIndexes: []table.Index{
				table.Index{
					Name:      "idx_name",
					IsPrimary: false,
					Columns: []string{
						"name",
					},
				},
			},
		},
		Statement:   "CREATE TABLE `TestTable` (`name` varchar(64) NOT NULL,`address` varchar(128) NOT NULL, PRIMARY KEY (`name`,`address`), KEY `idx_name` (`name`)) ENGINE=InnoDB DEFAULT CHARSET=latin1;",
		ExpectFail:  false,
		Description: "Create Table: Two Columns, Primary Index, Secondary Index",
	},

	{
		Table: table.Table{
			Name:    "TestTable",
			Engine:  "InnoDB",
			CharSet: "latin1",
			Columns: []table.Column{
				table.Column{
					Name:     "age",
					Type:     "int",
					Size:     []int{11},
					Default:  "1",
					Nullable: false,
					AutoInc:  false,
					Metadata: metadata.Metadata{
						Name:   "age",
						Type:   "Column",
						Exists: true,
					},
				},
			},
		},
		Statement:   "CREATE TABLE `TestTable` (`age` int(11) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=latin1;",
		ExpectFail:  false,
		Description: "Create Table: No Index",
	},
}

func TestCreateTable(t *testing.T) {
	var result SQLOperation

	for _, test := range createTableTests {
		result = generateCreateTable(test.Table)

		if result.Statement != test.Statement {
			t.Errorf("%s FAILED.", test.Description)
			util.LogAttentionf(" Expecting: %s", test.Statement)
			util.LogErrorf("Generated: %s", result.Statement)
		}
	}
}

const (
	Table = iota
	Column
	Index
)

type SQLGenTest struct {
	Diff        table.Diff
	Statement   string
	ExpectFail  bool
	Description string
	TestType    int
}

var genTableAlterTests = []SQLGenTest{
	// {
	// 	Diff:        table.Diff{},
	// 	Statement:   "",
	// 	ExpectFail:  false,
	// 	Description: "No Op",
	// },

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "Name",
			Value:    "TestTableS",
		},
		Statement:   "ALTER TABLE `TestTable` RENAME TO `TestTableS`;",
		ExpectFail:  false,
		Description: "Table Rename",
		TestType:    Table,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "AutoInc",
			Value:    1234,
		},
		Statement:   "ALTER TABLE `TestTable` AUTO_INCREMENT=1234;",
		ExpectFail:  false,
		Description: "Table Change Auto Increment value",
		TestType:    Table,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "AutoInc",
			Value:    1234,
		},
		Statement:   "ALTER TABLE `TestTable` AUTO_INCREMENT=1234;",
		ExpectFail:  false,
		Description: "Table Change Auto Increment value",
		TestType:    Table,
	},
}

func TestGenerateAlters(t *testing.T) {
	var results SQLOperations

	for _, test := range genTableAlterTests {

		switch test.TestType {
		case Table:
			results = generateAlterTable(test.Diff)

		case Column:
			results = generateAlterColumn(test.Diff)

		case Index:
			results = generateAlterIndex(test.Diff)

		}

		if len(results) > 0 {
			if results[0].Statement != test.Statement {
				t.Errorf("%s FAILED.", test.Description)
				util.LogAttentionf(" Expecting: %s", test.Statement)
				util.LogErrorf("Generated: %s", results[0].Statement)
			}
		} else {
			t.Errorf("%s FAILED.", test.Description)
			util.LogAttentionf("No generated statements")
		}
	}
}
