package cmd

import (
	"testing"

	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/test"
)

func TestManagementSetup(t *testing.T) {
	var mgmtDB test.ManagementDB
	var err error

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

func TestBuildSchema(t *testing.T) {
	var mgmtDB test.ManagementDB
	var err error

	testName := "TestSetupManagementDB"

	// Configuration
	testConfig := GetTestConfig()

	// Setup the mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	// Configure the Queries

	// If we have none of the tables
	mgmtDB.ShowTables(
		[]test.DBRow{},
		true,
	)

	// Build tables will check again, if we have none of the tables
	mgmtDB.ShowTables(
		[]test.DBRow{},
		true,
	)

	// create if not exists metadata
	mgmtDB.MetadataCreateTable()

	// create if not exists migration
	mgmtDB.MigrationCreateTable()

	// create if not exists migration step
	mgmtDB.MigrationStepCreateTable()

	// create if not exists target_database
	mgmtDB.DatabaseCreateTable()

	// Set the management DB
	management.SetManagementDB(mgmtDB.Db)

	// Configure the management DB
	err = management.Setup(testConfig)

	if err != nil {

		// Build the Management Tables
		err = management.BuildSchema(testConfig)

		if err != nil {
			t.Errorf("%s FAILED with err: %v", testName, err)
		}

	} else {
		t.Errorf("%s FAILED because configuration was successful and management tables are being detected", testName)
	}

	mgmtDB.ExpectionsMet(testName, t)

}

func TestSetupExistingDB(t *testing.T) {
	var mgmtDB test.ManagementDB
	var projectDB test.ProjectDB
	var err error

	testName := "TestSetupManagementDB"

	// Configuration
	testConfig := GetTestConfig()

	dogsTbl := GetTableDogs()

	// Setup the mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	// Setup the mock Project DB

	projectDB, err = test.CreateProjectDB(testName, t)

	mysql.Setup(testConfig)

	// Configure metadata
	metadata.Setup(mgmtDB.Db, 1)

	// Connect to Project DB
	mysql.SetProjectDB(projectDB.Db.Db)

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, GetMySQLCreateTableDogs())

	// Load Table Metadata - Expect empty because this is a new database
	mgmtDB.MetadataSelectName(
		"dogs",
		test.DBRow{},
		// test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
		true,
	)

	// mgmtDB.MetadataLoadAllTableMetadata("tbl1",
	// 	1,
	// 	[]test.DBRow{
	// 		test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
	// 		test.DBRow{2, 1, "col1", "tbl1", "Column", "id", 1},
	// 		test.DBRow{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
	// 	},
	// 	false,
	// )

	// metadata insert
	mgmtDB.MetadataInsert(
		test.DBRow{1, "tbl1", "", "Table", "dogs", 1},
		1,
		1,
	)

	// metadata insert
	mgmtDB.MetadataInsert(
		test.DBRow{1, "col1", "tbl1", "Column", "id", 1},
		2,
		1,
	)

	// metadata insert
	mgmtDB.MetadataInsert(
		test.DBRow{1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
		3,
		1,
	)

	if err != nil {
		t.Errorf("%s FAILED with err: %v", testName, err)
	}
}