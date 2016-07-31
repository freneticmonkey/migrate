package cmd

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/exec"
	"github.com/freneticmonkey/migrate/migrate/util"
)

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

const (
	ExecCmd = iota
	QueryCmd
)

type DBQueryMock struct {
	Type    int
	Query   string
	Args    []interface{}
	Columns []string
	Rows    [][]driver.Value
	Result  driver.Result
}

func (dbq *DBQueryMock) SetArgs(args ...interface{}) {
	dbq.Args = args
}

func expectDB(mockDb sqlmock.Sqlmock, query DBQueryMock) {
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

func DisableTestConfigReadURL(t *testing.T) {

	// TODO: Provide config
	var remoteConfig = `
    options:
        management:
            db:
                username: root
                password: test
                ip:       127.0.0.1
                port:     3400
                database: management

    # Project Definition
    project:
        # Project name - used to identify the project by the cli flags
        # and configure the table's namespace
        name: "animals"
        db:
            username:    root
            password:    test
            ip:          127.0.0.1
            port:        3500
            database:    test
            environment: UNITTEST
    `
	expectedConfig := config.Config{
		Options: config.Options{
			Management: config.Management{
				DB: config.DB{
					Username:    "root",
					Password:    "test",
					Ip:          "127.0.0.1",
					Port:        3400,
					Database:    "management",
					Environment: "",
				},
			},
		},
		Project: config.Project{
			Name: "animals",
			DB: config.DB{
				Username:    "root",
				Password:    "test",
				Ip:          "127.0.0.1",
				Port:        3500,
				Database:    "test",
				Environment: "UNITTEST",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, remoteConfig)
	}))
	defer ts.Close()

	urlConfig, err := loadConfig(ts.URL, "")

	if err != nil {
		t.Errorf("Config Read URL FAILED with Error: %v", err)
	}
	if !reflect.DeepEqual(expectedConfig, urlConfig) {
		t.Error("Config Read URL FAILED. Returned config does not match.")
		util.LogWarn("Config Read URL FAILED. Returned config does not match.")
		util.DebugDumpDiff(expectedConfig, urlConfig)
	}
}

func TestRecreateProjectDatabase(t *testing.T) {
	var pdb *gorp.DbMap
	var projectMock sqlmock.Sqlmock

	var err error

	// Test Configuration
	testConfig := config.Config{
		Project: config.Project{
			DB: config.DB{
				Database: "test",
			},
		},
	}

	// Setup the mock project database
	pdb, projectMock, err = createMockDB()

	if err != nil {
		t.Errorf("Test Recreate Project Database: Setup Project DB Failed with Error: %v", err)
	} else {
		// Connect to Project DB
		exec.SetProjectDB(pdb)
	}

	// Configure expected project database refresh queries
	query := DBQueryMock{
		Type:   ExecCmd,
		Query:  "DROP DATABASE `%s`",
		Result: sqlmock.NewResult(0, 0),
	}
	query.SetArgs(testConfig.Project.DB.Database)

	expectDB(projectMock, query)

	// Reuse the query object to define the create database
	query.Query = "CREATE DATABASE `%s`"
	expectDB(projectMock, query)

	// Execute the recreation!

	recreateProjectDatabase(&testConfig, false)

	if err = projectMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Test Recreate Database: Project DB queries failed expectations. Error: %s", err)
	}

}

func TestRefreshDatabase(t *testing.T) {

	// var mdb *gorp.DbMap
	// var mgmtMock sqlmock.Sqlmock
	//
	// var pdb *gorp.DbMap
	// var projectMock sqlmock.Sqlmock
	//
	// var err error
	//
	// // Useful for unit tests
	// util.SetVerbose(true)
	//
	// // Test Configuration
	// testConfig := config.Config{
	// 	Project: config.Project{
	// 		Name: "animals",
	// 		DB: config.DB{
	// 			Username:    "root",
	// 			Password:    "test",
	// 			Ip:          "127.0.0.1",
	// 			Port:        3500,
	// 			Database:    "test",
	// 			Environment: "SANDBOX",
	// 		},
	// 	},
	// }
	//
	// // Configure the management gorp DB
	// mdb, mgmtMock, err = createMockDB()
	//
	// if err != nil {
	// 	t.Errorf("Test Refresh Database: Setup Management DB Failed with Error: %v", err)
	// } else {
	// 	management.SetManagementDB(mdb)
	// }
	//
	// // Add mock db expect table 'database' accesses here.
	//
	// mgmtMock.ExpectQuery("SHOW TABLES IN management").
	// 	WillReturnRows(sqlmock.NewRows([]string{
	// 		"tables",
	// 	}).
	// 		AddRow("metadata").
	// 		AddRow("migration").
	// 		AddRow("migration_steps").
	// 		AddRow("target_database"))
	//
	// query := fmt.Sprintf("SELECT * FROM target_database WHERE project=\"%s\" AND name=\"%s\" AND env=\"%s\"",
	// 	testConfig.Project.Name,
	// 	testConfig.Project.DB.Database,
	// 	testConfig.Project.DB.Environment,
	// )
	// query = regexp.QuoteMeta(query)
	//
	// mgmtMock.ExpectQuery(query).
	// 	WillReturnRows(sqlmock.NewRows([]string{
	// 		"dbid",
	// 		"project",
	// 		"name",
	// 		"env",
	// 	}).AddRow(
	// 		1,
	// 		testConfig.Project.Name,
	// 		testConfig.Project.DB.Database,
	// 		testConfig.Project.DB.Environment,
	// 	))
	//
	// // Configure management
	// setConfig(testConfig)
	//
	// // Reset/recreate management database?
	//
	// // TODO: Add mock db expect management database recreation statements here.
	//
	// query = "select count(*) from migration"
	// query = regexp.QuoteMeta(query)
	//
	// mgmtMock.ExpectQuery(query).
	// 	WillReturnRows(sqlmock.NewRows([]string{
	// 		"count",
	// 	}).AddRow(
	// 		0,
	// 	))
	//
	// // Execute the migration
	// err = sandboxProcessFlags(testConfig, true, false, false, true)
	//
	// if err != nil {
	// 	t.Errorf("Sandbox Refresh FAILED with Error: %v", err)
	// }
	//
	// // Verify database operations
	//
	// if err = mgmtMock.ExpectationsWereMet(); err != nil {
	// 	t.Errorf("Refresh Database: Management DB access failed. Error: %s", err)
	// }
}

func TestNewTableApplyImmediately(t *testing.T) {

}

func TestNewColumnApplyImmediately(t *testing.T) {

}
