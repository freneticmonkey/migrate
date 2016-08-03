package cmd

import (
	"strings"
	"testing"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/yaml"
)

func TestDiffSchema(t *testing.T) {
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB
	var err error

	// Test Configuration
	testConfig := config.Config{
		Project: config.Project{
			DB: config.DB{
				Database: "project",
			},
			LocalSchema: config.LocalSchema{
				Path: "ignore",
			},
		},
	}

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
			},
		},
		Metadata: metadata.Metadata{
			PropertyID: "tbl1",
			Name:       "dogs",
		},
	}

	// Push Dogs table into YAML Schema for Diffing
	yaml.Schema = append(yaml.Schema, dogsTbl)

	// Setup the mock project database
	projectDB, err = test.CreateProjectDB("TestDiffSchema", t)

	// Configure the MySQL Read Tables queries

	// SHOW TABLES Query
	projectDB.ShowTables([]test.DBRow{{dogsTbl.Name}})

	// SHOW CREATE TABLE Query
	projectDB.ShowCreateTable(dogsTbl.Name, dogsTableStr)

	// Configure the Mock Managment DB

	// Setup the mock Managment DB
	mgmtDB, err = test.CreateManagementDB("TestDiffSchema", t)

	// mgmtDb, mgmtMock, err := test.CreateMockDB()

	if err != nil {
		t.Errorf("TestDiffSchema: Setup Project DB Failed with Error: %v", err)
	}

	query := test.DBQueryMock{
		Type:    test.QueryCmd,
		Columns: []string{"count"},
		Rows:    []test.DBRow{{1}},
	}
	query.FormatQuery("SELECT count(*) from metadata WHERE name=\"%s\" and type=\"Table\"", dogsTbl.Name)

	mgmtDB.ExpectQuery(query)

	// Search Metadata for `dogs` table query - MySQL
	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
	)

	mgmtDB.MetadataSelectNameParent(
		"id",
		"tbl1",
		test.DBRow{2, 1, "col1", "tbl1", "Column", "id", 1},
	)

	mgmtDB.MetadataSelectNameParent(
		"PrimaryKey",
		"tbl1",
		test.DBRow{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
	)

	mgmtDB.MetadataSelectName(
		dogsTbl.Name,
		test.DBRow{1, 1, "tbl1", "", "Table", "dogs", 1},
	)

	mgmtDB.MetadataSelectNameParent(
		"PrimaryKey",
		"tbl1",
		test.DBRow{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
	)

	mgmtDB.MetadataSelectNameParent(
		"id",
		"tbl1",
		test.DBRow{2, 1, "col1", "tbl1", "Column", "id", 1},
	)

	// Configure metadata
	metadata.Setup(mgmtDB.Db, 1)

	// Connect to Project DB
	mysql.SetProjectDB(projectDB.Db.Db)

	// Execute the schema diff
	diffSchema(testConfig, "Unit Test", false)

	projectDB.ExpectionsMet("TestDiffSchema", t)

	mgmtDB.ExpectionsMet("TestDiffSchema", t)

}

func TestRecreateProjectDatabase(t *testing.T) {
	var projectDB test.ProjectDB

	var err error

	// Test Configuration
	testConfig := config.Config{
		Project: config.Project{
			DB: config.DB{
				Database: "project",
			},
		},
	}

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

	mgmtDb.MigrationCount(test.DBRow{0})

	mgmtDb.MigrationInsert(test.DBRow{1, "UnitTestProject", "abc123", mysql.GetTimeNow(), "unit test", 0})

	mgmtDb.MigrationInsertStep(test.DBRow{1, 0, 4, "address", "ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;", "ALTER TABLE `dogs` DROP COLUMN `address`;", "", 0})

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
