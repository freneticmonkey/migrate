package id

import (
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

type validationTest struct {
	Description string
	YAMLSchema  []table.Table
	MySQLSchema []table.Table
	ExpectFail  bool
	Problems    ValidationErrors
}

var validationTests = []validationTest{
	{
		Description: "Standard Validation: No Errors",
		ExpectFail:  false,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				PrimaryIndex: table.Index{
					ID:        "primarykey",
					Name:      "PrimaryKey",
					IsPrimary: true,
					Metadata: metadata.Metadata{
						PropertyID: "primarykey",
						Name:       "primarykey",
						Type:       "PrimaryKey",
						ParentID:   "dogs",
					},
				},
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name",
						Name: "idx_name",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		MySQLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				PrimaryIndex: table.Index{
					ID:        "primarykey",
					Name:      "PrimaryKey",
					IsPrimary: true,
					Metadata: metadata.Metadata{
						PropertyID: "primarykey",
						Name:       "primarykey",
						Type:       "PrimaryKey",
						ParentID:   "dogs",
					},
				},
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name",
						Name: "idx_name",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
	},

	{
		Description: "Duplicate Table Names",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs1",
				Name: "dogs",
				Metadata: metadata.Metadata{
					PropertyID: "dogs1",
					Name:       "dogs",
					Type:       "Table",
				},
			},
			{
				ID:   "dogs2",
				Name: "dogs",
				Metadata: metadata.Metadata{
					PropertyID: "dogs2",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "Name is already defined: dogs",
					Items: []ValidationItem{
						{
							Context: "Duplicate",
							ID:      "dogs2",
							Name:    "dogs",
							Table:   "dogs",
							Type:    "Table",
							Source:  "",
						},
						{
							Context: "Existing",
							ID:      "dogs1",
							Name:    "dogs",
							Table:   "dogs",
							Type:    "Table",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Duplicate Table PropertyIDs",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs_one",
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs_one",
					Type:       "Table",
				},
			},
			{
				ID:   "dogs",
				Name: "dogs_two",
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs_two",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "PropertyID is already defined: dogs",
					Items: []ValidationItem{
						{
							Context: "Duplicate",
							ID:      "dogs",
							Name:    "dogs_two",
							Table:   "dogs_two",
							Type:    "Table",
							Source:  "",
						},
						{
							Context: "Existing",
							ID:      "dogs",
							Name:    "dogs_one",
							Table:   "dogs_one",
							Type:    "Table",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Duplicate Table Column Names in same Table",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name1",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name1",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
					{
						ID:   "name2",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name2",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "Name is already defined: name",
					Items: []ValidationItem{
						{
							Context: "Duplicate",
							ID:      "name2",
							Name:    "name",
							Table:   "dogs",
							Type:    "Column",
							Source:  "",
						},
						{
							Context: "Existing",
							ID:      "name1",
							Name:    "name",
							Table:   "dogs",
							Type:    "Column",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Duplicate Table Index Names in same Table",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name1",
						Name: "idx_name",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name1",
							Name:       "idx_name",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
					{
						ID:   "idx_name2",
						Name: "idx_name",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name2",
							Name:       "idx_name",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "Name is already defined: idx_name",
					Items: []ValidationItem{
						{
							Context: "Duplicate",
							ID:      "idx_name2",
							Name:    "idx_name",
							Table:   "dogs",
							Type:    "Index",
							Source:  "",
						},
						{
							Context: "Existing",
							ID:      "idx_name1",
							Name:    "idx_name",
							Table:   "dogs",
							Type:    "Index",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Duplicate Table Column PropertyIDs in same Table",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name_one",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name_one",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
					{
						ID:   "name",
						Name: "name_two",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name_two",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "PropertyID is already defined: name",
					Items: []ValidationItem{
						{
							Context: "Duplicate",
							ID:      "name",
							Name:    "name_two",
							Table:   "dogs",
							Type:    "Column",
							Source:  "",
						},
						{
							Context: "Existing",
							ID:      "name",
							Name:    "name_one",
							Table:   "dogs",
							Type:    "Column",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Duplicate Table Index PropertyIDs in same Table",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name",
						Name: "idx_name_one",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name_one",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
					{
						ID:   "idx_name",
						Name: "idx_name_two",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name_two",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "PropertyID is already defined: idx_name",
					Items: []ValidationItem{
						{
							Context: "Duplicate",
							ID:      "idx_name",
							Name:    "idx_name_two",
							Table:   "dogs",
							Type:    "Index",
							Source:  "",
						},
						{
							Context: "Existing",
							ID:      "idx_name",
							Name:    "idx_name_one",
							Table:   "dogs",
							Type:    "Index",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Duplicate Column PropertyIDs in different Tables",
		ExpectFail:  false,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name1",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name1",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
			{
				ID:   "cats",
				Name: "cats",
				Columns: []table.Column{
					{
						ID:   "name1",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name1",
							Name:       "name",
							Type:       "Column",
							ParentID:   "cats",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "cats",
					Name:       "cats",
					Type:       "Table",
				},
			},
		},
	},
	{
		Description: "Duplicate Index PropertyIDs in different Tables",
		ExpectFail:  false,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name",
						Name: "idx_name_one",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name_one",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
			{
				ID:   "cats",
				Name: "cats",
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name",
						Name: "idx_name_one",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name_one",
							Type:       "Index",
							ParentID:   "cats",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "cats",
					Name:       "cats",
					Type:       "Table",
				},
			},
		},
	},

	{
		Description: "Invalid PrimaryKey Name",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				PrimaryIndex: table.Index{
					ID:        "primarykey",
					Name:      "invalidpkname",
					IsPrimary: true,
					Columns: []table.IndexColumn{
						{
							Name: "name",
						},
					},
					Metadata: metadata.Metadata{
						PropertyID: "primarykey",
						Name:       "invalidpkname",
						Type:       "PrimaryKey",
						ParentID:   "dogs",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "INVALID_PK_NAME",
					Items: []ValidationItem{
						{
							Context: "INVALID_PK_NAME",
							ID:      "primarykey",
							Name:    "invalidpkname",
							Table:   "dogs",
							Type:    "PrimaryKey",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Invalid PrimaryKey PropertyID",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				PrimaryIndex: table.Index{
					ID:        "invalidpkname",
					Name:      "PrimaryKey",
					IsPrimary: true,
					Columns: []table.IndexColumn{
						{
							Name: "name",
						},
					},
					Metadata: metadata.Metadata{
						PropertyID: "invalidpkname",
						Name:       "PrimaryKey",
						Type:       "PrimaryKey",
						ParentID:   "dogs",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "INVALID_PK_ID",
					Items: []ValidationItem{
						{
							Context: "INVALID_PK_ID",
							ID:      "invalidpkname",
							Name:    "PrimaryKey",
							Table:   "dogs",
							Type:    "PrimaryKey",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "Invalid PrimaryKey PropertyID - Missing",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name_one",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name_one",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				PrimaryIndex: table.Index{
					ID:        "",
					Name:      "PrimaryKey",
					IsPrimary: true,
					Columns: []table.IndexColumn{
						{
							Name: "name",
						},
					},
					Metadata: metadata.Metadata{
						PropertyID: "",
						Name:       "PrimaryKey",
						Type:       "PrimaryKey",
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "INVALID_PK_ID",
					Items: []ValidationItem{
						{
							Context: "INVALID_PK_ID",
							Name:    "PrimaryKey",
							Table:   "dogs",
							Type:    "PrimaryKey",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "YAML Column Property ID Changed",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name_changed",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name_changed",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		MySQLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Columns: []table.Column{
					{
						ID:   "name",
						Name: "name",
						Metadata: metadata.Metadata{
							PropertyID: "name",
							Name:       "name",
							Type:       "Column",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "YAML PropertyID change detected. MySQL ID: [name]",
					Items: []ValidationItem{
						{
							Context: "CHANGED_ID",
							ID:      "name_changed",
							Name:    "name",
							Table:   "dogs",
							Type:    "Column",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "YAML Table Property ID Changed",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs_changed",
				Name: "dogs",
				Metadata: metadata.Metadata{
					PropertyID: "dogs_changed",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		MySQLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "YAML PropertyID change detected. MySQL ID: [dogs]",
					Items: []ValidationItem{
						{
							Context: "CHANGED_ID",
							ID:      "dogs_changed",
							Name:    "dogs",
							Table:   "dogs",
							Type:    "Table",
							Source:  "",
						},
					},
				},
			},
		},
	},

	{
		Description: "YAML Index Property ID Changed",
		ExpectFail:  true,
		YAMLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name_changed",
						Name: "idx_name",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name_changed",
							Name:       "idx_name",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		MySQLSchema: []table.Table{
			{
				ID:   "dogs",
				Name: "dogs",
				SecondaryIndexes: []table.Index{
					{
						ID:   "idx_name",
						Name: "idx_name",
						Columns: []table.IndexColumn{
							{
								Name: "name",
							},
						},
						Metadata: metadata.Metadata{
							PropertyID: "idx_name",
							Name:       "idx_name",
							Type:       "Index",
							ParentID:   "dogs",
						},
					},
				},
				Metadata: metadata.Metadata{
					PropertyID: "dogs",
					Name:       "dogs",
					Type:       "Table",
				},
			},
		},
		Problems: ValidationErrors{
			Errors: []ValidationError{
				{
					Desc: "YAML PropertyID change detected. MySQL ID: [idx_name]",
					Items: []ValidationItem{
						{
							Context: "CHANGED_ID",
							ID:      "idx_name_changed",
							Name:    "idx_name",
							Table:   "dogs",
							Type:    "Index",
							Source:  "",
						},
					},
				},
			},
		},
	},
}

func TestValidation(t *testing.T) {
	var problems ValidationErrors

	for i, test := range validationTests {

		if len(test.YAMLSchema) > 0 && len(test.MySQLSchema) == 0 {
			problems, _ = ValidateSchema(test.YAMLSchema, "YAML Schema", false)

			if problems.Count() > 0 {
				if test.ExpectFail {
					if !reflect.DeepEqual(problems, test.Problems) {
						t.Errorf("%d %s FAILED.", i, test.Description)
						util.LogWarnf("%s FAILED.", test.Description)
						util.DebugDumpDiff(problems, test.Problems)
					}

				} else {
					t.Errorf("%s FAILED. Unexpected problems.", test.Description)
					util.DebugDump(problems)
				}
			} else if test.ExpectFail {
				t.Errorf("%s FAILED.  Expected problems, didn't find any.", test.Description)
			}
		}

		if len(test.MySQLSchema) > 0 && len(test.YAMLSchema) == 0 {
			problems, _ = ValidateSchema(test.MySQLSchema, "Target Database Schema", false)

			if problems.Count() > 0 {
				if test.ExpectFail {
					if !reflect.DeepEqual(problems, test.Problems) {
						t.Errorf("%d %s FAILED.", i, test.Description)
						util.LogWarnf("%s FAILED.", test.Description)
						util.DebugDumpDiff(problems, test.Problems)
					}

				} else {
					t.Errorf("%s FAILED. Unexpected problems.", test.Description)
					util.DebugDump(problems)
				}
			} else if test.ExpectFail {
				t.Errorf("%s FAILED.  Expected problems, didn't find any.", test.Description)
			}
		}

		if len(test.YAMLSchema) > 0 && len(test.MySQLSchema) > 0 {
			problems, _ = ValidatePropertyIDs(test.YAMLSchema, test.MySQLSchema, false)

			if problems.Count() > 0 {
				if test.ExpectFail {
					if !reflect.DeepEqual(problems, test.Problems) {
						t.Errorf("%d %s FAILED.", i, test.Description)
						util.LogWarnf("%s FAILED.", test.Description)
						util.DebugDumpDiff(problems, test.Problems)
					}

				} else {
					t.Errorf("%s FAILED. Unexpected problems.", test.Description)
					util.DebugDump(problems)
				}
			} else if test.ExpectFail {
				t.Errorf("%s FAILED.  Expected problems, didn't find any.", test.Description)
			}
		}

	}
}
