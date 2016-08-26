package cmd

import (
	"testing"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/testdata"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/urfave/cli"
)

func TestDiff(t *testing.T) {
	testName := "TestDiff"

	util.LogAlert(testName)

	var err error
	var result *cli.ExitError

	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// No project or version for this test
	project := ""
	version := ""
	tableName := ""

	// Mock MySQL

	// Mock Table structs - with the new Address Column
	dogsTbl := testdata.GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema read
	//

	test.WriteFile(
		"UnitTestProject/dogs.yml",
		testdata.GetYAMLTableDogs(),
		0644,
		false,
	)

	//
	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure MySQL db reads for Schema read
	//

	// Configure the test databases
	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
		mysql.Setup(testConfig)

		// Connect to Project DB
		mysql.SetProjectDB(projectDB.Db.Db)
	} else {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	} else {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	// Expect some requests to determine the MySQL schema

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, testdata.GetMySQLCreateTableDogs())

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata(
		dogsTbl.Name,
		dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsTbl.Metadata.ToDBRow(),
			dogsTbl.Columns[0].Metadata.ToDBRow(),
			dogsTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// STARTING Diff Queries

	//
	//
	////////////////////////////////////////////////////////

	// Execute the schema diff
	result = diff(project, version, tableName, testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}

func TestDiffTableName(t *testing.T) {
	testName := "TestDiffTableName"

	util.LogAlert(testName)

	var err error
	var result *cli.ExitError

	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// No project or version for this test
	project := ""
	version := ""
	tableName := "dogs"

	// Mock MySQL

	// Mock Table structs - with the new Address Column
	dogsTbl := testdata.GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema read
	//

	test.WriteFile(
		"UnitTestProject/dogs.yml",
		testdata.GetYAMLTableDogs(),
		0644,
		false,
	)

	//
	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure MySQL db reads for Schema read
	//

	// Configure the test databases
	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
		mysql.Setup(testConfig)

		// Connect to Project DB
		mysql.SetProjectDB(projectDB.Db.Db)
	} else {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	} else {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	// Expect some requests to determine the MySQL schema

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, testdata.GetMySQLCreateTableDogs())

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata(
		dogsTbl.Name,
		dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsTbl.Metadata.ToDBRow(),
			dogsTbl.Columns[0].Metadata.ToDBRow(),
			dogsTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// STARTING Diff Queries

	//
	//
	////////////////////////////////////////////////////////

	// Execute the schema diff
	result = diff(project, version, tableName, testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}

func TestDiffTableNameFailed(t *testing.T) {
	testName := "TestDiffTableNameFailed"

	util.LogAlert(testName)

	var err error
	var result *cli.ExitError

	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// No project or version for this test
	project := ""
	version := ""
	tableName := "cats"

	// Mock MySQL

	// Mock Table structs - with the new Address Column
	dogsTbl := testdata.GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema read
	//

	test.WriteFile(
		"UnitTestProject/dogs.yml",
		testdata.GetYAMLTableDogs(),
		0644,
		false,
	)

	//
	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure MySQL db reads for Schema read
	//

	// Configure the test databases
	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
		mysql.Setup(testConfig)

		// Connect to Project DB
		mysql.SetProjectDB(projectDB.Db.Db)
	} else {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	} else {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	// Expect some requests to determine the MySQL schema

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, testdata.GetMySQLCreateTableDogs())

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata(
		dogsTbl.Name,
		dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsTbl.Metadata.ToDBRow(),
			dogsTbl.Columns[0].Metadata.ToDBRow(),
			dogsTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// STARTING Diff Queries

	//
	//
	////////////////////////////////////////////////////////

	// Execute the schema diff
	result = diff(project, version, tableName, testConfig)

	if result.ExitCode() == 0 {
		t.Errorf("%s SHOULD HAVE FAILED but didn't", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}
