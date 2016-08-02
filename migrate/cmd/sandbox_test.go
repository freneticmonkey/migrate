package cmd

import (
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/exec"
	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/mysql"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/test"
	"github.com/freneticmonkey/migrate/migrate/yaml"
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
		Query:   "SELECT count(*) from metadata WHERE name=\"%s\" and type=\"Table\"",
		Columns: []string{"count"},
		Rows:    [][]driver.Value{{1}},
	}
	query.SetArgs(dogsTbl.Name)

	test.ExpectDB(mgmtMock, query)

	// Search Metadata for `dogs` table query - MySQL
	query = test.DBQueryMock{
		Type:  test.QueryCmd,
		Query: "SELECT * FROM metadata WHERE name=\"%s\"",
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
	query.SetArgs(dogsTbl.Name)
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type:  test.QueryCmd,
		Query: "SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"",
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
	query.SetArgs("id", "tbl1")
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type:  test.QueryCmd,
		Query: "SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"",
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
	query.SetArgs("PrimaryKey", "tbl1")
	test.ExpectDB(mgmtMock, query)

	// Search Metadata for `dogs` table query - YAML
	query = test.DBQueryMock{
		Type:  test.QueryCmd,
		Query: "SELECT * FROM metadata WHERE name=\"%s\"",
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
	query.SetArgs(dogsTbl.Name)
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type:  test.QueryCmd,
		Query: "SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"",
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
	query.SetArgs("PrimaryKey", "tbl1")
	test.ExpectDB(mgmtMock, query)

	query = test.DBQueryMock{
		Type:  test.QueryCmd,
		Query: "SELECT * FROM metadata WHERE name=\"%s\" AND parent_id=\"%s\"",
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
	query.SetArgs("id", "tbl1")
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
		Query:  "DROP DATABASE `%s`",
		Result: sqlmock.NewResult(0, 0),
	}
	query.SetArgs(testConfig.Project.DB.Database)

	test.ExpectDB(projectMock, query)

	// Reuse the query object to define the create database
	query.Query = "CREATE DATABASE `%s`"
	test.ExpectDB(projectMock, query)

	projectMock.ExpectClose()

	// Execute the recreation!
	recreateProjectDatabase(testConfig, false)

	if err = projectMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Test Recreate Database: Project DB queries failed expectations. Error: %s", err)
	}
}

func TestMigrateSandbox(t *testing.T) {

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
