package git

import (
	"fmt"
	"testing"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
)

func TestGetVersionTime(t *testing.T) {
	project := "animals"
	version := "abc123"

	gitTime := "2006-01-02T15:04:05-07:00"
	expectedMySQLTime := "2006-01-02 22:04:05"

	// Configure unit test shell
	util.SetConfigTesting()
	testConfig := test.GetTestConfig()
	util.Config(testConfig)

	shell := util.GetShell().(*util.MockShellExecutor)

	params := []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--format=%%cI",
	}
	shell.ExpectExec("git", params, gitTime, nil)

	tm, err := GetVersionTime(project, version)

	if err != nil {
		t.Errorf("GetVersionTime failed with error: %v", err)
		return
	}

	if tm != expectedMySQLTime {
		t.Errorf("GetVersionTime: %s didn't match expected time: %s", tm, expectedMySQLTime)
	}

	if err = shell.ExpectationsWereMet(); err != nil {
		t.Errorf("Git GetVersionTime FAILED: Not all shell commands were executed: error [%v]", err)
	}
}

func TestGetVersion(t *testing.T) {
	project := "animals"
	version := "abc123"

	gitVersion := "ab3af30b4ccdcacffe321ef59eb031358a9890d2"

	// Configure unit test shell
	util.SetConfigTesting()
	testConfig := test.GetTestConfig()
	util.Config(testConfig)

	shell := util.GetShell().(*util.MockShellExecutor)

	params := []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--format=%%H",
	}
	shell.ExpectExec("git", params, gitVersion, nil)

	version, err := GetVersion(project)

	if err != nil {
		t.Errorf("GetVersionTime failed with error: %v", err)
		return
	}

	if version != gitVersion {
		t.Errorf("Git GetVersionVersion: %s didn't match expected version: %s", version, gitVersion)
	}

	if err = shell.ExpectationsWereMet(); err != nil {
		t.Errorf("Git GetVersionVersion FAILED: Not all shell commands were executed: error [%v]", err)
	}

}

func TestGetVersionDetails(t *testing.T) {
	project := "animals"
	version := "abc123"

	gitDetails := `commit ab3af30b4ccdcacffe321ef59eb031358a9890d2
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 12 11:52:03 2016 +1000

    Updated column size declarations to new array format: Missed a column`

	// Configure unit test shell
	util.SetConfigTesting()
	testConfig := test.GetTestConfig()
	util.Config(testConfig)

	shell := util.GetShell().(*util.MockShellExecutor)

	params := []string{
		"-C",
		util.WorkingSubDir(project),
		"show",
		"-s",
		"--pretty=medium",
	}
	shell.ExpectExec("git", params, gitDetails, nil)

	details, err := GetVersionDetails(project, version)

	if err != nil {
		t.Errorf("GetVersionDetails failed with error: %v", err)
		return
	}

	if details != gitDetails {
		t.Errorf("GetVersionDetails: [%s] didn't match expected details: [%s]", details, gitDetails)
	}
}

func TestClone(t *testing.T) {
	var sparseExists bool
	var data []byte
	version := "abc123"

	expectedSparseFile := `schema/*
schemaTwo/*`

	testConfig := test.GetTestConfig()

	// Add additional git configuration
	testConfig.Project.Git.Url = "http://git.test.com/test/repo"
	testConfig.Project.Git.Version = version
	testConfig.Project.Schema.Namespaces = []config.SchemaNamespace{
		{
			Name:        "Schema",
			TablePrefix: "SS",
			SchemaPath:  "schema",
		},
		{
			Name:        "SchemaTwo",
			TablePrefix: "ST",
			SchemaPath:  "schemaTwo",
		},
	}

	// Configure unit test shell
	util.SetConfigTesting()
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
		testConfig.Project.Git.Url,
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
		testConfig.Project.Git.Version,
	}
	shell.ExpectExec("git", params, "", nil)

	err := Clone(testConfig.Project)

	if err != nil {
		t.Errorf("Git Clone FAILED with error: %v", err)
		return
	}

	// Validate that the sparse checkout file was created correctly
	sparseFile := fmt.Sprintf("%s/.git/info/sparse-checkout", checkoutPath)
	sparseExists, err = util.FileExists(sparseFile)

	if err != nil {
		t.Errorf("Git Clone FAILED: there was an error while reading the sparse checkut file at path: [%s] with error: [%v]", sparseFile, err)
		return
	}

	if !sparseExists {
		t.Errorf("Git Clone FAILED: sparse checkut file missing from path: [%s]", sparseFile)
		return
	}

	data, err = util.ReadFile(sparseFile)
	sparseData := string(data)

	if sparseData != expectedSparseFile {
		t.Errorf("Git Clone FAILED: sparse data file contents: [%s] doesn't match expected contents: [%s]", sparseData, expectedSparseFile)
		return
	}

	// Ensure that all of the anticipated shell calls were made
	if err = shell.ExpectationsWereMet(); err != nil {
		t.Errorf("Git Clone FAILED: Not all shell commands were executed: error [%v]", err)
	}
}
