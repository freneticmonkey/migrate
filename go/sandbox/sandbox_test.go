package sandbox

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/testdata"
	"github.com/freneticmonkey/migrate/go/util"
)

func setupRecreateDBSchema(projectDB *test.ProjectDB, result []test.DBRow, tables []string) {

	projectDB.ShowTables(result, false)

	// Update the MigrationStep with completed
	projectDB.Mock.ExpectExec(
		fmt.Sprintf("DROP TABLE `%s`", strings.Join(tables, "`,`"))).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func TestDiffSchema(t *testing.T) {
	util.LogAlert("TestDiffSchema")

	var forwards mysql.SQLOperations
	var backwards mysql.SQLOperations
	var err error

	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	testName := "TestDiffSchema"

	// Test Configuration
	testConfig := test.GetTestConfig()
	util.SetConfigTesting()
	util.Config(testConfig)

	// Create a YAML file for the diff to read
	test.WriteFile(
		"unittestproject/dogs.yml",
		testdata.GetYAMLTableAddressDogs(),
		0644,
		false,
	)

	// Mock MySQL

	// Mock Table structs - with the new Address Column
	dogsTbl := testdata.GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	expectedAddressMetadata := dogsTbl.Columns[1].Metadata
	expectedAddressMetadata.MDID = 4

	expectedForwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: "ALTER TABLE `dogs` ADD COLUMN `address` varchar(128) NOT NULL;",
			Op:        table.Add,
			Name:      "address",
			Metadata:  expectedAddressMetadata,
		},
	}

	expectedBackwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: "ALTER TABLE `dogs` DROP COLUMN `address`;",
			Op:        table.Del,
			Name:      "address",
			Metadata:  expectedAddressMetadata,
		},
	}

	// Setup the mock project database
	projectDB, _ = test.CreateProjectDB(testName+"", t)

	// Configure the Mock Managment DB

	// Setup the mock Managment DB
	mgmtDB, _ = test.CreateManagementDB(testName, t)

	mysql.Setup(testConfig)

	// Configure metadata
	metadata.Setup(mgmtDB.Db, 1)

	// Connect to Project DB
	mysql.SetProjectDB(projectDB.Db.Db)

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

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	// Diff will also sync metadata for the YAML Schema
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

	// Expect an insert for Metadata for the new column
	mgmtDB.MetadataInsert(
		test.DBRow{
			dogsTbl.Columns[1].Metadata.DB,
			dogsTbl.Columns[1].Metadata.PropertyID,
			dogsTbl.Columns[1].Metadata.ParentID,
			dogsTbl.Columns[1].Metadata.Type,
			dogsTbl.Columns[1].Metadata.Name,
			false,
		},
		expectedAddressMetadata.MDID,
		1,
	)

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

	// Execute the schema diff
	forwards, backwards, err = diffSchema(testConfig, "Unit Test", false)

	if err != nil {
		t.Errorf(testName+": Failed with Error: %v", err)
	}

	if !reflect.DeepEqual(expectedForwards, forwards) {
		util.DebugDumpDiff(expectedForwards, forwards)
		t.Error(testName + ": Forwards Operation differs from expected.")
	}

	if !reflect.DeepEqual(expectedBackwards, backwards) {
		util.DebugDumpDiff(expectedBackwards, backwards)
		t.Error(testName + ": Backwards Operation differs from expected.")
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestCreateMigration(t *testing.T) {
	util.LogAlert("TestCreateMigration")
	var mgmtDb test.ManagementDB
	var err error
	var m migration.Migration

	testConfig := test.GetTestConfig()
	testName := "TestCreateMigration"

	util.SetConfigTesting()
	util.Config(testConfig)

	// Create a YAML file for the diff to read
	test.WriteFile(
		"unittestproject/dogs.yml",
		testdata.GetYAMLTableAddressDogs(),
		0644,
		false,
	)

	forwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: "ALTER TABLE `dogs` ADD COLUMN `address` varchar(128) NOT NULL;",
			Op:        table.Add,
			Name:      "address",
			Metadata: metadata.Metadata{
				MDID:       4,
				DB:         1,
				PropertyID: "col2",
				ParentID:   "tbl1",
				Name:       "address",
			},
		},
	}

	backwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: "ALTER TABLE `dogs` DROP COLUMN `address`;",
			Op:        table.Del,
			Name:      "address",
			Metadata: metadata.Metadata{
				MDID:       4,
				DB:         1,
				PropertyID: "col2",
				ParentID:   "tbl1",
				Name:       "address",
			},
		},
	}

	// Configure the Mock Managment DB
	mgmtDb, err = test.CreateManagementDB(testName, t)
	if err == nil {
		migration.Setup(mgmtDb.Db, 1)
	}

	// Counting the Migrations
	mgmtDb.MigrationCount(test.DBRow{0}, false)

	// Inserting a new Migration
	mgmtDb.MigrationInsert(
		test.DBRow{
			1,
			testConfig.Project.Name,
			testConfig.Project.Schema.Version,
			mysql.GetTimeNow(),
			testName,
			0,
		},
		1,
		1,
	)

	// Inserting the Migration Step
	forward := forwards[0]
	mgmtDb.MigrationInsertStep(
		test.DBRow{
			1,
			forward.Op,
			forward.Metadata.MDID,
			forward.Name,
			forward.Statement,
			backwards[0].Statement,
			"",
			0,
		},
		1,
		1,
	)

	// Execute the migration
	m, err = createMigration(testConfig, testName, false, forwards, backwards)

	if err != nil {
		t.Errorf(testName+" Failed. Error: %v", err)
	}

	if m.MID == 0 {
		t.Errorf(testName + " Failed. There was a problem inserting Migration into the DB.  Final Migration malformed")
	}

	// Validate the DB access
	mgmtDb.ExpectionsMet(testName, t)

	testdata.Teardown()
}

