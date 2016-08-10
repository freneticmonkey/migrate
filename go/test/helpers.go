package test

import (
	"os"
	"regexp"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
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

// CreateTestConfigFile Creates a standard configuration YAML in the working path
// which contains all of the unit test default settings
func CreateTestConfigFile() {
	WriteFile(
		"content.yml",
		`
		# Project Definition for unit tests
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
		        path: "test/working"`,
		0644,
		true,
	)
}

// FormatFileContent Removes a prefix \n and any whitespace indentation that matches
// the indentation of the first line of content
func FormatFileContent(content string) (string, error) {
	// Determine the whitespace indent on line 2
	lines := strings.Split(content, "\n")
	if len(lines) > 1 {
		// select all whitespace prefixing the line
		re, err := regexp.Compile("^(\\s+)")
		if err != nil {
			return "", err
		}

		// Extract any whitespace
		whitespace := string(re.Find([]byte(lines[1])))

		// If there is any whitespace
		if len(whitespace) > 0 {
			cleanLines := []string{}

			// skip the first line which should be empty if it's a new line
			for _, line := range lines[1:] {
				cleanLines = append(cleanLines, strings.TrimPrefix(line, whitespace))
			}

			// Reassemble the file content without the prefix newline and indentation
			content = strings.Join(cleanLines, "\n")
		}
	}
	// else don't do anything

	return content, nil
}

// WriteFile Unit Test Helper function for creating files during tests with neat formatting
func WriteFile(path, content string, perm os.FileMode, neaten bool) (err error) {
	if neaten {
		content, err = FormatFileContent(content)
		if err != nil {
			return err
		}
	}
	return util.WriteFile(path, []byte(content), perm)
}
