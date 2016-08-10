package test

import (
	"testing"

	"github.com/freneticmonkey/migrate/go/util"
)

func TestCreateFile(t *testing.T) {
	testConfig := GetTestConfig()
	expectedData := `# Project Definition for unit tests
project:
    # Project name - used to identify the project by the cli flags
    # and configure the table's namespace
    name: "UnitTestProject"
    db:
        database:    project
        environment: SANDBOX
    # The Project schema configuration
    schema:
        # Schema name.  Not currently used
        name: "unittestconfig"
        # Git Repo
        url:  "http://git.test.com/testing/schema.git"
        # Default Version of the Schema
        version: "abc123"
        # Subfolders within the Git repo to checkout which contain db schema
        folders:
            - "schema"
            - "schemaTwo"
    # local project settings used for sandbox development
    localschema:
        # This is the working folder, however it is intended to be a path to
        # schema within a cloned repo
        path: "test/working"`

	util.SetConfigTesting()
	util.Config(testConfig)

	CreateTestConfigFile()

	data, err := util.ReadFile("content.yml")
	if err != nil {
		t.Errorf("Test Helper Create File FAILED: There was a problem reading the config file: error [%v]", err)
	}

	if string(data) != expectedData {
		t.Errorf("Test Helper Create File FAILED: File contents doesn't match expected contents: expected: [%s] received: [%s]", expectedData, data)
	}

}
