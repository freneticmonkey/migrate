package mysql

import (
	"testing"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

type SQLGenTest struct {
	Diff        table.Diff
	Statement   string
	ExpectFail  bool
	Description string
}

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

		// Parse
		// _, err := sqlparser.Parse(result.Statement)
		// if err != nil {
		// 	util.LogErrorf("Generated: %s", result.Statement)
		// 	t.Errorf("%s FAILED SQL Parse with error: %v", test.Description, err)
		// }

		if result.Statement != test.Statement {
			t.Errorf("%s FAILED.", test.Description)
			util.LogAttentionf(" Expecting: %s", test.Statement)
			util.LogErrorf("Generated: %s", result.Statement)
		}
	}
}

var genTests = []SQLGenTest{
	{
		Diff:        table.Diff{},
		Statement:   "",
		ExpectFail:  false,
		Description: "No Op",
	},
}
