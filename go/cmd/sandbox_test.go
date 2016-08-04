package cmd

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
)

func TestDiffSchema(t *testing.T) {

	var err error
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	var forwards mysql.SQLOperations
	var backwards mysql.SQLOperations

	// Test Configuration
	testConfig := GetTestConfig()

	// Mock MySQL
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"  `id` int(11) NOT NULL,",
		"  PRIMARY KEY (`id`),",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1",
	}
	var dogsTableStr = strings.Join(dogsTable, "\n")

	// Mock Table structs
	dogsTbl := table.Table{
		Name:    "dogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					PropertyID: "col1",
					ParentID:   "tbl1",
					Name:       "id",
					Type:       "Column",
				},
			},
			{
				Name: "address",
				Type: "varchar",
				Size: []int{128},
				Metadata: metadata.Metadata{
					PropertyID: "col2",
					ParentID:   "tbl1",
					Name:       "address",
					Type:       "Column",
				},
			},
		},
		PrimaryIndex: table.Index{
			IsPrimary: true,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				PropertyID: "pi",
				ParentID:   "tbl1",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Metadata: metadata.Metadata{
			PropertyID: "tbl1",
			Name:       "dogs",
			Type:       "Table",
		},
	}

	expectedForwards := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: "ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;",
			Op:        table.Add,
			Name:      "address",
			Metadata: metadata.Metadata{
				MDID:       4,
				DB:         1,
				PropertyID: "col2",
				ParentID:   "tbl1",
				Name:       "address",
				Type:       "Column",
			},
		},
	}

	expectedBackwards := mysql.SQLOperations{
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
				Type:       "Column",
			},
		},
	}

	// Push Dogs table into YAML Schema for Diffing
	yaml.Schema = append(yaml.Schema, dogsTbl)

	// Setup the mock project database
	projectDB, _ = test.CreateProjectDB("TestDiffSchema", t)

	// Configure the MySQL Read Tables queries

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}})

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, dogsTableStr)

	// Configure the Mock Managment DB

	// Setup the mock Managment DB
	mgmtDB, _ = test.CreateManagementDB("TestDiffSchema", t)

	mysql.Setup(testConfig)

	mgmtDB.MetadataSelectName(
		"dogs",
		test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata("tbl1",
		1,
		[]test.DBRow{
			test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
			test.DBRow{2, 1, "col1", "tbl1", "Column", "id", 1},
			test.DBRow{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
		},
		false,
	)

	mgmtDB.MetadataSelectName(
		"dogs",
		test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
		false,
	)

	// Diff will also sync metadata for the YAML Schema
	mgmtDB.MetadataLoadAllTableMetadata("tbl1",
		1,
		[]test.DBRow{
			test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
			test.DBRow{2, 1, "col1", "tbl1", "Column", "id", 1},
			test.DBRow{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
		},
		false,
	)

	// Expect an insert for Metadata for the new column
	mgmtDB.MetadataInsert(
		test.DBRow{1, "col2", "tbl1", "Column", "address", false},
		4,
		1,
	)

	mgmtDB.MetadataSelectName(
		"dogs",
		test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
		false,
	)

	mgmtDB.MetadataLoadAllTableMetadata("tbl1",
		1,
		[]test.DBRow{
			test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
			test.DBRow{2, 1, "col1", "tbl1", "Column", "id", 1},
			test.DBRow{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
		},
		false,
	)

	// Configure metadata
	metadata.Setup(mgmtDB.Db, 1)

	// Connect to Project DB
	mysql.SetProjectDB(projectDB.Db.Db)

	// Execute the schema diff
	forwards, backwards, err = diffSchema(testConfig, "Unit Test", false)

	if err != nil {
		t.Errorf("TestDiffSchema: Failed with Error: %v", err)
	}

	if !reflect.DeepEqual(expectedForwards, forwards) {
		util.DebugDumpDiff(expectedForwards, forwards)
		t.Error("TestDiffSchema: Forwards Operation differs from expected.")
	}

	if !reflect.DeepEqual(expectedBackwards, backwards) {
		util.DebugDumpDiff(expectedBackwards, backwards)
		t.Error("TestDiffSchema: Backwards Operation differs from expected.")
	}

	projectDB.ExpectionsMet("TestDiffSchema", t)

	mgmtDB.ExpectionsMet("TestDiffSchema", t)
	util.SetVerbose(false)

}

func TestRecreateProjectDatabase(t *testing.T) {
	var projectDB test.ProjectDB

	var err error

	// Test Configuration
	testConfig := GetTestConfig()

	// Setup the mock project database
	projectDB, err = test.CreateProjectDB("TestDiffSchema", t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
	}

	projectDB.DropDatabase()

	// Reuse the query object to define the create database
	projectDB.CreateDatabase()

	projectDB.Close()

	// Execute the recreation!
	recreateProjectDatabase(testConfig, false)

	projectDB.ExpectionsMet("Test Recreate Database", t)
}

func TestCreateMigration(t *testing.T) {

	var mgmtDb test.ManagementDB
	var err error

	var m migration.Migration

	// Configure the Mock Managment DB
	mgmtDb, err = test.CreateManagementDB("TestCreateMigration", t)
	if err == nil {
		migration.Setup(mgmtDb.Db, 1)
	}

	// Test Configuration
	testConfig := config.Config{
		Project: config.Project{
			Name: "UnitTestProject",
			Schema: config.Schema{
				Version: "abc123",
			},
			DB: config.DB{
				Database: "test",
			},
		},
	}

	mgmtDb.MigrationCount(test.DBRow{0}, false)

	mgmtDb.MigrationInsert(test.DBRow{1, "UnitTestProject", "abc123", mysql.GetTimeNow(), "unit test", 0}, 1, 1)

	mgmtDb.MigrationInsertStep(test.DBRow{1, 0, 4, "address", "ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;", "ALTER TABLE `dogs` DROP COLUMN `address`;", "", 0}, 1, 1)

	forwardOps := mysql.SQLOperations{
		mysql.SQLOperation{
			Statement: "ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;",
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

	backwardOps := mysql.SQLOperations{
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

	m, err = createMigration(testConfig, "unit test", false, forwardOps, backwardOps)

	if err != nil {
		t.Errorf("TestMigrateSandbox Failed. Error: %v", err)
	}

	if m.MID == 0 {
		t.Errorf("TestMigrateSandbox Failed. There was a problem inserting Migration into the DB.  Final Migration malformed")
	}

	mgmtDb.ExpectionsMet("TestRecreateProjectDatabase", t)
}

func TestMigrateSandbox(t *testing.T) {
	var projectDB test.ProjectDB
	var mgmtDb test.ManagementDB
	var err error

	// Test Configuration
	testConfig := GetTestConfig()

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
				Forward:  "ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;",
				Backward: "ALTER TABLE `dogs` DROP COLUMN `address`;",
				Output:   "",
				Status:   migration.Unapproved,
			},
		},
		Sandbox: true,
	}

	// Setup the mock project database
	projectDB, err = test.CreateProjectDB("TestMigrateSandbox", t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
	}

	// Configure the Mock Managment DB
	mgmtDb, err = test.CreateManagementDB("TestMigrateSandbox", t)

	if err == nil {
		// migration.Setup(mgmtDb.Db, 1)
		exec.Setup(mgmtDb.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDb.Db, 1)
		metadata.Setup(mgmtDb.Db, 1)
	}

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
		test.DBRow{1, 1, "col1", "tbl1", "Column", "address", 1},
		false,
	)

	// Set Step to InProgress
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		1,
		table.Add,
		1,
		"address",
		"ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;",
		"ALTER TABLE `dogs` DROP COLUMN `address`;",
		"",
		migration.InProgress,
		1,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	query := test.DBQueryMock{
		Type:   test.ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}
	query.FormatQuery("ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;")

	projectDB.ExpectExec(query)

	// Set Step to Forced
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		1,
		table.Add,
		1,
		"address",
		"ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;",
		"ALTER TABLE `dogs` DROP COLUMN `address`;",
		"Row(s) Affected: 1",
		migration.Forced,
		1,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Metadata
	mgmtDb.MetadataGet(
		1,
		test.DBRow{1, 1, "col1", "tbl1", "Column", "address", 1},
		false,
	)

	// Update Migrationt with completed
	mgmtDb.Mock.ExpectExec("update `migration`").WithArgs(
		1,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		mysql.GetTimeNow(),
		"Testing a Migration",
		migration.Forced,
		1,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the MigrationStep with completed
	mgmtDb.Mock.ExpectExec("update `migration_steps`").WithArgs(
		1,
		table.Add,
		1,
		"address",
		"ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;",
		"ALTER TABLE `dogs` DROP COLUMN `address`;",
		"Row(s) Affected: 1",
		migration.Forced,
		1,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrateSandbox("TestMigrateSandbox", false, &m)

	if err != nil {
		t.Errorf("TestMigrateSandbox Failed. There was a problem executing the Migration.")
	}

	mgmtDb.ExpectionsMet("TestMigrateSandbox", t)
	projectDB.ExpectionsMet("TestMigrateSandbox", t)

}

func TestRefreshDatabase(t *testing.T) {

}

func TestNewTableApplyImmediately(t *testing.T) {

}

func TestNewColumnApplyImmediately(t *testing.T) {

}