func TestRecreateProjectDatabase(t *testing.T) {
	util.LogAlert("TestRecreateProjectDatabase")
	var err error

	testName := "TestRecreateProjectDatabase"

	var projectDB test.ProjectDB
	var mgmtDb test.ManagementDB
	// Test Configuration
	testConfig := test.GetTestConfig()

	util.SetConfigTesting()
	util.Config(testConfig)

	// Create a YAML file for the diff to read
	test.WriteFile(
		"unittestproject/dogs.yml",
		testdata.GetYAMLTableAddressDogs(),
		0644,
		false,
	)

	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
		mysql.SetProjectDB(projectDB.Db.Db)
	}

	// Configure the Mock Managment DB
	mgmtDb, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDb.Db, 1)
		exec.Setup(mgmtDb.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDb.Db, 1)
		metadata.Setup(mgmtDb.Db, 1)
	}

	// Expect all tables to be dropped
	setupRecreateDBSchema(&projectDB, []test.DBRow{{"dogs"}}, []string{"dogs"})

	// Expect all metadata to be deleted
	mgmtDb.Mock.ExpectExec("DELETE FROM metadata WHERE db = 1").WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the recreation!
	recreateProjectDatabase(testConfig, false)

	projectDB.ExpectionsMet(testName, t)

	testdata.Teardown()

}

