package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

func GetTestConfig() config.Config {
	return config.Config{
		Project: config.Project{
			Name: "UnitTestProject",
			Schema: config.Schema{
				Version: "abc123",
			},
			LocalSchema: config.LocalSchema{
				Path: "ignore",
			},
			DB: config.DB{
				Database:    "project",
				Environment: "SANDBOX",
			},
		},
	}
}

func GetMySQLCreateTableDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"`id` int(11) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "\n")
}

func GetCreateTableDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"`id` int(11) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "")
}

func GetCreateTableAddressColumnDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"`id` int(11) NOT NULL,",
		"`address` varchar(128) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "")
}

func GetTableDogs() table.Table {
	return table.Table{
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
}

func GetTableAddressDogs() table.Table {
	return table.Table{
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
}

func DisableTestConfigReadURL(t *testing.T) {

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
