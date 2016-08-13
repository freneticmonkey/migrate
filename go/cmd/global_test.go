package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
)

func Teardown() {
	// Empty Schema
	yaml.Schema = []table.Table{}
	mysql.Schema = []table.Table{}

	// Disassociate Test Databases
	// Connect to Project DB
	exec.SetProjectDB(nil)
	management.SetManagementDB(nil)
	mysql.SetProjectDB(nil)
	exec.Setup(nil, 0, "")
	migration.Setup(nil, 1)
	metadata.Setup(nil, 1)

	util.SetVerbose(false)
}

func GetMySQLCreateTableDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `unittestproject_dogs` (",
		"`id` int(11) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "\n")
}

func GetCreateTableDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `unittestproject_dogs` (",
		"`id` int(11) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "")
}

func GetCreateTableAddressColumnDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `unittestproject_dogs` (",
		"`id` int(11) NOT NULL,",
		"`address` varchar(128) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "")
}

func GetYAMLTableDogs() string {
	return `id: table_unittestproject_dogs
name: dogs
engine: InnoDB
charset: latin1
columns:
- id: unittestproject_dogs_col_id
  name: id
  type: int
  size: [11]
primaryindex:
  id: unittestproject_dogs_primarykey
  name: PrimaryKey
  columns:
  - name: id
  isprimary: true
`
}

func GetTableDogs() table.Table {
	return table.Table{
		Name:    "unittestproject_dogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					MDID:       2,
					DB:         1,
					PropertyID: "unittestproject_dogs_col_id",
					ParentID:   "table_unittestproject_dogs",
					Name:       "id",
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
				MDID:       3,
				DB:         1,
				PropertyID: "unittestproject_dogs_primarykey",
				ParentID:   "table_unittestproject_dogs",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Metadata: metadata.Metadata{
			MDID:       1,
			DB:         1,
			PropertyID: "table_unittestproject_dogs",
			Name:       "unittestproject_dogs",
			Type:       "Table",
		},
	}
}

func GetYAMLTableAddressDogs() string {
	return `id: table_unittestproject_dogs
name: dogs
engine: InnoDB
charset: latin1
columns:
- id: unittestproject_dogs_col_id
  name: id
  type: int
  size: [11]
- id: unittestproject_dogs_col_address
  name: address
  type: varchar
  size: [128]
primaryindex:
  id: unittestproject_dogs_primarykey
  name: PrimaryKey
  columns:
  - name: id
  isprimary: true
`
}

func GetTableAddressDogs() table.Table {
	return table.Table{
		Name:    "unittestproject_dogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					MDID:       2,
					DB:         1,
					PropertyID: "unittestproject_dogs_col_id",
					ParentID:   "table_unittestproject_dogs",
					Name:       "id",
					Type:       "Column",
				},
			},
			{
				Name: "address",
				Type: "varchar",
				Size: []int{128},
				Metadata: metadata.Metadata{
					// MDID is not defined here as this
					// instance is typically used to test diffing,
					// during which this column needs to be inserted
					// into the DB and as such, the trigger for
					// insertion is MDID < 1
					DB:         1,
					PropertyID: "unittestproject_dogs_col_address",
					ParentID:   "table_unittestproject_dogs",
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
				MDID:       3,
				DB:         1,
				PropertyID: "unittestproject_dogs_primarykey",
				ParentID:   "table_unittestproject_dogs",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Metadata: metadata.Metadata{
			MDID:       1,
			DB:         1,
			PropertyID: "table_unittestproject_dogs",
			Name:       "unittestproject_dogs",
			Type:       "Table",
		},
	}
}

func TestConfigReadFile(t *testing.T) {
	var mgmtDB test.ManagementDB
	testName := "TestConfigReadFile"

	// TODO: Provide config
	configFilename := "config.yml"
	var configContents = `
    options:
        management:
            db:
                username: root
                password: test
                ip:       127.0.0.1
                port:     3400
                database: management

    # Project Definition
    project:
        # Project name - used to identify the project by the cli flags
        # and configure the table's namespace
        name: "animals"
        db:
            username:    root
            password:    test
            ip:          127.0.0.1
            port:        3500
            database:    test
            environment: UNITTEST
    `
	expectedConfig := config.Config{
		Options: config.Options{
			Management: config.Management{
				DB: config.DB{
					Username:    "root",
					Password:    "test",
					Ip:          "127.0.0.1",
					Port:        3400,
					Database:    "management",
					Environment: "",
				},
			},
		},
		Project: config.Project{
			Name: "animals",
			DB: config.DB{
				Username:    "root",
				Password:    "test",
				Ip:          "127.0.0.1",
				Port:        3500,
				Database:    "test",
				Environment: "UNITTEST",
			},
		},
	}

	// Set Testing FileSystem
	util.SetConfigTesting()
	util.Config(expectedConfig)

	// Write a test configuration YAML file
	err := util.WriteFile(configFilename, []byte(configContents), 0644)

	if err != nil {
		t.Errorf("Config Read File: Write test config FAILED with Error: %v", err)
		return
	}

	// manually setting the default global config filename
	configFile = configFilename

	// Check for mananagement tables

	// Setup the mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

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

	//  Get Database from Project table - Add an entry for the SANDBOX database
	mgmtDB.DatabaseGet(
		expectedConfig.Project.Name,
		expectedConfig.Project.DB.Database,
		expectedConfig.Project.DB.Environment,
		test.DBRow{1, "UnitTestProject", "project", "SANDBOX"},
		false,
	)

	// Set the management DB
	management.SetManagementDB(mgmtDB.Db)

	fileConfig, err := configureManagement()

	if err != nil {
		t.Errorf("Config Read File FAILED with Error: %v", err)
		return
	}

	if !reflect.DeepEqual(expectedConfig, fileConfig) {
		t.Error("Config Read File FAILED. Returned config does not match.")
		util.LogWarn("Config Read File FAILED. Returned config does not match.")
		util.DebugDumpDiff(expectedConfig, fileConfig)
	}

	Teardown()
}

func TestConfigReadURL(t *testing.T) {

	// TODO: Provide config
	var remoteConfig = `
    options:
        management:
            db:
                username: root
                password: test
                ip:       127.0.0.1
                port:     3400
                database: management

    # Project Definition
    project:
        # Project name - used to identify the project by the cli flags
        # and configure the table's namespace
        name: "animals"
        db:
            username:    root
            password:    test
            ip:          127.0.0.1
            port:        3500
            database:    test
            environment: UNITTEST
    `
	expectedConfig := config.Config{
		Options: config.Options{
			Management: config.Management{
				DB: config.DB{
					Username:    "root",
					Password:    "test",
					Ip:          "127.0.0.1",
					Port:        3400,
					Database:    "management",
					Environment: "",
				},
			},
		},
		Project: config.Project{
			Name: "animals",
			DB: config.DB{
				Username:    "root",
				Password:    "test",
				Ip:          "127.0.0.1",
				Port:        3500,
				Database:    "test",
				Environment: "UNITTEST",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, remoteConfig)
	}))
	defer ts.Close()

	urlConfig, err := loadConfig(ts.URL, "")

	if err != nil {
		t.Errorf("Config Read URL FAILED with Error: %v", err)
	}
	if !reflect.DeepEqual(expectedConfig, urlConfig) {
		t.Error("Config Read URL FAILED. Returned config does not match.")
		util.LogWarn("Config Read URL FAILED. Returned config does not match.")
		util.DebugDumpDiff(expectedConfig, urlConfig)
	}

	Teardown()
}
