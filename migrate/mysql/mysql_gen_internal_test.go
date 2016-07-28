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
				Columns: []table.IndexColumn{
					{
						Name: "name",
					},
					{
						Name: "address",
					},
				},
			},
			SecondaryIndexes: []table.Index{
				table.Index{
					Name:      "idx_name",
					IsPrimary: false,
					Columns: []table.IndexColumn{
						{
							Name: "name",
						},
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
		Statement:   "CREATE TABLE `TestTable` (`age` int(11) NOT NULL DEFAULT '1') ENGINE=InnoDB DEFAULT CHARSET=latin1;",
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
	Statements  []string
	ExpectFail  bool
	Description string
	TestType    int
}

var genTableAlterTests = []SQLGenTest{
	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "Name",
			Value:    "TestTableS",
		},
		Statements: []string{
			"ALTER TABLE `TestTable` RENAME TO `TestTableS`;",
		},
		ExpectFail:  false,
		Description: "Table Rename",
		TestType:    Table,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "Engine",
			Value:    "InnoDB",
		},
		Statements: []string{
			"ALTER TABLE `TestTable` ENGINE=InnoDB;",
		},
		ExpectFail:  false,
		Description: "Table Change Engine Type",
		TestType:    Table,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "AutoInc",
			Value:    int64(1234),
		},
		Statements: []string{
			"ALTER TABLE `TestTable` AUTO_INCREMENT=1234;",
		},
		ExpectFail:  false,
		Description: "Table Change Auto Increment value",
		TestType:    Table,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Property: "CharSet",
			Value:    "french",
		},
		Statements: []string{
			"ALTER TABLE `TestTable` DEFAULT CHARACTER SET `french`;",
		},
		ExpectFail:  false,
		Description: "Table Change Character Set",
		TestType:    Table,
	},

	// Columns

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Field:    "Columns",
			Property: "Name",
			Value: table.DiffPair{
				From: table.Column{
					ID:   "col1",
					Name: "Add",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				To: table.Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "col1",
			},
		},
		Statements: []string{
			"ALTER TABLE `TestTable` CHANGE COLUMN `Add` `Address` varchar(64) NOT NULL",
		},
		ExpectFail:  false,
		Description: "Table Rename Column",
		TestType:    Column,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Field:    "Columns",
			Property: "Nullable",
			Value: table.DiffPair{
				From: table.Column{
					ID:   "col1",
					Name: "Add",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				To: table.Column{
					ID:       "col1",
					Name:     "Add",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: true,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "col1",
			},
		},
		Statements: []string{
			"ALTER TABLE `TestTable` MODIFY COLUMN `Add` varchar(64)",
		},
		ExpectFail:  false,
		Description: "Table Column: Make nullable",
		TestType:    Column,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Field:    "Columns",
			Property: "AutoInc",
			Value: table.DiffPair{
				From: table.Column{
					ID:   "col1",
					Name: "Add",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				To: table.Column{
					ID:      "col1",
					Name:    "Add",
					Type:    "varchar",
					Size:    []int{64},
					AutoInc: true,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "col1",
			},
		},
		Statements: []string{
			"ALTER TABLE `TestTable` MODIFY COLUMN `Add` varchar(64) NOT NULL AUTO_INCREMENT",
		},
		ExpectFail:  false,
		Description: "Table Column: Make Auto Increment",
		TestType:    Column,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Field:    "Columns",
			Property: "AutoInc",
			Value: table.DiffPair{
				From: table.Column{
					ID:   "col1",
					Name: "Add",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				To: table.Column{
					ID:      "col1",
					Name:    "Add",
					Type:    "varchar",
					Size:    []int{64},
					Default: "hello",
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "col1",
			},
		},
		Statements: []string{
			"ALTER TABLE `TestTable` MODIFY COLUMN `Add` varchar(64) NOT NULL DEFAULT 'hello'",
		},
		ExpectFail:  false,
		Description: "Table Column: Add DEFAULT value",
		TestType:    Column,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Op:       table.Mod,
			Field:    "Columns",
			Property: "Nullable",
			Value: table.DiffPair{
				From: table.Column{
					ID:   "col1",
					Name: "Add",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				To: table.Column{
					ID:   "col1",
					Name: "Add",
					Type: "text",
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "col1",
			},
		},
		Statements: []string{
			"ALTER TABLE `TestTable` MODIFY COLUMN `Add` text NOT NULL",
		},
		ExpectFail:  false,
		Description: "Table Column: Change type to text",
		TestType:    Column,
	},

	// Indexes
	{
		Diff: table.Diff{
			Table:    "TestTable",
			Field:    "SecondaryIndexes",
			Op:       table.Mod,
			Property: "Name",
			Value: table.DiffPair{
				From: table.Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []table.IndexColumn{
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
				To: table.Index{
					ID:   "sc1",
					Name: "idx_address",
					Columns: []table.IndexColumn{
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "sc1",
			},
		},
		Statements: []string{
			"ALTER TABLE `TestTable` RENAME idx_test idx_address",
		},
		ExpectFail:  false,
		Description: "Table Index: Rename Index",
		TestType:    Index,
	},

	{
		Diff: table.Diff{
			Table:    "TestTable",
			Field:    "SecondaryIndexes",
			Op:       table.Mod,
			Property: "Columns",
			Value: table.DiffPair{
				From: table.Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []table.IndexColumn{
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
				To: table.Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []table.IndexColumn{
						{
							Name: "address",
						},
						{
							Name: "add",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "sc1",
			},
		},
		Statements: []string{
			"DROP INDEX `idx_test` ON `TestTable`",
			"CREATE INDEX `idx_test` ON `TestTable` (`address`,`add`)",
		},
		ExpectFail:  false,
		Description: "Table Index: Add Column",
		TestType:    Index,
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

		pass := false
		for i := 0; i < len(results); i++ {
			pass = true
			if results[i].Statement != test.Statements[i] {
				t.Errorf("%s FAILED.", test.Description)
				util.LogWarnf("%s FAILED.", test.Description)
				util.DebugDiffString(test.Statements[i], results[i].Statement)
			}
		}

		if !pass {
			t.Errorf("%s FAILED. No Generated Statements", test.Description)
		}
	}
}
