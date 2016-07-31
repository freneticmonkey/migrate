package test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// Mock DB operation types
const (
	// Exec Type
	ExecCmd = iota
	// Query Type
	QueryCmd
)

// DBQueryMock Helper struct for configuring a Mock DB request and return
type DBQueryMock struct {
	Type    int
	Query   string
	Args    []interface{}
	Columns []string
	Rows    [][]driver.Value
	Result  driver.Result
}

// SetArgs Set the arguments for the query
func (dbq *DBQueryMock) SetArgs(args ...interface{}) {
	dbq.Args = args
}

// CreateMockDB Configure Gorp with Mock DB
func CreateMockDB() (gdb *gorp.DbMap, mock sqlmock.Sqlmock, err error) {
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

// ExpectDB Helper function for configuring expected DB calls and the request results
func ExpectDB(mockDb sqlmock.Sqlmock, query DBQueryMock) {
	var builtQuery string
	builtQuery = regexp.QuoteMeta(fmt.Sprintf(query.Query, query.Args...))

	switch query.Type {
	case ExecCmd:
		mockDb.ExpectExec(builtQuery).
			WithArgs().
			WillReturnResult(query.Result)
	case QueryCmd:

		rows := sqlmock.NewRows(query.Columns)
		for _, r := range query.Rows {
			rows.AddRow(r...)
		}

		mockDb.ExpectQuery(builtQuery).WillReturnRows(rows)
	}
}
