package cmd

import (
	"testing"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/test"
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

	// Teardown() - Pre test cleanup
	Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// No project or version for this test
	project := ""
	version := ""

	// Mock MySQL

	// Mock Table structs - with the new Address Column
	dogsTbl := GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema read
	//

	test.WriteFile(
		"UnitTestProject/dogs.yml",
		GetYAMLTableDogs(),
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
	projectDB.ShowCreateTable(dogsTbl.Name, GetMySQLCreateTableDogs())

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		test.GetDBRowMetadata(dogsTbl.Metadata),
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata(dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			test.GetDBRowMetadata(dogsTbl.Metadata),
			test.GetDBRowMetadata(dogsTbl.Columns[0].Metadata),
			test.GetDBRowMetadata(dogsTbl.PrimaryIndex.Metadata),
		},
		false,
	)

	// STARTING Diff Queries

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		test.GetDBRowMetadata(dogsTbl.Metadata),
		false,
	)

	//Diff will also sync metadata for the YAML Schema
	mgmtDB.MetadataLoadAllTableMetadata(dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			test.GetDBRowMetadata(dogsTbl.Metadata),
			test.GetDBRowMetadata(dogsTbl.Columns[0].Metadata),
			test.GetDBRowMetadata(dogsTbl.PrimaryIndex.Metadata),
		},
		false,
	)

	//
	//
	////////////////////////////////////////////////////////

	// Execute the schema diff
	result = diff(project, version, testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

	Teardown()
}
