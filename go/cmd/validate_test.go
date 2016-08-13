package cmd

import (
	"fmt"
	"testing"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/urfave/cli"
)

func TestValidate(t *testing.T) {
	var err error
	var result *cli.ExitError
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	// Test Configuration
	testConfig := test.GetTestConfig()

	// Teardown() - Pre test cleanup
	util.SetConfigTesting()
	util.Config(testConfig)

	testName := "TestValidate"

	// No project or version for this test
	project := ""
	version := ""

	// Configure testing data
	dogsTbl := GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema validation
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
	// Configure MySQL db reads for Schema validation
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
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	}

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, GetMySQLCreateTableDogs())

	// ParseCreateTable which includes a Table.LoadDBMetadata() call
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

	//
	//
	////////////////////////////////////////////////////////

	result = validate(project, version, "both", testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("TestValidate failed with error: %v", err)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

	Teardown()

}

func TestValidateYAML(t *testing.T) {
	var err error
	var result *cli.ExitError

	// Test Configuration
	testConfig := test.GetTestConfig()

	// Teardown() - Pre test cleanup
	util.SetConfigTesting()
	util.Config(testConfig)

	testName := "TestValidate"

	// No project or version for this test
	project := ""
	version := ""

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema validation
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

	result = validate(project, version, "yaml", testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("%s failed with error: %v", testName, err)
		return
	}

	Teardown()

}

func TestValidateMySQL(t *testing.T) {
	var err error
	var result *cli.ExitError
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	// Test Configuration
	testConfig := test.GetTestConfig()

	// Teardown() - Pre test cleanup
	util.SetConfigTesting()
	util.Config(testConfig)

	testName := "TestValidate"

	// No project or version for this test
	project := ""
	version := ""

	// Configure testing data
	dogsTbl := GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure MySQL db reads for Schema validation
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
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	}

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, GetMySQLCreateTableDogs())

	// ParseCreateTable which includes a Table.LoadDBMetadata() call
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

	//
	//
	////////////////////////////////////////////////////////

	result = validate(project, version, "mysql", testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("TestValidate failed with error: %v", err)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

	Teardown()

}

func TestGitCloneValidate(t *testing.T) {
	var err error
	var result *cli.ExitError
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	// Teardown()

	testName := "TestValidate"

	// No project or version for this test
	project := "UnitTestProject"
	version := "abc123"

	// Configure testing data
	testConfig := test.GetTestConfig()
	dogsTbl := GetTableAddressDogs()

	////////////////////////////////////////////////////////
	// Configure Git Checkout

	var sparseExists bool
	var data []byte

	expectedSparseFile := `schema/*
schemaTwo/*`

	// Configure unit test shell
	util.Config(testConfig)

	checkoutPath := util.WorkingSubDir(testConfig.Project.Name)

	// Build Git Commands

	shell := util.GetShell().(*util.MockShellExecutor)

	// init
	params := []string{
		"-C",
		checkoutPath,
		"init",
	}
	shell.ExpectExec("git", params, "", nil)

	// git add remote
	params = []string{
		"-C",
		checkoutPath,
		"remote",
		"add",
		"-f",
		"origin",
		testConfig.Project.Schema.Url,
	}
	shell.ExpectExec("git", params, "", nil)

	// git config sparse checkout
	params = []string{
		"-C",
		checkoutPath,
		"config",
		"core.sparseCheckout",
		"true",
	}
	shell.ExpectExec("git", params, "", nil)

	// Checkout
	// git config sparse checkout
	params = []string{
		"-C",
		checkoutPath,
		"checkout",
		testConfig.Project.Schema.Version,
	}
	shell.ExpectExec("git", params, "", nil)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema validation
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
	// Configure MySQL db reads for Schema validation
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
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	}

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}}, false)

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, GetMySQLCreateTableDogs())

	// ParseCreateTable which includes a Table.LoadDBMetadata() call
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

	//
	//
	////////////////////////////////////////////////////////

	result = validate(project, version, "both", testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("TestValidate failed with error: %v", result.Error())
		return
	}

	////////////////////////////////////////////////////////
	// DB was accessed correctly
	projectDB.ExpectionsMet(testName, t)
	mgmtDB.ExpectionsMet(testName, t)

	////////////////////////////////////////////////////////
	// Git was checked out correctly

	if err != nil {
		t.Errorf("%s FAILED with error: %v", testName, err)
		return
	}

	// Validate that the sparse checkout file was created correctly
	sparseFile := fmt.Sprintf("%s/.git/info/sparse-checkout", checkoutPath)
	sparseExists, err = util.FileExists(sparseFile)

	if err != nil {
		t.Errorf("%s FAILED: there was an error while reading the sparse checkut file at path: [%s] with error: [%v]", testName, sparseFile, err)
		return
	}

	if !sparseExists {
		t.Errorf("%s FAILED: sparse checkut file missing from path: [%s]", testName, sparseFile)
		return
	}

	data, err = util.ReadFile(sparseFile)
	sparseData := string(data)

	if sparseData != expectedSparseFile {
		t.Errorf("%s FAILED: sparse data file contents: [%s] doesn't match expected contents: [%s]", testName, sparseData, expectedSparseFile)
		return
	}

	// Ensure that all of the anticipated shell calls were made
	if err = shell.ExpectationsWereMet(); err != nil {
		t.Errorf("%s FAILED: Not all shell commands were executed: error [%v]", testName, err)
	}

	Teardown()
}