package table

import (
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
)

type DiffOrderTest struct {
	Generated   []Diff
	Sorted      []Diff
	Forward     bool
	ExpectFail  bool
	Description string
}

var diffOrderTests = []DiffOrderTest{

	{
		Generated: []Diff{
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Add,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_address",
				Value: Index{
					ID:   "idx_address",
					Name: "idx_address",
					Columns: []IndexColumn{
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_address",
				},
			},
		},
		Sorted: []Diff{
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Add,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_address",
				Value: Index{
					ID:   "idx_address",
					Name: "idx_address",
					Columns: []IndexColumn{
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_address",
				},
			},
		},
		Forward:     true,
		ExpectFail:  false,
		Description: "Basic Sort Integrity",
	},

	{
		Generated: []Diff{
			// Add a new Column called Address
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Add,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
			// Delete the existing index
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Del,
				Property: "idx_id_address",
				Value: Index{
					ID:   "idx_id_address",
					Name: "idx_id_address",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_id_address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_id_address",
				},
			},
			// Create a new Index with the new Column
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_id_address",
				Value: Index{
					ID:   "idx_id_address",
					Name: "idx_id_address",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_id_address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_id_address",
				},
			},
		},
		Sorted: []Diff{
			// Delete the existing index
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Del,
				Property: "idx_id_address",
				Value: Index{
					ID:   "idx_id_address",
					Name: "idx_id_address",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_id_address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_id_address",
				},
			},
			// Add a new Column called Address
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Add,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},

			// Create a new Index with the new Column
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_id_address",
				Value: Index{
					ID:   "idx_id_address",
					Name: "idx_id_address",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_id_address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_id_address",
				},
			},
		},
		Forward:     true,
		ExpectFail:  false,
		Description: "Column Add Index Recreation",
	},

	{
		Generated: []Diff{

			// Delete an existing Column called Address
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Del,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
			// Delete the existing index
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Del,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
			// Create the new index without the Column
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
		},
		Sorted: []Diff{

			// Delete the existing index
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Del,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
			// Create the new index without the Column
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
			// Delete an existing Column called Address
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Del,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
		},
		Forward:     true,
		ExpectFail:  false,
		Description: "Column Del Index Recreation",
	},

	{
		Generated: []Diff{
			// Delete an existing Column called Address
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Del,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
			// Add a new Column called Phone
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Add,
				Property: "phone",
				Value: Column{
					ID:   "phone",
					Name: "Phone",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "phone",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "phone",
				},
			},
			// Drop the existing Index on Id and Address
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Del,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
			// Create a replacement index on Id and Phone
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "phone",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
		},
		Sorted: []Diff{

			// Drop the dependent index first
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Del,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "address",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
			// Add the Phone Column
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Add,
				Property: "phone",
				Value: Column{
					ID:   "phone",
					Name: "Phone",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "phone",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "phone",
				},
			},
			// Rebuild the index with Id and Phone
			{
				Table:    tblName,
				Field:    "SecondaryIndexes",
				Op:       Add,
				Property: "idx_idlookup",
				Value: Index{
					ID:   "idx_idlookup",
					Name: "idx_idlookup",
					Columns: []IndexColumn{
						{
							Name: "id",
						},
						{
							Name: "phone",
						},
					},
					IsPrimary: false,
					IsUnique:  false,
					Metadata: metadata.Metadata{
						PropertyID: "idx_idlookup",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "idx_idlookup",
				},
			},
			// Drop the Address Column
			{
				Table:    tblName,
				Field:    "Columns",
				Op:       Del,
				Property: "address",
				Value: Column{
					ID:   "address",
					Name: "Address",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						PropertyID: "address",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "address",
				},
			},
		},
		Forward:     true,
		ExpectFail:  false,
		Description: "Index Recreation w/ Dependencies on Column Del and Add",
	},
}

func TestDiffOrder(t *testing.T) {
	util.VerboseOverrideSet(true)
	for _, test := range diffOrderTests {

		util.LogAttentionf("Testing: %s", test.Description)
		sorted, err := orderDiffs(Differences{Slice: test.Generated}, test.Forward)

		// If the sort failed unexpectedly
		if err != nil && !test.ExpectFail {
			t.Errorf("%s Sort Failed with ERROR", test.Description)
			util.LogError(err)
		}

		// If the number of items from the sort is incorrect
		if len(sorted.Slice) != len(test.Sorted) {
			t.Errorf("%s Failed. Missing Diffs", test.Description)

			util.DebugDumpDiffDetail(test.Sorted, sorted.Slice, "Expected", "Received")
		} else {

			sort := sorted.Slice
			expectedSort := test.Sorted

			// If sort is INCORRECT
			if !test.ExpectFail && !reflect.DeepEqual(sort, expectedSort) {
				t.Errorf("%s Failed. Sort is not correct", test.Description)

				util.LogAttentionf("%s Failed. Incorrect Sort.", test.Description)
				util.DebugDumpDiffDetail(expectedSort, sort, "Expected", "Received")

				// if the difference is CORRECT and we expected it to be INCORRECT!
			} else if !test.ExpectFail && reflect.DeepEqual(sort, expectedSort) {
				// Success

				// if the difference is INCORECT and we expected it to be INCORRECT
			} else if test.ExpectFail && !reflect.DeepEqual(sort, expectedSort) {
				// Successfully failed
			}
		}

	}
}
