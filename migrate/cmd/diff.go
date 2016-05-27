package cmd

import (
	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/git"
	"github.com/freneticmonkey/migrate/migrate/id"
	"github.com/freneticmonkey/migrate/migrate/mysql"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
	"github.com/urfave/cli"
)

// GetDiffCommand Configure the validate command
func GetDiffCommand(conf *config.Config) (setup cli.Command) {
	problems := 0
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
			if ctx.IsSet("project") && ctx.IsSet("version") {
				git.Clone(conf.Project)
			}

			// Read the YAML files cloned from the repo
			err := yaml.ReadTables(conf.Options.WorkingPath)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Diff failed. Unable to read YAML Tables", 1)
			}
			problems, err = id.ValidateSchema(yaml.Schema, "YAML Schema")
			if util.ErrorCheck(err) {
				return cli.NewExitError("Validation failed. YAML Errors found", problems)
			}

			// Read the MySQL tables from the target database
			err = mysql.ReadTables(conf.Project)
			if util.ErrorCheck(err) {
				return cli.NewExitError("Diff failed. Unable to read MySQL Tables", 1)
			}
			problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema")
			if util.ErrorCheck(err) {
				return cli.NewExitError("Validation failed. Problems with Target Database Detected", problems)
			}

			forwardDiff := table.DiffTables(yaml.Schema, mysql.Schema)
			mysql.GenerateAlters(forwardDiff)

			return cli.NewExitError("Diff completed successfully.", 0)
		},
	}
	return setup
}
