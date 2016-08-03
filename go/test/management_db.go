package test

import (
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type ManagementDB struct {
	MockDB
}

func CreateManagementDB(context string, t *testing.T) (p ManagementDB, err error) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("%s: Setup Project DB Failed with Error: %v", context, err)
		return p, err
	}
	p = ManagementDB{MockDB{db, mock, "management"}}

	return p, nil
}

// Metadata Helpers

var metadataColumns = []string{
	"mdid",
	"db",
	"property_id",
	"parent_id",
	"type",
	"name",
	"exists",
}

func (m *ManagementDB) MetadataSelectName(name string, result DBRow) {
	query := DBQueryMock{
		Columns: metadataColumns,
		Rows: []DBRow{
			result,
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\"", name)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MetadataSelectNameParent(name string, parentId string, result DBRow) {
	query := DBQueryMock{
		Columns: metadataColumns,
		Rows: []DBRow{
			result,
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", name, parentId)

	m.ExpectQuery(query)
}

// Migration Helpers

var migrationColumns = []string{
	"mid",
	"db",
	"project",
	"version",
	"version_timestamp",
	"version_description",
	"status",
}

var migrationStepsColumns = []string{
	"sid",
	"mid",
	"op",
	"mdid",
	"name",
	"forward",
	"backward",
	"output",
	"status",
}

var migrationValuesTemplate = " values (null,?,?,?,?,?,?)"
var migrationStepsValuesTemplate = " values (null,?,?,?,?,?,?,?,?)"

func (m *ManagementDB) MigrationCount(result DBRow) {

	query := DBQueryMock{
		Type:    QueryCmd,
		Query:   "select count(*) from migration",
		Columns: []string{"count"},
		Rows:    []DBRow{result},
	}

	m.ExpectQuery(query)
}

func (m *ManagementDB) MigrationInsert(args DBRow) {

	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}
	query.FormatQuery("insert into `migration` (`%s`)%s", strings.Join(migrationColumns, "`,`"), migrationValuesTemplate)
	query.SetArgs(args...)

	m.ExpectExec(query)
}

func (m *ManagementDB) MigrationInsertStep(args DBRow) {

	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}
	query.FormatQuery("insert into `migration_steps` (`%s`)%s", strings.Join(migrationStepsColumns, "`,`"), migrationStepsValuesTemplate)
	query.SetArgs(args...)

	m.ExpectExec(query)
}
