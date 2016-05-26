package main

import (
	"flag"
	"os"

	"github.com/urfave/cli"

	"github.com/freneticmonkey/migrate/migrate/cmd"
	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/management"
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

	app := cli.NewApp()
	app.Name = "migrate"
	app.Usage = "Migrate MySQL databases using a YAML defined target schema"
	app.Author = "Scott Porter"
	app.Copyright = "MIT"
	app.Email = "scottporter@neuroticstudios.com"
	app.Version = "0.0.1"

	// Configure the app

	app.Flags = cmd.GetGlobalFlags()

	app.Commands = []cli.Command{
		cmd.GetSetupCommand(),
		cmd.GetDiffCommand(&conf),
		cmd.GetValidateCommand(),
		cmd.GetCreateCommand(),
		cmd.GetExecCommand(),
	}
	app.Before = func(ctx *cli.Context) error {

		configURL := ctx.String("config-url")
		err := yaml.ReadFile(configURL, &conf)
		if !util.ErrorCheckf(err, "Configuration read failed for: %s", configURL) {
			util.LogInfo("Configuration Read Success: " + configURL)

			// Initialise any utility configuration
			util.Config(conf)

			// Configure access to the management DB
			management.Setup(conf)

		} else {
			return cli.NewExitError("Unable to read configuration: "+configURL, 1)
		}
		return nil
	}

	app.Run(os.Args)
	/*
		processFlags()

		util.LogInfo("Migrations!")

		readConfig()

		management.Setup(conf)

		util.LogInfo("Running Git functions")
		git.Clone(conf.Project)

		// Read the YAML files cloned from the repo
		yaml.ReadTables(conf.Options.WorkingPath)

		// Read the MySQL tables from the target database
		mysql.ReadTables(conf.Project)

		forwardDiff := table.DiffTables(yaml.Schema, mysql.Schema)
		forwardOps := mysql.GenerateAlters(forwardDiff)

		backwardDiff := table.DiffTables(mysql.Schema, yaml.Schema)
		backwardOps := mysql.GenerateAlters(backwardDiff)

		version := conf.Project.Version
		if len(version) == 0 {
			version, _ = git.GetVersion(conf.Project.Name)
		}

		ts, _ := git.GetVersionTime(conf.Project.Name, version)
		info, _ := git.GetVersionDetails(conf.Project.Name, version)

		m, err := migration.New(migration.Param{
			Project:     conf.Project.Name,
			Version:     version,
			Timestamp:   ts,
			Description: info,
			Forwards:    forwardOps,
			Backwards:   backwardOps,
		})

		if !util.ErrorCheckf(err, "Unable to create Migration for Project: [%s] Version: [%s]", conf.Project.Name, version) {

			util.LogInfof("Created Migration with ID: %d", m.MID)

			exec.Exec(exec.ExecOptions{
				MID:              m.MID,
				Dryrun:           false,
				Force:            true,
				Rollback:         false,
				PTODisabled:      true,
				AllowDestructive: true,
			})

			// yamlPath := filepath.Join(config.Options.WorkingPath, config.Project.Name)
			//yaml.WriteTables(yamlPath, migrate.DBSchema.Tables)
		}
	*/
}
