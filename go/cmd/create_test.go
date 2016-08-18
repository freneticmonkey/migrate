package cmd

import (
	"fmt"
	"testing"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/urfave/cli"
)

func TestCreateFailNoProject(t *testing.T) {
	testName := "TestCreateFailNoProject"

	util.LogAlert(testName)

	// Test Configuration
	testConfig := test.GetTestConfig()

	project := ""
	version := ""
	rollback := false

	testConfig.Project.Name = project

	result := create(project, version, rollback, testConfig)

	if result.ExitCode() < 1 {
		t.Errorf("%s succeeded when it should have failed.", testName)
		return
	}
}

func TestCreateFailNoProjectVersion(t *testing.T) {
	testName := "TestCreateFailNoProjectVersion"

	util.LogAlert(testName)

	// Test Configuration
	testConfig := test.GetTestConfig()

	project := "UnitTestProject"
	version := ""
	rollback := false

	testConfig.Project.Name = project
	testConfig.Project.Schema.Version = version

	result := create(project, version, rollback, testConfig)

	if result.ExitCode() < 1 {
		t.Errorf("%s succeeded when it should have failed.", testName)
		return
	}
}

func TestCreate(t *testing.T) {
	testName := "TestCreate"
	util.LogAlert(testName)

	var err error
	var result *cli.ExitError
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	project := "UnitTestProject"
	version := "abc123"
	rollback := false

	// Configure testing data
	testConfig := test.GetTestConfig()
	dogsTbl := GetTableDogs()
	dogsAddTbl := GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	expectedAddressMetadata := dogsAddTbl.Columns[1].Metadata
	expectedAddressMetadata.MDID = 4

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

	shell.ExpectExecWithLambda(
		"git",
		params,
		"",
		nil,
		func(cmd string, args []string) error {

			////////////////////////////////////////////////////////
			// Configure source YAML files for the 'checked out' Schema
			//

			test.WriteFile(
				"UnitTestProject/dogs.yml",
				GetYAMLTableAddressDogs(),
				0644,
				false,
			)

			//
			////////////////////////////////////////////////////////
			return nil
		},
	)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure MySQL db reads for MySQL Schema
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
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata(dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsTbl.Metadata.ToDBRow(),
			dogsTbl.Columns[0].Metadata.ToDBRow(),
			dogsTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// STARTING Diff Queries - forwards

	mgmtDB.MetadataSelectName(
		dogsAddTbl.Name,
		dogsAddTbl.Metadata.ToDBRow(),
		false,
	)

	//Diff will also sync metadata for the YAML Schema
	mgmtDB.MetadataLoadAllTableMetadata(dogsAddTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsAddTbl.Metadata.ToDBRow(),
			dogsAddTbl.Columns[0].Metadata.ToDBRow(),
			dogsAddTbl.Columns[1].Metadata.ToDBRow(),
			dogsAddTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// Expect an insert for Metadata for the new column
	mgmtDB.MetadataInsert(
		test.DBRow{
			dogsAddTbl.Columns[1].Metadata.DB,
			dogsAddTbl.Columns[1].Metadata.PropertyID,
			dogsAddTbl.Columns[1].Metadata.ParentID,
			dogsAddTbl.Columns[1].Metadata.Type,
			dogsAddTbl.Columns[1].Metadata.Name,
			false,
		},
		expectedAddressMetadata.MDID,
		1,
	)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Git requests to pull back state of current checkout

	// GitVersionTime
	gitTime := "2016-07-12T22:04:05+10:00"
	mysqlTime := "2016-07-12 12:04:05"

	params = []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--format=%%cI",
	}
	shell.ExpectExec("git", params, gitTime, nil)

	// GitVersionDetails
	gitDetails := `commit abc123
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 12 22:04:05 2016 +1000

    An example git commit for unit testing`

	params = []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--pretty=medium",
	}
	shell.ExpectExec("git", params, gitDetails, nil)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// New migration being inserted into the database

	// Helper operation
	forward := mysql.SQLOperation{
		Statement: "ALTER TABLE `unittestproject_dogs` COLUMN `address` varchar(128) NOT NULL;",
		Op:        table.Add,
		Name:      "address",
		Metadata: metadata.Metadata{
			MDID:       4,
			DB:         1,
			PropertyID: "unittestproject_dogs_col_address",
			ParentID:   "unittestproject_dogs",
			Name:       "address",
		},
	}
	backwardsStatement := "ALTER TABLE `unittestproject_dogs` DROP COLUMN `address`;"

	// Pulling table metadata - diff backwards
	mgmtDB.MetadataSelectName(
		dogsAddTbl.Name,
		dogsAddTbl.Metadata.ToDBRow(),
		false,
	)

	//Diff will also sync metadata for the YAML Schema
	mgmtDB.MetadataLoadAllTableMetadata(dogsAddTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsAddTbl.Metadata.ToDBRow(),
			dogsAddTbl.Columns[0].Metadata.ToDBRow(),
			dogsAddTbl.Columns[1].Metadata.ToDBRow(),
			dogsAddTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// Counting the Migrations
	mgmtDB.MigrationCount(test.DBRow{0}, false)

	// Inserting a new Migration
	mgmtDB.MigrationInsert(
		test.DBRow{
			1,
			testConfig.Project.Name,
			testConfig.Project.Schema.Version,
			mysqlTime,
			gitDetails,
			0,
		},
		1,
		1,
	)

	// Inserting the Migration Step
	mgmtDB.MigrationInsertStep(
		test.DBRow{
			1,
			forward.Op,
			forward.Metadata.MDID,
			forward.Name,
			forward.Statement,
			backwardsStatement,
			"",
			0,
		},
		1,
		1,
	)

	//
	////////////////////////////////////////////////////////

	result = create(project, version, rollback, testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

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

func TestCreateRollback(t *testing.T) {
	testName := "TestCreateRollback"
	util.LogAlert(testName)

	var err error
	var result *cli.ExitError
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	project := "UnitTestProject"
	version := "abc123"
	rollback := true

	// Configure testing data
	testConfig := test.GetTestConfig()
	dogsTbl := GetTableDogs()
	dogsAddTbl := GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	expectedAddressMetadata := dogsAddTbl.Columns[1].Metadata
	expectedAddressMetadata.MDID = 4

	// GitVersionDetails
	gitDetails := `commit abc123
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 12 22:04:05 2016 +1000

    An example git commit for unit testing`

	//
	////////////////////////////////////////////////////////////////////

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

	shell.ExpectExecWithLambda(
		"git",
		params,
		"",
		nil,
		func(cmd string, args []string) error {

			////////////////////////////////////////////////////////
			// Configure source YAML files for the 'checked out' Schema
			//

			test.WriteFile(
				"UnitTestProject/dogs.yml",
				GetYAMLTableAddressDogs(),
				0644,
				false,
			)

			//
			////////////////////////////////////////////////////////
			return nil
		},
	)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure MySQL db reads for MySQL Schema
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
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata(dogsTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsTbl.Metadata.ToDBRow(),
			dogsTbl.Columns[0].Metadata.ToDBRow(),
			dogsTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// STARTING Diff Queries - forwards

	mgmtDB.MetadataSelectName(
		dogsAddTbl.Name,
		dogsAddTbl.Metadata.ToDBRow(),
		false,
	)

	//Diff will also sync metadata for the YAML Schema
	mgmtDB.MetadataLoadAllTableMetadata(dogsAddTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsAddTbl.Metadata.ToDBRow(),
			dogsAddTbl.Columns[0].Metadata.ToDBRow(),
			dogsAddTbl.Columns[1].Metadata.ToDBRow(),
			dogsAddTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// Expect an insert for Metadata for the new column
	mgmtDB.MetadataInsert(
		test.DBRow{
			dogsAddTbl.Columns[1].Metadata.DB,
			dogsAddTbl.Columns[1].Metadata.PropertyID,
			dogsAddTbl.Columns[1].Metadata.ParentID,
			dogsAddTbl.Columns[1].Metadata.Type,
			dogsAddTbl.Columns[1].Metadata.Name,
			false,
		},
		expectedAddressMetadata.MDID,
		1,
	)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Git requests to pull back state of current checkout

	// GitVersionTime
	gitTime := "2016-07-12T22:04:05+10:00"
	mysqlTime := "2016-07-12 12:04:05"

	params = []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--format=%%cI",
	}
	shell.ExpectExec("git", params, gitTime, nil)

	// GitVersionDetails

	params = []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--pretty=medium",
	}
	shell.ExpectExec("git", params, gitDetails, nil)

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// New migration being inserted into the database

	// Helper operation
	forward := mysql.SQLOperation{
		Statement: "ALTER TABLE `unittestproject_dogs` COLUMN `address` varchar(128) NOT NULL;",
		Op:        table.Add,
		Name:      "address",
		Metadata: metadata.Metadata{
			MDID:       4,
			DB:         1,
			PropertyID: "unittestproject_dogs_col_address",
			ParentID:   "unittestproject_dogs",
			Name:       "address",
		},
	}
	backwardsStatement := "ALTER TABLE `unittestproject_dogs` DROP COLUMN `address`;"

	// Pulling table metadata - diff backwards
	mgmtDB.MetadataSelectName(
		dogsAddTbl.Name,
		dogsAddTbl.Metadata.ToDBRow(),
		false,
	)

	//Diff will also sync metadata for the YAML Schema
	mgmtDB.MetadataLoadAllTableMetadata(dogsAddTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsAddTbl.Metadata.ToDBRow(),
			dogsAddTbl.Columns[0].Metadata.ToDBRow(),
			dogsAddTbl.Columns[1].Metadata.ToDBRow(),
			dogsAddTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// Counting the Migrations - expecting an existing migration
	mgmtDB.MigrationCount(
		test.DBRow{1},
		false,
	)

	// Checking for version - Doesn't exist
	mgmtDB.MigrationGetVersionExists(
		"abc123",
		test.DBRow{},
		true,
	)

	// Inserting a new Migration
	mgmtDB.MigrationInsert(
		test.DBRow{
			1,
			testConfig.Project.Name,
			testConfig.Project.Schema.Version,
			mysqlTime,
			gitDetails,
			0,
		},
		1,
		1,
	)

	// Inserting the Migration Step
	mgmtDB.MigrationInsertStep(
		test.DBRow{
			1,
			forward.Op,
			forward.Metadata.MDID,
			forward.Name,
			forward.Statement,
			backwardsStatement,
			"",
			0,
		},
		1,
		1,
	)

	//
	////////////////////////////////////////////////////////

	result = create(project, version, rollback, testConfig)

	if result.ExitCode() > 0 {
		t.Errorf("%s failed with error: %v", testName, result)
		return
	}

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
