package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/exec"
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
func GetCreateCommand(conf *config.Config) (setup cli.Command) {
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
				Name:  "force-sandbox",
				Usage: "Immediately apply the new migration to the target database. Will only function in the sandbox environment.",
			},
			cli.BoolFlag{
				Name:  "rollback",
				Usage: "Allows for a rollback to be created",
			},
			cli.BoolFlag{
				Name:  "pto-disabled",
				Usage: "Execute the migration without using pt-online-schema-change.",
			},
			cli.BoolFlag{
				Name:  "allow-destructive",
				Usage: "Explictly allow rename and delete migration actions",
			},
		},
		Action: func(ctx *cli.Context) error {
			var problems int
			var ts string
			var info string
			dryrun := false
			force := true
			rollback := false
			PTODisabled := true
			allowDestructive := true

			// Override the project settings with the command line flags
			if ctx.IsSet("project") {
				conf.Project.Name = ctx.String("project")
			}

			if ctx.IsSet("version") {
				conf.Project.Version = ctx.String("version")
			}

			// if the version hasn't been defined
			if len(conf.Project.Name) == 0 {
				return cli.NewExitError("Creation failed.  Unable to generate a migration as no project was defined", 1)
			} else if len(conf.Project.Version) == 0 {
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

			ts, err = git.GetVersionTime(conf.Project.Name, conf.Project.Version)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Create failed. Unable to obtain Version Timestamp from Git checkout", 1)
			}
			info, err = git.GetVersionDetails(conf.Project.Name, conf.Project.Version)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Create failed. Unable to obtain Version Details from Git checkout", 1)
			}

			m, err := migration.New(migration.Param{
				Project:     conf.Project.Name,
				Version:     conf.Project.Version,
				Timestamp:   ts,
				Description: info,
				Forwards:    forwardOps,
				Backwards:   backwardOps,
			})
			if util.ErrorCheck(err) {
				return cli.NewExitError("Create failed. Unable to create new Migration in the management database", 1)
			}

			util.LogInfof("Created Migration with ID: %d", m.MID)

			// Process command line flags
			if ctx.IsSet("force-sandbox") {

				dryrun = ctx.Bool("dryrun")
				force = ctx.Bool("force")
				rollback = ctx.Bool("rollback")
				PTODisabled = ctx.Bool("pto-disabled")
				allowDestructive = ctx.Bool("allow-destructive")

				exec.Exec(exec.ExecOptions{
					MID:              m.MID,
					Dryrun:           dryrun,
					Force:            force,
					Rollback:         rollback,
					PTODisabled:      PTODisabled,
					AllowDestructive: allowDestructive,
				})
				if util.ErrorCheck(err) {
					errmsg := fmt.Sprintf("Create and Execute failed. Unable to execute new Migration with ID: [%d]", m.MID)
					return cli.NewExitError(errmsg, 1)
				}
				util.LogInfof("Successfully executed Migration with ID: %d", m.MID)
			}

			success := fmt.Sprintf("Migration successfully with ID: %d", m.MID)
			return cli.NewExitError(success, 0)
		},
	}
	return setup
}
