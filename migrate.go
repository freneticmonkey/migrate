package main

import (
	"flag"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/git"
	"github.com/freneticmonkey/migrate/migrate/mysql"
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

	// Initialise any utility configuration
	util.Config(conf)
}

func main() {
	processFlags()

	util.LogInfo("Migrations!")

	readConfig()

	util.LogInfo("Running Git functions")
	git.Clone(conf.Project)

	// Read the YAML files cloned from the repo
	yaml.ReadTables(conf.Options.WorkingPath)

	// Read the MySQL tables from the target database
	mysql.ReadTables(conf.Project)
	//
	// differences := migrate.DiffTables(yaml.Schema, mysql.Schema)
	//
	// migrate.GenerateMySQLAlters(differences)

	// yamlPath := filepath.Join(config.Options.WorkingPath, config.Project.Name)
	//yaml.WriteTables(yamlPath, migrate.DBSchema.Tables)
}
