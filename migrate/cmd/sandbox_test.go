package cmd

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/management"
	"github.com/freneticmonkey/migrate/migrate/util"
)

var mockDb *sql.DB
var mock sqlmock.Sqlmock

// Configure Gorp with Mock DB
func dbSetup() (gdb *gorp.DbMap, err error) {

	mockDb, mock, err = sqlmock.New()

	if err != nil {
		return nil, err
	}

	gdb = &gorp.DbMap{
		Db: mockDb,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		},
	}

	return gdb, err
}

func dbTearDown() {
	mockDb.Close()
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

func TestRefreshDatabase(t *testing.T) {
	util.SetVerbose(true)
	// Test Configuration

	testConfig := config.Config{
		Project: config.Project{
			Name: "animals",
			DB: config.DB{
				Username:    "root",
				Password:    "test",
				Ip:          "127.0.0.1",
				Port:        3500,
				Database:    "test",
				Environment: "SANDBOX",
			},
		},
	}

	// Configure the management gorp DB

	mdb, err := dbSetup()

	if err != nil {
		t.Errorf("Test Refresh Database: Setup Management DB Failed with Error: %v", err)
	} else {
		management.SetManagementDB(mdb)
	}

	// TODO: Add mock db expect table 'database' accesses here.

	mock.ExpectQuery("SHOW TABLES IN management").
		WillReturnRows(sqlmock.NewRows([]string{
			"tables",
		}).
			AddRow("metadata").
			AddRow("migration").
			AddRow("migration_steps").
			AddRow("target_database"))

	query := fmt.Sprintf("SELECT * FROM target_database WHERE project=\"%s\" AND name=\"%s\" AND env=\"%s\"",
		testConfig.Project.Name,
		testConfig.Project.DB.Database,
		testConfig.Project.DB.Environment,
	)
	query = regexp.QuoteMeta(query)

	mock.ExpectQuery(query).
		WillReturnRows(sqlmock.NewRows([]string{
			"dbid",
			"project",
			"name",
			"env",
		}).AddRow(
			1,
			testConfig.Project.Name,
			testConfig.Project.DB.Database,
			testConfig.Project.DB.Environment,
		))

	// Configure management
	setConfig(testConfig)

	// Reset/recreate management database?

	// TODO: Add mock db expect database recreation statements here.

	// Setup expected database operations

	query = "select count(*) from migration"
	query = regexp.QuoteMeta(query)
	mock.ExpectQuery(query).
		WillReturnRows(sqlmock.NewRows([]string{
			"count",
		}).AddRow(
			0,
		))

	// Execute the migration
	err = sandboxProcessFlags(testConfig, true, false, false, true)

	if err != nil {
		t.Errorf("Sandbox Refresh FAILED with Error: %v", err)
	}

	// Verify database operations
}

func TestNewTableApplyImmediately(t *testing.T) {

}

func TestNewColumnApplyImmediately(t *testing.T) {

}
