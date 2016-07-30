package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/git"
	"github.com/freneticmonkey/migrate/migrate/id"
	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/mysql"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
	"github.com/urfave/cli"
)

// GetCreateCommand Create a new migration for the target project at the version indicated by hash
func GetCreateCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "create",
		Usage: "This subcommand is used to create a migration and register it with the management database.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "project",
				Value: "",
				Usage: "The target project",
			},
			cli.StringFlag{
				Name:  "version",
				Value: "",
				Usage: "The target git version",
			},
			cli.BoolFlag{
				Name:  "rollback",
				Usage: "Force a rollback (backward) migration to be created",
			},
		},
		Action: func(ctx *cli.Context) error {
			var problems int
			var ts string
			var info string
			rollback := false

			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			configureManagement()

			// Override the project settings with the command line flags
			if ctx.IsSet("project") {
				conf.Project.Name = ctx.String("project")
			}

			if ctx.IsSet("version") {
				conf.Project.Schema.Version = ctx.String("version")
			}

			// if the version hasn't been defined
			if len(conf.Project.Name) == 0 {
				return cli.NewExitError("Creation failed.  Unable to generate a migration as no project was defined", 1)
			} else if len(conf.Project.Schema.Version) == 0 {
				return cli.NewExitError("Creation failed.  Unable to generate a migration as no version was defined to migrate to", 1)
			} else {
				git.Clone(conf.Project)
			}

			// Read the YAML files cloned from the repo
			err := yaml.ReadTables(conf.Options.WorkingPath)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Creation failed. Unable to read YAML Tables", 1)
			}
			problems, err = id.ValidateSchema(yaml.Schema, "YAML Schema")
			if util.ErrorCheck(err) {
				return cli.NewExitError("Creation failed. YAML Validation Errors Detected", problems)
			}

			// Read the MySQL tables from the target database
			err = mysql.ReadTables(conf.Project)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Creation failed. Unable to read MySQL Tables", 1)
			}
			problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema")
			if util.ErrorCheck(err) {
				return cli.NewExitError("Creation failed. Target Database Validation Errors Detected", problems)
			}

			forwardDiff := table.DiffTables(yaml.Schema, mysql.Schema)
			forwardOps := mysql.GenerateAlters(forwardDiff)

			backwardDiff := table.DiffTables(mysql.Schema, yaml.Schema)
			backwardOps := mysql.GenerateAlters(backwardDiff)

			ts, err = git.GetVersionTime(conf.Project.Name, conf.Project.Schema.Version)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Create failed. Unable to obtain Version Timestamp from Git checkout", 1)
			}
			info, err = git.GetVersionDetails(conf.Project.Name, conf.Project.Schema.Version)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Create failed. Unable to obtain Version Details from Git checkout", 1)
			}

			m, err := migration.New(migration.Param{
				Project:     conf.Project.Name,
				Version:     conf.Project.Schema.Version,
				Timestamp:   ts,
				Description: info,
				Forwards:    forwardOps,
				Backwards:   backwardOps,
				Rollback:    rollback,
			})
			if util.ErrorCheck(err) {
				return cli.NewExitError("Create failed. Unable to create new Migration in the management database", 1)
			}

			success := fmt.Sprintf("Created Migration successfully with ID: [%d]", m.MID)

			return cli.NewExitError(success, 0)
		},
	}
	return setup
}
