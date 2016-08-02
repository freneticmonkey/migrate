package cmd

import (
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

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
	var pdb *gorp.DbMap
	var projectMock sqlmock.Sqlmock
	var err error

	// Test Configuration
	testConfig := config.Config{
		Project: config.Project{
			DB: config.DB{
				Database: "test",
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
	pdb, projectMock, err = test.CreateMockDB()

	if err != nil {
		t.Errorf("TestDiffSchema: Setup Project DB Failed with Error: %v", err)
	}

	// Configure the MySQL Read Tables queries

	// SHOW TABLES Query
	query := test.DBQueryMock{
		Type:    test.QueryCmd,
		Query:   "show tables",
		Columns: []string{"table"},
		Rows: [][]driver.Value{
			{
				dogsTbl.Name,
			},
		},
	}

	test.ExpectDB(projectMock, query)

	// SHOW CREATE TABLE Query
	query = test.DBQueryMock{
		Type:    test.QueryCmd,
		Query:   "show create table dogs",
		Columns: []string{"name", "create_table"},
		Rows: [][]driver.Value{
			{
				dogsTbl.Name,
				dogsTableStr,
			},
		},
	}

	test.ExpectDB(projectMock, query)

	// Configure the Mock Managment DB

	mgmtDb, mgmtMock, err := test.CreateMockDB()

	if err != nil {
		t.Errorf("TestDiffSchema: Setup Project DB Failed with Error: %v", err)
	}

	query = test.DBQueryMock{
		Type:    test.QueryCmd,
		Columns: []string{"count"},
		Rows:    [][]driver.Value{{1}},
	}
	query.FormatQuery("SELECT count(*) from metadata WHERE name=\"%s\" and type=\"Table\"", dogsTbl.Name)

	test.ExpectDB(mgmtMock, query)

	// Search Metadata for `dogs` table query - MySQL
	query = test.DBQueryMock{
		Type: test.QueryCmd,

		Columns: []string{
			"mdid",
			"db",
			"property_id",
			"parent_id",
			"type",
			"name",
			"exists",
		},
		Rows: [][]driver.Value{
			{1, 1, "tbl1", "", "Table", "dogs", 1},
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\"", dogsTbl.Name)
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type: test.QueryCmd,
		Columns: []string{
			"mdid",
			"db",
			"property_id",
			"parent_id",
			"type",
			"name",
			"exists",
		},
		Rows: [][]driver.Value{
			{2, 1, "col1", "tbl1", "Column", "id", 1},
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", "id", "tbl1")
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type: test.QueryCmd,
		Columns: []string{
			"mdid",
			"db",
			"property_id",
			"parent_id",
			"type",
			"name",
			"exists",
		},
		Rows: [][]driver.Value{
			{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", "PrimaryKey", "tbl1")
	test.ExpectDB(mgmtMock, query)

	// Search Metadata for `dogs` table query - YAML
	query = test.DBQueryMock{
		Type: test.QueryCmd,
		Columns: []string{
			"mdid",
			"db",
			"property_id",
			"parent_id",
			"type",
			"name",
			"exists",
		},
		Rows: [][]driver.Value{
			{1, 1, "tbl1", "", "Table", "dogs", 1},
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\"", dogsTbl.Name)
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type: test.QueryCmd,
		Columns: []string{
			"mdid",
			"db",
			"property_id",
			"parent_id",
			"type",
			"name",
			"exists",
		},
		Rows: [][]driver.Value{
			{3, 1, "pi", "tbl1", "PrimaryKey", "PrimaryKey", 1},
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", "PrimaryKey", "tbl1")
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type: test.QueryCmd,
		Columns: []string{
			"mdid",
			"db",
			"property_id",
			"parent_id",
			"type",
			"name",
			"exists",
		},
		Rows: [][]driver.Value{
			{2, 1, "col1", "tbl1", "Column", "id", 1},
		},
	}
	query.FormatQuery("SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"", "id", "tbl1")
	test.ExpectDB(mgmtMock, query)

	// Configure metadata
	metadata.Setup(mgmtDb, 1)

	// Connect to Project DB
	mysql.SetProjectDB(pdb.Db)

	// Execute the schema diff
	diffSchema(testConfig, "Unit Test", false)

	if err = projectMock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestDiffSchema: Project DB queries failed expectations. Error: %s", err)
	}

	if err = mgmtMock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestDiffSchema: Management DB queries failed expectations. Error: %s", err)
	}

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
		Result: sqlmock.NewResult(0, 0),
	}
	query.FormatQuery("DROP DATABASE `%s`", testConfig.Project.DB.Database)

	test.ExpectDB(projectMock, query)

	// Reuse the query object to define the create database
	query.FormatQuery("CREATE DATABASE `%s`", testConfig.Project.DB.Database)
	test.ExpectDB(projectMock, query)

	projectMock.ExpectClose()

	// Execute the recreation!
	recreateProjectDatabase(testConfig, false)

	if err = projectMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Test Recreate Database: Project DB queries failed expectations. Error: %s", err)
	}
}

func TestMigrateSandbox(t *testing.T) {

	var mgmtDb *gorp.DbMap
	var mgmtMock sqlmock.Sqlmock
	var err error

	var m migration.Migration

	// Configure the Mock Managment DB

	mgmtDb, mgmtMock, err = test.CreateMockDB()

	if err != nil {
		t.Errorf("TestMigrateSandbox: Setup Project DB Failed with Error: %v", err)
	} else {
		// Connect to Project DB
		migration.Setup(mgmtDb, 1)
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

	query := test.DBQueryMock{
		Type:    test.QueryCmd,
		Query:   "select count(*) from migration",
		Columns: []string{"count"},
		Rows:    [][]driver.Value{{0}},
	}

	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type:   test.ExecCmd,
		Query:  "insert into `migration` (`mid`,`db`,`project`,`version`,`version_timestamp`,`version_description`,`status`) values (null,?,?,?,?,?,?);",
		Result: sqlmock.NewResult(1, 1),
	}
	query.SetArgs(1, "UnitTestProject", "abc123", mysql.GetTimeNow(), "unit test", 0)

	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type:   test.ExecCmd,
		Query:  "insert into `migration_steps` (`sid`,`mid`,`op`,`mdid`,`name`,`forward`,`backward`,`output`,`status`) values (null,?,?,?,?,?,?,?,?);",
		Result: sqlmock.NewResult(1, 1),
	}
	query.SetArgs(1, 0, 4, "address", "ALTER TABLE `dogs` COLUMN `address` varchar(128) NOT NULL;", "ALTER TABLE `dogs` DROP COLUMN `address`;", "", 0)

	test.ExpectDB(mgmtMock, query)

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

	if err = mgmtMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Test Recreate Database: Management DB queries failed expectations. Error: %s", err)
	}

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
