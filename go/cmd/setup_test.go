package cmd

import (
	"testing"

	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
)

func TestManagementSetup(t *testing.T) {
	var mgmtDB test.ManagementDB
	var err error

	util.SetVerbose(true)

	testName := "TestSetupManagementDB"

	// Configuration
	testConfig := GetTestConfig()

	// Setup the mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	// Configure the Queries

	// If we have the tables
	mgmtDB.ShowTables(
		[]test.DBRow{
			{"metadata"},
			{"migration"},
			{"migration_steps"},
			{"target_database"},
		},
		false,
	)

	// And an entry for the SANDBOX database
	mgmtDB.DatabaseGet(
		testConfig.Project.Name,
		testConfig.Project.DB.Database,
		testConfig.Project.DB.Environment,
		test.DBRow{1, "UnitTestProject", "project", "SANDBOX"},
		false,
	)

	// Set the management DB
	management.SetManagementDB(mgmtDB.Db)

	// Configure the management DB
	err = management.Setup(testConfig)

	if err != nil {
		t.Errorf("%s FAILED with err: %v", testName, err)
	}

	mgmtDB.ExpectionsMet(testName, t)

}

// func TestBuildSchema(t *testing.T) {
// 	var mgmtDB test.ManagementDB
// 	var err error
//
// 	testName := "TestSetupManagementDB"
//
// 	// Configuration
// 	testConfig := GetTestConfig()
//
// 	// Setup the mock Managment DB
// 	mgmtDB, err = test.CreateManagementDB(testName, t)
//
// 	// Configure the Queries
//
// 	// If we have the tables
// 	mgmtDB.ShowTables(
// 		[]test.DBRow{},
// 		true,
// 	)
//
// 	// Set the management DB
// 	management.SetManagementDB(mgmtDB.Db)
//
// 	// Configure the management DB
// 	err = management.Setup(testConfig)
//
// 	if err != nil {
//
// 		// Build the Management Tables
// 		err = management.BuildSchema(testConfig)
//
// 		if err != nil {
// 			t.Errorf("%s FAILED with err: %v", testName, err)
// 		}
//
// 	} else {
// 		t.Errorf("%s FAILED because configuration was successfull", testName)
// 	}
//
// 	mgmtDB.ExpectionsMet(testName, t)
//
// }

func TestSetupExistingDB(t *testing.T) {

}
