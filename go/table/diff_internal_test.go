package table

import (
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
)

var tblPropertyID = "testtbl"
var tblName = "test"

type DiffTest struct {
	From        Table
	To          Table
	Expected    []Diff
	ExpectFail  bool
	Description string
}

var diffTests = []DiffTest{
	// Table Differences - No modifications
	{
		From: Table{
			Name:    "TestTable",
			Engine:  "InnoDB",
			AutoInc: 1234,
			CharSet: "latin1",
		},
		To: Table{
			Name:    "TestTable",
			Engine:  "InnoDB",
			AutoInc: 1234,
			CharSet: "latin1",
		},
		Expected:    []Diff{},
		ExpectFail:  false,
		Description: "Bulk Table Field Diff: No differences",
	},
	// Table Differences
	{
		From: Table{
			Name: "TestTable",
		},
		To: Table{
			Name: "TestTableS",
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Property: "Name",
				Value:    "TestTableS",
			},
		},
		ExpectFail:  false,
		Description: "Table Field Diff: Name",
	},
	{
		From: Table{
			Name:   "TestTable",
			Engine: "InnoDB",
		},
		To: Table{
			Name:   "TestTable",
			Engine: "InnoDB2",
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Property: "Engine",
				Value:    "InnoDB2",
			},
		},
		ExpectFail:  false,
		Description: "Table Field Diff: Engine",
	},
	{
		From: Table{
			Name:    "TestTable",
			AutoInc: 1234,
		},
		To: Table{
			Name:    "TestTable",
			AutoInc: 2468,
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Property: "AutoInc",
				Value:    int64(2468),
			},
		},
		ExpectFail:  false,
		Description: "Table Field Diff: AutoInc",
	},
	{
		From: Table{
			Name:      "TestTable",
			RowFormat: "DYNAMIC",
		},
		To: Table{
			Name:      "TestTable",
			RowFormat: "FIXED",
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Property: "RowFormat",
				Value:    "FIXED",
			},
		},
		ExpectFail:  false,
		Description: "Table Field Diff: RowFormat",
	},
	{
		From: Table{
			Name:      "TestTable",
			Collation: "utf8_bin",
		},
		To: Table{
			Name:      "TestTable",
			Collation: "latin1_german2_ci",
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Property: "Collation",
				Value:    "latin1_german2_ci",
			},
		},
		ExpectFail:  false,
		Description: "Table Field Diff: Changing Table collation to latin1_german2_ci",
	},
	{
		From: Table{
			Name:    "TestTable",
			CharSet: "latin1",
		},
		To: Table{
			Name:    "TestTable",
			CharSet: "french",
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Property: "CharSet",
				Value:    "french",
			},
		},
		ExpectFail:  false,
		Description: "Table Field Diff: CharSet",
	},

	// Column Differences
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
				},
			},
		},
		Expected:    []Diff{},
		ExpectFail:  false,
		Description: "Column Field Diff: No differences",
	},
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Add",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Name",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Add",
						Type: "varchar",
						Size: []int{64},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
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
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Rename",
	},
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "text",
					Size:     []int{64},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Type",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Address",
						Type: "varchar",
						Size: []int{64},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:   "col1",
						Name: "Address",
						Type: "text",
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
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Change type",
	},
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{12},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Size",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Address",
						Type: "varchar",
						Size: []int{64},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:   "col1",
						Name: "Address",
						Type: "text",
						Size: []int{12},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "col1",
				},
			},
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Change size",
	},
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: true,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Nullable",
				Value: DiffPair{
					From: Column{
						ID:       "col1",
						Name:     "Address",
						Type:     "varchar",
						Size:     []int{64},
						Nullable: false,
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:       "col1",
						Name:     "Address",
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
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Change NOT NULL",
	},
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					AutoInc:  false,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:       "col1",
					Name:     "Address",
					Type:     "varchar",
					Size:     []int{64},
					Nullable: false,
					AutoInc:  true,
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "AutoInc",
				Value: DiffPair{
					From: Column{
						ID:       "col1",
						Name:     "Address",
						Type:     "varchar",
						Size:     []int{64},
						Nullable: false,
						AutoInc:  false,
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:       "col1",
						Name:     "Address",
						Type:     "varchar",
						Size:     []int{64},
						Nullable: false,
						AutoInc:  true,
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "col1",
				},
			},
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Change Set AutoInc",
	},

	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "text",
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:        "col1",
					Name:      "Address",
					Type:      "text",
					Collation: "utf8_bin",
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Collation",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Address",
						Type: "text",
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:        "col1",
						Name:      "Address",
						Type:      "text",
						Collation: "utf8_bin",
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "col1",
				},
			},
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Add collation to column",
	},

	// Removing Column Size
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Size",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Address",
						Type: "varchar",
						Size: []int{64},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:   "col1",
						Name: "Address",
						Type: "varchar",
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "col1",
				},
			},
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Removing Size",
	},

	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "text",
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Type",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Address",
						Type: "varchar",
						Size: []int{64},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:   "col1",
						Name: "Address",
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
			Diff{
				Table:    "TestTable",
				Op:       Mod,
				Field:    "Columns",
				Property: "Size",
				Value: DiffPair{
					From: Column{
						ID:   "col1",
						Name: "Address",
						Type: "varchar",
						Size: []int{64},
						Metadata: metadata.Metadata{
							PropertyID: "col1",
						},
					},
					To: Column{
						ID:   "col1",
						Name: "Address",
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
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Changing varchar to text",
	},

	// Adding a Column
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				Column{
					ID:   "col1",
					Name: "Age",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						PropertyID: "col2",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Field:    "Columns",
				Op:       Add,
				Property: "Age",
				Value: Column{
					ID:   "col1",
					Name: "Age",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						PropertyID: "col2",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "col2",
				},
			},
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Adding a column",
	},

	// Deleting a Column
	{
		From: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				Column{
					ID:   "col1",
					Name: "Age",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						PropertyID: "col2",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Field:    "Columns",
				Op:       Del,
				Property: "Age",
				Value: Column{
					ID:   "col1",
					Name: "Age",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						PropertyID: "col2",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "col2",
				},
			},
		},
		ExpectFail:  false,
		Description: "Column Field Diff: Deleting a column",
	},

	// Primary Key

	{
		From: Table{
			Name: "TestTable",
			PrimaryIndex: Index{
				Columns: []IndexColumn{
					{
						Name: "address",
					},
					{
						Name: "age",
					},
				},
				IsPrimary: true,
				IsUnique:  false,
				Metadata: metadata.Metadata{
					PropertyID: "pk1",
				},
			},
		},
		To: Table{
			Name: "TestTable",
			PrimaryIndex: Index{
				Columns: []IndexColumn{
					{
						Name: "address",
					},
					{
						Name: "age",
					},
				},
				IsPrimary: true,
				IsUnique:  false,
				Metadata: metadata.Metadata{
					PropertyID: "pk1",
				},
			},
		},
		Expected:    []Diff{},
		ExpectFail:  false,
		Description: "Primary Key Field Diff: No Differences",
	},

	{
		From: Table{
			Name: "TestTable",
			PrimaryIndex: Index{
				Columns: []IndexColumn{
					{
						Name: "address",
					},
					{
						Name: "age",
					},
				},
				IsPrimary: true,
				IsUnique:  false,
				Metadata: metadata.Metadata{
					PropertyID: "pk1",
				},
			},
		},
		To: Table{
			Name: "TestTable",
			PrimaryIndex: Index{
				Columns: []IndexColumn{
					{
						Name: "address",
					},
				},
				IsPrimary: true,
				IsUnique:  false,
				Metadata: metadata.Metadata{
					PropertyID: "pk1",
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Field:    "PrimaryIndex",
				Op:       Mod,
				Property: "Columns",
				Value: DiffPair{
					From: Index{
						Columns: []IndexColumn{
							{
								Name: "address",
							},
							{
								Name: "age",
							},
						},
						IsPrimary: true,
						IsUnique:  false,
						Metadata: metadata.Metadata{
							PropertyID: "pk1",
						},
					},
					To: Index{
						Columns: []IndexColumn{
							{
								Name: "address",
							},
						},
						IsPrimary: true,
						IsUnique:  false,
						Metadata: metadata.Metadata{
							PropertyID: "pk1",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "pk1",
				},
			},
		},
		ExpectFail:  false,
		Description: "Primary Key Field Diff: Change Columns",
	},

	// Secondary Indexes

	{
		From: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
						{
							Name: "address",
						},
						{
							Name: "age",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
						{
							Name: "address",
						},
						{
							Name: "age",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
			},
		},
		Expected:    []Diff{},
		ExpectFail:  false,
		Description: "Secondary Index Field Diff: No Differences",
	},

	{
		From: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
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
		},
		To: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_address",
					Columns: []IndexColumn{
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
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Field:    "SecondaryIndexes",
				Op:       Mod,
				Property: "Name",
				Value: DiffPair{
					From: Index{
						ID:   "sc1",
						Name: "idx_test",
						Columns: []IndexColumn{
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
					To: Index{
						ID:   "sc1",
						Name: "idx_address",
						Columns: []IndexColumn{
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
		},
		ExpectFail:  false,
		Description: "Secondary Index Field Diff: Renamed index",
	},

	{
		From: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
						{
							Name: "address",
						},
						{
							Name: "age",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
			},
		},
		To: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
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
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Field:    "SecondaryIndexes",
				Op:       Mod,
				Property: "Columns",
				Value: DiffPair{
					From: Index{
						ID:   "sc1",
						Name: "idx_test",
						Columns: []IndexColumn{
							{
								Name: "address",
							},
							{
								Name: "age",
							},
						},
						IsPrimary: false,
						IsUnique:  false,
						Metadata: metadata.Metadata{
							PropertyID: "sc1",
						},
					},
					To: Index{
						ID:   "sc1",
						Name: "idx_test",
						Columns: []IndexColumn{
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
		},
		ExpectFail:  false,
		Description: "Secondary Index Field Diff: Removed a Column",
	},

	{
		From: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
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
		},
		To: Table{
			Name: "TestTable",
			SecondaryIndexes: []Index{
				Index{
					ID:   "sc1",
					Name: "idx_test",
					Columns: []IndexColumn{
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  true,
					Metadata: metadata.Metadata{
						PropertyID: "sc1",
					},
				},
			},
		},
		Expected: []Diff{
			Diff{
				Table:    "TestTable",
				Field:    "SecondaryIndexes",
				Op:       Mod,
				Property: "IsUnique",
				Value: DiffPair{
					From: Index{
						ID:   "sc1",
						Name: "idx_test",
						Columns: []IndexColumn{
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
					To: Index{
						ID:   "sc1",
						Name: "idx_test",
						Columns: []IndexColumn{
							{
								Name: "address",
							},
						},
						IsPrimary: false,
						IsUnique:  true,
						Metadata: metadata.Metadata{
							PropertyID: "sc1",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "sc1",
				},
			},
		},
		ExpectFail:  false,
		Description: "Secondary Index Field Diff: Change IsUnqiue",
	},
}

func TestDifferences(t *testing.T) {

	for _, test := range diffTests {
		if hasDiff, difference := diffTable(test.To, test.From); hasDiff {
			diffs := difference.Slice

			// If difference is INCORRECT
			if !test.ExpectFail && !reflect.DeepEqual(diffs, test.Expected) {
				t.Errorf("%s Failed. Difference is not correct", test.Description)

				util.LogAttentionf("%s Failed. Return object differs from expected object.", test.Description)
				util.LogWarn("Expected")
				util.DebugDump(test.Expected)

				util.LogWarn("Result")
				util.DebugDump(diffs)

				// if the difference is CORRECT and we expected it to be INCORRECT!
			} else if !test.ExpectFail && reflect.DeepEqual(diffs, test.Expected) {
				// Success

				// if the difference is INCORECT and we expected it to be INCORRECT
			} else if test.ExpectFail && !reflect.DeepEqual(diffs, test.Expected) {
				// Successfully failed
			}
		}

	}
}

func TestTables(t *testing.T) {

	var fromTables = []Table{
		Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
			},
		},
	}

	var toTables = []Table{
		Table{
			Name: "TestTable",
			Columns: []Column{
				Column{
					ID:   "col1",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "col1",
					},
				},
				Column{
					ID:   "col1",
					Name: "Age",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						PropertyID: "col2",
					},
				},
			},
		},
	}

	var expectedDiffs = []Diff{
		Diff{
			Table:    "TestTable",
			Field:    "Columns",
			Op:       Add,
			Property: "Age",
			Value: Column{
				ID:   "col1",
				Name: "Age",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					PropertyID: "col2",
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "col2",
			},
		},
	}

	// TODO: Fix Mock DB to get this working :(
	var diffs Differences
	if false {
		diffs, _ = DiffTables(toTables, fromTables, true)
	}

	if len(diffs.Slice) > 0 {
		if !reflect.DeepEqual(diffs, expectedDiffs) {
			t.Errorf("Tables Difference Failed. Difference is not correct")

			util.LogAttentionf("Tables Difference Failed. Return object differs from expected object.")
			util.LogWarn("Expected")
			util.DebugDump(expectedDiffs)

			util.LogWarn("Result")
			util.DebugDump(diffs)
		}
	}
}
