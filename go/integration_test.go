package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/testdata"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
)

func getYAMLTableDogs() string {
	return `id: dogs
name: dogs
engine: InnoDB
charset: latin1
columns:
- id: id
  name: id
  type: int
  size: [11]
primaryindex:
  id: primarykey
  name: PrimaryKey
  columns:
  - name: id
  isprimary: true
`
}

func getYAMLNamespacedTableDogs() string {
	return `id: ssdogs
name: ssdogs
engine: InnoDB
charset: latin1
columns:
- id: id
  name: id
  type: int
  size: [11]
primaryindex:
  id: primarykey
  name: PrimaryKey
  columns:
  - name: id
  isprimary: true
`
}

func getTableDogs() table.Table {
	return table.Table{
		ID:      "dogs",
		Name:    "dogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				ID:   "id",
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					MDID:       0,
					DB:         0,
					PropertyID: "id",
					ParentID:   "dogs",
					Name:       "id",
					Type:       "Column",
				},
			},
		},
		PrimaryIndex: table.Index{
			ID:        "primarykey",
			Name:      "PrimaryKey",
			IsPrimary: true,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				MDID:       0,
				DB:         0,
				PropertyID: "primarykey",
				ParentID:   "dogs",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Namespace: table.Namespace{
			SchemaName:    "",
			TablePrefix:   "",
			Path:          "",
			TableName:     "dogs",
			TableFilename: "dogs",
		},
		Metadata: metadata.Metadata{
			MDID:       0,
			DB:         0,
			PropertyID: "dogs",
			Name:       "dogs",
			Type:       "Table",
		},
	}
}

func getTableNamespacedDogs() table.Table {
	return table.Table{
		ID:      "ssdogs",
		Name:    "ssdogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				ID:   "id",
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					MDID:       0,
					DB:         0,
					PropertyID: "id",
					ParentID:   "ssdogs",
					Name:       "id",
					Type:       "Column",
				},
			},
		},
		PrimaryIndex: table.Index{
			ID:        "primarykey",
			Name:      "PrimaryKey",
			IsPrimary: true,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				MDID:       0,
				DB:         0,
				PropertyID: "primarykey",
				ParentID:   "ssdogs",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Namespace: table.Namespace{
			SchemaName:    "Schema",
			TablePrefix:   "ss",
			Path:          "schema",
			TableName:     "dogs",
			TableFilename: "dogs",
		},
		Metadata: metadata.Metadata{
			MDID:       0,
			DB:         0,
			PropertyID: "ssdogs",
			Name:       "ssdogs",
			Type:       "Table",
		},
	}
}

func writeTableYAMLToPath(path string, tableName string, contents string) {

	////////////////////////////////////////////////////////
	// Configure source YAML files for Schema read
	//

	path = fmt.Sprintf("%s/%s.yml", strings.ToLower(path), tableName)

	err := test.WriteFile(
		path,
		contents,
		0644,
		false,
	)

	if err != nil {
		util.LogErrorf("What? %v", err)
	}
}

func writeTableStructToPath(testName string, t *testing.T, path string, expectedContent string, tbl table.Table) {

	var err error
	var exists bool
	var data []byte

	util.LogErrorf("Writing to PATH: %s", path)

	err = yaml.WriteTable(path, tbl)

	if err != nil {
		t.Errorf("%s FAILED to write YAML with err: %v", testName, err)
	}

	// Verify that the generated YAML is in the correct path and in the expected format
	fp := filepath.Join(
		path,
		tbl.Namespace.Path,
		tbl.Name+".yml",
	)

	util.LogErrorf("Checking PATH: %s", fp)

	exists, err = util.FileExists(fp)

	if !exists {
		t.Errorf("%s FAILED YAML NOT written!", testName)
	} else {
		data, err = util.ReadFile(fp)

		if err != nil {
			t.Errorf("%s FAILED to read written YAML with err: %v", testName, err)
		} else {
			tblStr := string(data)

			if tblStr != expectedContent {
				util.DebugDiffString(expectedContent, tblStr)
				t.Errorf("%s FAILED generated YAML doesn't match expected YAML", testName)
			}
		}
	}
}

// TestRead Test Read - no Namespace
func TestRead(t *testing.T) {

	testName := "TestRead"

	util.LogAlert(testName)

	var err error

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// Mock Table structs
	dogsTbl := getTableDogs()

	// Write Table to Path
	writeTableYAMLToPath(testConfig.Project.Name, dogsTbl.Name, getYAMLTableDogs())

	// STARTING Table Read

	err = yaml.ReadTables(testConfig)

	if err != nil {
		t.Errorf("%s FAILED to read YAML with err: %v", testName, err)
	}

	if len(yaml.Schema) != 1 {
		t.Errorf("%s FAILED to read YAML. Schema has invalid number of tables", testName)
	} else {
		if !reflect.DeepEqual(dogsTbl, yaml.Schema[0]) {
			t.Errorf("%s FAILED to read YAML. Read Table differs from expected.", testName)
			util.LogWarn("Result")
			util.DebugDumpDiff(dogsTbl, yaml.Schema[0])
		}
	}

	testdata.Teardown()
}

// TestWrite Test Write - no Namespace
func TestWrite(t *testing.T) {

	testName := "TestWrite"

	util.LogAlert(testName)

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// Mock Table structs - with the new Address Column
	dogsTbl := getTableDogs()

	path := util.WorkingSubDir(strings.ToLower(testConfig.Project.Name))

	writeTableStructToPath(testName, t, path, getYAMLTableDogs(), dogsTbl)

	testdata.Teardown()
}

// TestNamespaceRead Test Read - Namespace
func TestNamespaceRead(t *testing.T) {

	testName := "TestNamespaceRead"

	util.LogAlert(testName)

	var err error

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// Mock Table structs
	dogsTbl := getTableNamespacedDogs()

	path := filepath.Join(
		// strings.ToLower(testConfig.Project.Name),
		"schema",
	)

	path = "schema"

	// Write Table to Path
	writeTableYAMLToPath(path, dogsTbl.Name, getYAMLNamespacedTableDogs())

	// STARTING Table Read

	err = yaml.ReadTables(testConfig)

	if err != nil {
		t.Errorf("%s FAILED to read YAML with err: %v", testName, err)
	}

	if len(yaml.Schema) != 1 {
		t.Errorf("%s FAILED to read YAML. Schema has invalid number of tables", testName)
	} else {
		if !reflect.DeepEqual(dogsTbl, yaml.Schema[0]) {
			t.Errorf("%s FAILED to read YAML. Read Table differs from expected.", testName)
			util.LogWarn("Result")
			util.DebugDumpDiff(dogsTbl, yaml.Schema[0])
		}
	}

	testdata.Teardown()
}

// TestNamespacedWrite Test Write - Namespace
func TestNamespacedWrite(t *testing.T) {

	testName := "TestNamespacedWrite"

	util.LogAlert(testName)

	// Test Configuration
	testConfig := test.GetTestConfig()

	// testdata.Teardown() - Pre test cleanup
	testdata.Teardown()

	util.SetConfigTesting()
	util.Config(testConfig)

	// Mock Table structs - with the new Address Column
	dogsTbl := getTableNamespacedDogs()

	// path := util.WorkingSubDir(testConfig.Project.Name)
	path := util.WorkingSubDir("")

	writeTableStructToPath(testName, t, path, getYAMLNamespacedTableDogs(), dogsTbl)

	testdata.Teardown()
}
