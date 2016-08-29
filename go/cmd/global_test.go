package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/configsetup"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/testdata"
	"github.com/freneticmonkey/migrate/go/util"
)

func TestConfigReadFile(t *testing.T) {
	var mgmtDB test.ManagementDB
	testName := "TestConfigReadFile"

	// TODO: Provide config
	configFilename := "config.yml"
	var configContents = `
    options:
        namespaces: No
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
			Namespaces: false,
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
	configsetup.SetConfigFile(configFile)
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
		test.DBRow{
			1,
			expectedConfig.Project.Name,
			expectedConfig.Project.DB.Database,
			expectedConfig.Project.DB.Environment,
		},
		false,
	)

	// Set the management DB
	management.SetManagementDB(mgmtDB.Db)

	fileConfig, err := configsetup.ConfigureManagement()

	if err != nil {
		t.Errorf("Config Read File FAILED with Error: %v", err)
		return
	}

	if !reflect.DeepEqual(expectedConfig, fileConfig) {
		t.Error("Config Read File FAILED. Returned config does not match.")
		util.LogWarn("Config Read File FAILED. Returned config does not match.")
		util.DebugDumpDiff(expectedConfig, fileConfig)
	}

	mgmtDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}

func TestConfigReadURL(t *testing.T) {
	var mgmtDB test.ManagementDB
	var err error

	testName := "TestConfigReadURL"

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

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		// migration.Setup(mgmtDB.Db, 1)
		exec.Setup(mgmtDB.Db, 1, expectedConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	} else {
		t.Errorf("%s failed with error: %v", testName, err)
		return
	}

	// Configure the mock remote HTTP config host
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, remoteConfig)
	}))
	defer ts.Close()

	configsetup.SetConfigURL(ts.URL)

	urlConfig, err := configsetup.LoadConfig(ts.URL, "")

	if err != nil {
		t.Errorf("Config Read URL FAILED with Error: %v", err)
	}
	if !reflect.DeepEqual(expectedConfig, urlConfig) {
		t.Error("Config Read URL FAILED. Returned config does not match.")
		util.LogWarn("Config Read URL FAILED. Returned config does not match.")
		util.DebugDumpDiff(expectedConfig, urlConfig)
	}
	mgmtDB.ExpectionsMet(testName, t)

	testdata.Teardown()
}
