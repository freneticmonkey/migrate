package test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var MgmtDB *ManagementDB

// Mock DB operation types
const (
	// Exec Type
	ExecCmd = iota
	// Query Type
	QueryCmd
)

// DBRow Helper type for defining DB Rows
type DBRow []driver.Value

// DBQueryMock Helper struct for configuring a Mock DB request and return
type DBQueryMock struct {
	Type    int
	Query   string
	Args    []driver.Value
	Columns []string
	Rows    []DBRow
	Result  driver.Result
}

// FormatQuery Build the Query from a format
func (dbq *DBQueryMock) FormatQuery(query string, args ...interface{}) {
	dbq.Query = fmt.Sprintf(query, args...)
}

// SetArgs Set the arguments for the query
func (dbq *DBQueryMock) SetArgs(args ...driver.Value) {
	dbq.Args = args
}

type MockDB struct {
	Db   *gorp.DbMap
	Mock sqlmock.Sqlmock
	Name string
}

func (m *MockDB) ExpectExec(query DBQueryMock) {

	m.Mock.ExpectExec(regexp.QuoteMeta(query.Query)).
		WithArgs(query.Args...).
		WillReturnResult(query.Result)
}

func (m *MockDB) ExpectQuery(query DBQueryMock) {

	rows := sqlmock.NewRows(query.Columns)
	for _, r := range query.Rows {
		rows.AddRow(r...)
	}
	m.Mock.ExpectQuery(regexp.QuoteMeta(query.Query)).WillReturnRows(rows)
}

func (m *MockDB) ExpectionsMet(context string, t *testing.T) {
	if err := m.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("%s: %s DB queries failed expectations. Error: %s", context, m.Name, err)
	}
}

func (m *MockDB) CreateDatabase() {

	// Configure expected project database refresh queries
	query := DBQueryMock{
		Result: sqlmock.NewResult(0, 0),
	}
	query.FormatQuery("CREATE DATABASE `%s`", m.Name)

	m.ExpectExec(query)
}

func (m *MockDB) DropDatabase() {

	// Configure expected project database refresh queries
	query := DBQueryMock{
		Result: sqlmock.NewResult(0, 0),
	}
	query.FormatQuery("DROP DATABASE `%s`", m.Name)

	m.ExpectExec(query)
}

func (m *MockDB) Close() {
	m.Mock.ExpectClose()
}

// createMockDB Configure Gorp with Mock DB
func createMockDB() (gdb *gorp.DbMap, mock sqlmock.Sqlmock, err error) {
	var mockDb *sql.DB

	mockDb, mock, err = sqlmock.New()

	if err != nil {
		return nil, mock, err
	}

	gdb = &gorp.DbMap{
		Db: mockDb,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		},
	}

	return gdb, mock, err
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
