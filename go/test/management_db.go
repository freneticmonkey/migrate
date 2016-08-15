package test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"

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

func (m *ManagementDB) DatabaseCreateTable() {

	ct := []string{
		"CREATE TABLE `target_database` (",
		" `dbid` int(11) NOT NULL AUTO_INCREMENT,",
		" `project` varchar(255) DEFAULT NULL,",
		" `name` varchar(255) DEFAULT NULL,",
		" `env` varchar(255) DEFAULT NULL,",
		" PRIMARY KEY (`dbid`) ",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}

	ctStr := strings.Join(ct, "")
	ctStr = regexp.QuoteMeta(ctStr)
	m.Mock.ExpectExec(ctStr).WillReturnResult(sqlmock.NewResult(0, 0))
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

var metadataValuesTemplate = " values (null,?,?,?,?,?,?)"

func (m *ManagementDB) MetadataGet(mdid int, result DBRow, expectEmtpy bool) {

	query := DBQueryMock{
		Columns: metadataColumns,
	}
	if !expectEmtpy {
		query.Rows = []DBRow{result}
	}
	query.FormatQuery("SELECT * FROM `metadata` WHERE mdid=%d", mdid)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MetadataInsert(args DBRow, lastInsert int64, rowsAffected int64) {

	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(lastInsert, rowsAffected),
	}
	query.FormatQuery("insert into `metadata` (`%s`)%s", strings.Join(metadataColumns, "`,`"), metadataValuesTemplate)
	query.SetArgs(args...)

	m.ExpectExec(query)
}

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

func (m *ManagementDB) MetadataCreateTable() {

	ct := []string{
		"CREATE TABLE `metadata` (",
		" `mdid` bigint(20) NOT NULL AUTO_INCREMENT,",
		" `db` int(11) DEFAULT NULL,",
		" `property_id` varchar(255) DEFAULT NULL,",
		" `parent_id` varchar(255) DEFAULT NULL,",
		" `type` varchar(255) DEFAULT NULL,",
		" `name` varchar(255) DEFAULT NULL,",
		" `exists` tinyint(1) DEFAULT NULL,",
		" PRIMARY KEY (`mdid`)",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}

	ctStr := strings.Join(ct, "")
	ctStr = regexp.QuoteMeta(ctStr)
	m.Mock.ExpectExec(ctStr).WillReturnResult(sqlmock.NewResult(0, 0))
}

func GetDBRowMetadata(m metadata.Metadata) DBRow {
	return DBRow{
		m.MDID,
		m.DB,
		m.PropertyID,
		m.ParentID,
		m.Type,
		m.Name,
		m.Exists,
	}
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
	"timestamp",
}

var migrationValuesTemplate = " values (null,?,?,?,?,?,?,?)"

func (m *ManagementDB) MigrationGet(mid int64, result DBRow, expectEmpty bool) {
	query := DBQueryMock{
		Columns: migrationColumns,
	}
	if !expectEmpty {
		query.Rows = append(query.Rows, result)
	}
	query.FormatQuery("SELECT * FROM `migration` WHERE mid=%d", mid)

	m.ExpectQuery(query)
}

func (m *ManagementDB) MigrationGetLatest(result DBRow, expectEmpty bool) {
	query := DBQueryMock{
		Columns: migrationColumns,
	}
	if !expectEmpty {
		query.Rows = append(query.Rows, result)
	}
	query.FormatQuery("select * from migration ORDER BY version_timestamp DESC LIMIT 1")

	m.ExpectQuery(query)
}

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

	queryStr := fmt.Sprintf(
		"insert into `migration` (`%s`) values (null,?,?,?,?,?,?)",
		strings.Join(migrationColumns[:len(migrationColumns)-1], "`,`"),
	)
	query := DBQueryMock{
		Type:   ExecCmd,
		Result: sqlmock.NewResult(lastInsert, rowsAffected),
	}
	query.FormatQuery(queryStr)
	query.SetArgs(args...)

	m.ExpectExec(query)
}

func (m *ManagementDB) MigrationCreateTable() {

	ct := []string{
		"CREATE TABLE `migration` (",
		" `mid` bigint(20) NOT NULL AUTO_INCREMENT,",
		" `db` int(11) NOT NULL,",
		" `project` varchar(255) NOT NULL,",
		" `version` varchar(255) NOT NULL,",
		" `version_timestamp` datetime NOT NULL,",
		" `version_description` text,",
		" `status` int(11) NOT NULL,",
		" `timestamp` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,",
		" PRIMARY KEY (`mid`) ",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}

	ctStr := strings.Join(ct, "")
	ctStr = regexp.QuoteMeta(ctStr)
	m.Mock.ExpectExec(ctStr).WillReturnResult(sqlmock.NewResult(0, 0))

}

func GetDBRowMigration(m migration.Migration) DBRow {
	return DBRow{
		m.MID,
		m.DB,
		m.Project,
		m.Version,
		m.VersionTimestamp,
		m.VersionDescription,
		m.Status,
		m.Timestamp,
	}
}

// Migration Steps Helpers

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

var migrationStepsValuesTemplate = " values (null,?,?,?,?,?,?,?,?)"

func (m *ManagementDB) MigrationStepGet(mid int64, result DBRow, expectEmpty bool) {
	query := DBQueryMock{
		Columns: migrationStepsColumns,
	}
	if !expectEmpty {
		query.Rows = append(query.Rows, result)
	}
	query.FormatQuery("SELECT * FROM `migration_steps` WHERE mid=%d", mid)

	m.ExpectQuery(query)
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
	query.FormatQuery("update migration_steps WHERE mid = %d SET status = %d", sid, status)

	m.ExpectExec(query)
}

func (m *ManagementDB) MigrationStepCreateTable() {

	ct := []string{
		"CREATE TABLE `migration_steps` (",
		" `sid` bigint(20) NOT NULL AUTO_INCREMENT,",
		" `mid` bigint(20) DEFAULT NULL,",
		" `op` int(11) DEFAULT NULL,",
		" `mdid` bigint(20) DEFAULT NULL,",
		" `name` varchar(255) DEFAULT NULL,",
		" `forward` varchar(255) DEFAULT NULL,",
		" `backward` varchar(255) DEFAULT NULL,",
		" `output` text,",
		" `status` int(11) DEFAULT NULL,",
		" PRIMARY KEY (`sid`) ",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}

	ctStr := strings.Join(ct, "")
	ctStr = regexp.QuoteMeta(ctStr)
	m.Mock.ExpectExec(ctStr).WillReturnResult(sqlmock.NewResult(0, 0))

}

func GetDBRowMigrationStep(s migration.Step) DBRow {
	return DBRow{
		s.SID,
		s.MID,
		s.Op,
		s.MDID,
		s.Name,
		s.Forward,
		s.Backward,
		s.Output,
		s.Status,
	}
}
