package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/util"
)

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
}

func TestRefreshDatabase(t *testing.T) {
	// Configure management

	// Setup Configuration somehow

	// Reset/recreate management database?

	// Setup expected database operations

	// Execute the migration
	err := sandboxProcessFlags(true, false, false, false)

	if err != nil {
		t.Errorf("Sandbox Refresh FAILED with Error: %v", err)
	}

	// Verify database operations
}

func TestNewTableApplyImmediately(t *testing.T) {

}

func TestNewColumnApplyImmediately(t *testing.T) {

}
