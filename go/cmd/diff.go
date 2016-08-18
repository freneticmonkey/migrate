package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/git"
	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/urfave/cli"
)

// GetDiffCommand Configure the validate command
func GetDiffCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "diff",
		Usage: "Diff the MySQL target database and the YAML schema.",
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
		},
		Action: func(ctx *cli.Context) error {

			var version string
			var project string

			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			conf, err := configureManagement()

			if err != nil {
				return cli.NewExitError("Configuration Load failed.", 1)
			}

			// Override the project settings with the command line flags
			if ctx.IsSet("version") {
				version = ctx.String("version")
			}

			if ctx.IsSet("project") {
				project = ctx.String("project")
			}

			return diff(version, project, conf)
		},
	}
	return setup
}

func diff(project, version string, conf config.Config) *cli.ExitError {
	var forwardDiff table.Differences
	var problems int
	var err error

	// Override the project settings with the command line flags
	if version != "" {
		conf.Project.Schema.Version = version
	}

	if project != "" {
		conf.Project.Name = project
		git.Clone(conf.Project)
	} else {
		util.LogInfo("No project specified.  Comparing the current state of the YAML schema in working path.")
	}

	// Read the YAML files cloned from the repo
	err = yaml.ReadTables(conf.Options.WorkingPath)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Diff failed. Unable to read YAML Tables", 1)
	}
	problems, err = id.ValidateSchema(yaml.Schema, "YAML Schema")
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. YAML Errors found", problems)
	}

	// Read the MySQL tables from the target database
	err = mysql.ReadTables()
	if util.ErrorCheck(err) {
		return cli.NewExitError("Diff failed. Unable to read MySQL Tables", 1)
	}
	problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema")
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. Problems with Target Database Detected", problems)
	}

	forwardDiff, err = table.DiffTables(yaml.Schema, mysql.Schema, true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. Problems determining differences", 1)
	}
	mysql.GenerateAlters(forwardDiff)

	completeMessage := "Diff completed successfully."

	if len(forwardDiff.Slice) > 0 {
		completeMessage += fmt.Sprintf(" %d differences found.", len(forwardDiff.Slice))
	} else {
		completeMessage += " No differences found."
	}

	return cli.NewExitError(completeMessage, 0)
}
