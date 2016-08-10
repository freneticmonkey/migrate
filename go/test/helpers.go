package test

import "github.com/freneticmonkey/migrate/go/config"

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
