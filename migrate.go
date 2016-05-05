package main

import (
	"flag"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
)

var conf config.Config

var flags struct {
	config string
}

//
func processFlags() {
	flags.config = "config.yml"
	flag.Parse()
}

// Read Configuration
func readConfig() {
	err := yaml.ReadFile(flags.config, &conf)
	util.ErrorCheck(err)
	util.LogInfo("Configuration Read Success: " + flags.config)
}

func main() {
	processFlags()

	util.LogInfo("Migrations!")

	readConfig()

	//   LogInfo("Running Git functions")
	//   git.Clone(config.Project.Schema)

	// yaml.ReadTables(config.Options.WorkingPath)
	//
	// mysql.ReadTables(config.DB)
	//
	// differences := migrate.DiffTables(yaml.Schema, mysql.Schema)
	//
	// migrate.GenerateMySQLAlters(differences)

	// yamlPath := filepath.Join(config.Options.WorkingPath, config.Project.Name)
	//yaml.WriteTables(yamlPath, migrate.DBSchema.Tables)
}
