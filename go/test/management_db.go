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

func (m *ManagementDB) ShowTables(results []DBRow, expectEmpty bool) {

	query := DBQueryMock{
		Query:   "SHOW TABLES IN management",
		Columns: []string{"table"},
	}

	if !expectEmpty {
		query.Rows = results
	}
	m.ExpectQuery(query)
}

// Database Helpers

var databaseColumns = []string{
	"dbid",
	"project",
	"name",
	"env",
}

func (m *ManagementDB) DatabaseGet(project string, name string, env string, result DBRow, expectEmtpy bool) {

	query := DBQueryMock{
		Columns: databaseColumns,
	}
	if !expectEmtpy {
		query.Rows = []DBRow{result}
	}
	query.FormatQuery("SELECT * FROM target_database WHERE project=\"%s\" AND name=\"%s\" AND env=\"%s\"", project, name, env)

	m.ExpectQuery(query)
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

func (m *ManagementDB) MetadataGet(mid int, result DBRow, expectEmtpy bool) {

	query := DBQueryMock{
		Columns: metadataColumns,
	}
	if !expectEmtpy {
		query.Rows = []DBRow{result}
	}
	query.FormatQuery("SELECT * FROM `metadata` WHERE mdid=%d", mid)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MetadataInsert(args DBRow, lastInsert int64, rowsAffected int64) {

	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(lastInsert, rowsAffected),
	}
	query.FormatQuery("insert into `metadata` (`%s`)%s", strings.Join(metadataColumns, "`,`"), migrationValuesTemplate)
	query.SetArgs(args...)

	m.ExpectExec(query)
}

//'select * from migration WHERE status = 6'

//'select `mdid`,`db`,`property_id`,`parent_id`,`type`,`name`,`exists` from `metadata` where `mdid`=?;' with args [1] was not expected]

func (m *ManagementDB) MetadataSelectName(name string, result DBRow, expectEmpty bool) {
	query := DBQueryMock{
		Columns: metadataColumns,
	}
	if !expectEmpty {
		query.Rows = append(query.Rows, result)
	}

	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\"", name)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MetadataSelectNameParent(name string, parentId string, result DBRow, expectEmpty bool) {
	query := DBQueryMock{
		Columns: metadataColumns,
	}
	if !expectEmpty {
		query.Rows = append(query.Rows, result)
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", name, parentId)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MetadataLoadAllTableMetadata(tblPropertyID string, dbID int64, results []DBRow, expectEmpty bool) {
	query := DBQueryMock{
		Columns: metadataColumns,
	}
	if !expectEmpty {
		query.Rows = results
	}
	query.FormatQuery("select * from metadata WHERE name = \"%s\" OR parent_id = \"%s\" AND db=%d", tblPropertyID, tblPropertyID, dbID)

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

func (m *ManagementDB) MigrationCount(result DBRow, expectEmpty bool) {

	query := DBQueryMock{
		Type:    QueryCmd,
		Query:   "select count(*) from migration",
		Columns: []string{"count"},
	}
	if !expectEmpty {
		query.Rows = []DBRow{result}
	}

	m.ExpectQuery(query)
}

func (m *ManagementDB) MigrationGetStatus(status int, results []DBRow, expectEmpty bool) {

	query := DBQueryMock{
		Columns: migrationColumns,
	}
	if !expectEmpty {
		query.Rows = results
	}
	query.FormatQuery("select * from migration WHERE status = %d", status)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MigrationSetStatus(mid int64, status int) {

	query := DBQueryMock{
		Columns: migrationColumns,
		Result:  sqlmock.NewResult(0, 1),
	}
	query.FormatQuery("update migration WHERE mid = %d SET status = %d", mid, status)

	m.ExpectExec(query)
}

func (m *ManagementDB) MigrationInsert(args DBRow, lastInsert int64, rowsAffected int64) {

	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(lastInsert, rowsAffected),
	}
	query.FormatQuery("insert into `migration` (`%s`)%s", strings.Join(migrationColumns, "`,`"), migrationValuesTemplate)
	query.SetArgs(args...)

	m.ExpectExec(query)
}

func (m *ManagementDB) MigrationInsertStep(args DBRow, lastInsert int64, rowsAffected int64) {

	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(lastInsert, rowsAffected),
	}
	query.FormatQuery("insert into `migration_steps` (`%s`)%s", strings.Join(migrationStepsColumns, "`,`"), migrationStepsValuesTemplate)
	query.SetArgs(args...)

	m.ExpectExec(query)
}

func (m *ManagementDB) StepSetStatus(sid int64, status int) {

	query := DBQueryMock{
		Columns: migrationStepsColumns,
		Result:  sqlmock.NewResult(0, 1),
	}
	query.FormatQuery("update migration_steps WHERE sid = %d SET status = %d", sid, status)

	m.ExpectExec(query)
}
