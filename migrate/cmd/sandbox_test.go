package cmd

import (
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/exec"
	"github.com/freneticmonkey/migrate/migrate/test"
)

func TestDiffSchema(t *testing.T) {

}

func TestCreateMigration(t *testing.T) {

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
	pdb, projectMock, err = test.CreateMockDB()

	if err != nil {
		t.Errorf("Test Recreate Project Database: Setup Project DB Failed with Error: %v", err)
	} else {
		// Connect to Project DB
		exec.SetProjectDB(pdb)
	}

	// Configure expected project database refresh queries
	query := test.DBQueryMock{
		Type:   test.ExecCmd,
		Query:  "DROP DATABASE `%s`",
		Result: sqlmock.NewResult(0, 0),
	}
	query.SetArgs(testConfig.Project.DB.Database)

	test.ExpectDB(projectMock, query)

	// Reuse the query object to define the create database
	query.Query = "CREATE DATABASE `%s`"
	test.ExpectDB(projectMock, query)

	projectMock.ExpectClose()

	// Execute the recreation!
	recreateProjectDatabase(testConfig, false)

	if err = projectMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Test Recreate Database: Project DB queries failed expectations. Error: %s", err)
	}
}

func TestMigrateSandbox(t *testing.T) {

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
