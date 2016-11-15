package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/configsetup"
	"github.com/freneticmonkey/migrate/go/git"
	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/urfave/cli"
)

// GetCreateCommand Create a new migration for the target project at the version indicated by hash
func GetCreateCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "create",
		Usage: "This subcommand is used to create a migration and register it with the management database.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "version",
				Value: "",
				Usage: "The target git version",
			},
			cli.BoolFlag{
				Name:  "rollback",
				Usage: "Force a rollback (backward) migration to be created",
			},
			cli.StringFlag{
				Name:  "gitinfo",
				Value: "",
				Usage: "Provide a git info file for the schema in the working path",
			},
			cli.BoolFlag{
				Name:  "no-clone",
				Usage: "Do not clone from git.  Use this when the yaml files have already been checked out.",
			},
		},
		Action: func(ctx *cli.Context) error {
			var version string
			var gitinfo string
			var rollback bool
			var clone bool

			rollback = false
			clone = true

			// Override the project settings with the command line flags
			if ctx.IsSet("version") {
				version = ctx.String("version")
			} else if ctx.IsSet("gitinfo") {
				gitinfo = ctx.String("gitinfo")
			} else {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("Unable to generate a migration.  Please specify a target version to migrate to.", 1)
			}

			if ctx.IsSet("no-clone") {
				clone = false
			}

			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			conf, err := configsetup.ConfigureManagement()

			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", err), 1)
			}

			if ctx.IsSet("rollback") {
				rollback = ctx.Bool("rollback")
			}

			return create(version, gitinfo, clone, rollback, conf)

		},
	}
	return setup
}

func create(version string, gitinfo string, clone bool, rollback bool, conf config.Config) *cli.ExitError {
	var problems id.ValidationErrors
	var ts string
	var info string
	var err error

	// extract the contents of the gitinfo file
	if len(gitinfo) > 0 {
		version, info, ts, err = git.GetVersionDetailsFile(gitinfo)

		util.LogInfof("Detected gitinfo file. Parsed:\nVersion: %s\nTime: %s\nInfo:\n>>>\n%s\n<<<", version, ts, info)

		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Creation failed.  Unable to read the gitinfo file: %v", err), 1)
		}
	}

	// Override the project settings with the command line flags
	if version != "" {
		conf.Project.Schema.Version = version
	} else {
		// if the version hasn't been defined
		return cli.NewExitError("Creation failed.  Unable to generate a migration as no version was defined to migrate to", 1)
	}

	util.LogInfo("Creating Migration for Version: " + version)

	// If the schema needs to be cloned from git.  Extract it now
	if clone {
		// Clone the target Git Repo
		git.Clone(conf.Project)

		ts, err = git.GetVersionTime(conf.Project.Name, conf.Project.Schema.Version)
		if util.ErrorCheck(err) {
			return cli.NewExitError("Create failed. Unable to obtain Version Timestamp from Git checkout", 1)
		}
		info, err = git.GetVersionDetails(conf.Project.Name, conf.Project.Schema.Version)
		if util.ErrorCheck(err) {
			return cli.NewExitError("Create failed. Unable to obtain Version Details from Git checkout", 1)
		}
	}

	// Read the YAML files cloned from the repo
	err = yaml.ReadTables(conf)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Creation failed. Unable to read YAML Tables", 1)
	}
	problems, err = id.ValidateSchema(yaml.Schema, "YAML Schema", true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Creation failed. YAML Validation Errors Detected", problems.Count())
	}

	// Read the MySQL tables from the target database
	err = mysql.ReadTables(conf)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Creation failed. Unable to read MySQL Tables", 1)
	}
	problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema", true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Creation failed. Target Database Validation Errors Detected", problems.Count())
	}

	problems, err = id.ValidatePropertyIDs(yaml.Schema, mysql.Schema, true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. Detected YAML PropertyID problems", problems.Count())
	}

	forwardDiff, err := table.DiffTables(yaml.Schema, mysql.Schema, false)
	if util.ErrorCheckf(err, "Diff Failed while generating forward migration") {
		return cli.NewExitError("Create failed. Unable to generate a forward migration", 1)
	}
	forwardOps := mysql.GenerateAlters(forwardDiff)

	backwardDiff, err := table.DiffTables(mysql.Schema, yaml.Schema, false)
	if util.ErrorCheckf(err, "Diff Failed while generating backward migration") {
		return cli.NewExitError("Create failed. Unable to generate a backward migration", 1)
	}
	backwardOps := mysql.GenerateAlters(backwardDiff)

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
}