func TestMigrateSandbox(t *testing.T) {
	util.LogAlert("TestMigrateSandbox")
	var err error

	// Test Configuration
	testConfig := test.GetTestConfig()

	testName := "TestMigrateSandbox"

	dogsAddTbl := testdata.GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	colMd := dogsAddTbl.Columns[1].Metadata
	colMd.MDID = 4

	m := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   mysql.GetTimeNow(),
		VersionDescription: "Testing a Migration",
		Status:             migration.Unapproved,
		Timestamp:          mysql.GetTimeNow(),
		Steps: []migration.Step{
			{
				SID:      1,
				MID:      1,
				Op:       table.Add,
				MDID:     1,
				Name:     "address",
				Forward:  "ALTER TABLE `dogs` ADD COLUMN `address` varchar(128) NOT NULL;",
				Backward: "ALTER TABLE `dogs` DROP COLUMN `address`;",
				Output:   "",
				Status:   migration.Unapproved,
			},
		},
		Sandbox: true,
	}

	var projectDB test.ProjectDB
	var mgmtDb test.ManagementDB

	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
	}

	// Configure the Mock Managment DB
	mgmtDb, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDb.Db, 1)
		exec.Setup(mgmtDb.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDb.Db, 1)
		metadata.Setup(mgmtDb.Db, 1)
	}

	// Grab the step and use it to populate the expected database mock queries
	step := m.Steps[0]

	// Check for running migrations
	mgmtDb.MigrationGetStatus(
		migration.InProgress,
		[]test.DBRow{
			{},
		},
		true,
	)

	// Set this migration to running
	mgmtDb.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	// Set Step to InProgress
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		step.Output,
		migration.InProgress,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	query := test.DBQueryMock{
		Type:   test.ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}
	query.FormatQuery("ALTER TABLE `dogs` ADD COLUMN `address` varchar(128) NOT NULL;")

	projectDB.ExpectExec(query)

	// Set Step to Forced
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Forced,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Metadata
	mgmtDb.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	// Update Metadata with completed
	mgmtDb.Mock.ExpectExec("update `metadata`").WithArgs(
		colMd.DB,
		colMd.PropertyID,
		colMd.ParentID,
		colMd.Type,
		colMd.Name,
		true,
		colMd.MDID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Migrationt with completed
	mgmtDb.Mock.ExpectExec("update `migration`").WithArgs(
		m.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		m.VersionTimestamp,
		m.VersionDescription,
		migration.Forced,
		m.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the MigrationStep with completed
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Forced,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrateSandbox(testName, false, &m)

	if err != nil {
		t.Errorf("%s Failed. There was a problem executing the Migration. Error: %v", testName, err)
	}

	mgmtDb.ExpectionsMet(testName, t)
	projectDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}

func TestRefreshDatabase(t *testing.T) {
	var projectDB test.ProjectDB
	var mgmtDb test.ManagementDB
	var err error

	util.LogAlert("Starting Refresh Database")

	testName := "TestRefreshDatabase"

	// Test Configuration
	testConfig := test.GetTestConfig()

	util.SetConfigTesting()
	util.Config(testConfig)

	// Configure the Test Datadata

	dogsTbl := testdata.GetTableDogs()
	dogsAddressTbl := testdata.GetTableAddressDogs()

	// Create a YAML file for the diff to read
	test.WriteFile(
		"unittestproject/dogs.yml",
		testdata.GetYAMLTableDogs(),
		0644,
		false,
	)

	// The recreation Migration
	m := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   mysql.GetTimeNow(),
		VersionDescription: testName,
		Status:             migration.Unapproved,
		Timestamp:          mysql.GetTimeNow(),
		Steps: []migration.Step{
			{
				SID:      1,
				MID:      1,
				Op:       table.Add,
				MDID:     1,
				Name:     "dogs",
				Forward:  testdata.GetCreateTableDogs(),
				Backward: "DROP TABLE `dogs`;",
				Output:   "",
				Status:   migration.Unapproved,
			},
		},
		Sandbox: true,
	}

	step := m.Steps[0]
	forwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: step.Forward,
			Op:        step.Op,
			Name:      step.Name,
			Metadata: metadata.Metadata{
				MDID:       step.MDID,
				DB:         m.DB,
				PropertyID: "table_dogs",
				ParentID:   "",
				Name:       step.Name,
				Type:       "Table",
			},
		},
	}

	backwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: step.Backward,
			Op:        table.Del,
			Name:      step.Name,
			Metadata: metadata.Metadata{
				MDID:       step.MDID,
				DB:         m.DB,
				PropertyID: "table_dogs",
				ParentID:   "",
				Name:       step.Name,
				Type:       "Table",
			},
		},
	}

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
	mgmtDb, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDb.Db, 1)
		exec.Setup(mgmtDb.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDb.Db, 1)
		metadata.Setup(mgmtDb.Db, 1)
	}

	// Configure Schema access

	// Wipe the Project DB
	setupRecreateDBSchema(&projectDB, []test.DBRow{{"dogs"}}, []string{"dogs"})

	// SHOW TABLES Query - Expecting it to be empty
	projectDB.ShowTables([]test.DBRow{{}}, true)

	// There won't be any searching for MySQL Table Metadata

	// DiffSchema - Forwards

	// Sync Metadata

	mgmtDb.MetadataSelectName(
		dogsAddressTbl.Name,
		dogsAddressTbl.Metadata.ToDBRow(),
		false,
	)

	// Diff will also sync metadata for the YAML Schema
	mgmtDb.MetadataLoadAllTableMetadata(
		dogsAddressTbl.Name,
		dogsAddressTbl.Metadata.PropertyID,
		1,
		[]test.DBRow{
			dogsAddressTbl.Metadata.ToDBRow(),
			dogsAddressTbl.Columns[0].Metadata.ToDBRow(),
			dogsAddressTbl.Columns[1].Metadata.ToDBRow(),
			dogsAddressTbl.PrimaryIndex.Metadata.ToDBRow(),
		},
		false,
	)

	// DiffSchema - Backwards

	// Create Migration

	// Counting the Migrations
	mgmtDb.MigrationCount(test.DBRow{0}, false)

	// Inserting a new Migration
	mgmtDb.MigrationInsert(
		test.DBRow{
			1,
			testConfig.Project.Name,
			testConfig.Project.Schema.Version,
			mysql.GetTimeNow(),
			testName,
			0,
		},
		1,
		1,
	)

	// Inserting the Migration Step
	forward := forwards[0]
	mgmtDb.MigrationInsertStep(
		test.DBRow{
			1,
			forward.Op,
			forward.Metadata.MDID,
			forward.Name,
			forward.Statement,
			backwards[0].Statement,
			"",
			0,
		},
		1,
		1,
	)

	// Migrate Sandbox

	// Check for running migrations
	mgmtDb.MigrationGetStatus(
		migration.InProgress,
		[]test.DBRow{
			{},
		},
		true,
	)

	// Set this migration to running
	mgmtDb.MetadataGet(
		1,
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	// Set Step to InProgress
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		step.Output,
		migration.InProgress,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	query := test.DBQueryMock{
		Type:   test.ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}
	query.FormatQuery(testdata.GetCreateTableDogs())

	projectDB.ExpectExec(query)

	// Set Step to Forced
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Forced,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Load the Metadata for the Step
	mgmtDb.MetadataGet(
		1,
		dogsTbl.Metadata.ToDBRow(),
		false,
	)

	// Update Metadata to Exists
	md := dogsTbl.Metadata
	mgmtDb.Mock.ExpectExec("update `metadata`").WithArgs(
		1,
		md.PropertyID,
		md.ParentID,
		md.Type,
		md.Name,
		true,
		1,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Migration with completed
	mgmtDb.Mock.ExpectExec("update `migration`").WithArgs(
		m.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		m.VersionTimestamp,
		m.VersionDescription,
		migration.Forced,
		m.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the MigrationStep with completed
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Forced,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	Action(testConfig, false, true, "TestRefreshDatabase")

	mgmtDb.ExpectionsMet(testName, t)
	projectDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}

func TestNewTableApplyImmediately(t *testing.T) {
	util.LogAlert("TestNewTableApplyImmediately")
	testdata.Teardown()

}

func TestNewColumnApplyImmediately(t *testing.T) {
	util.LogAlert("TestNewColumnApplyImmediately")

	testdata.Teardown()
}
