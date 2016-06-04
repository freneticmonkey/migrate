package cmd

import (
	"github.com/freneticmonkey/migrate/migrate/git"
	"github.com/freneticmonkey/migrate/migrate/id"
	"github.com/freneticmonkey/migrate/migrate/mysql"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
	"github.com/urfave/cli"
)

// GetValidateCommand Configure the validate command
func GetValidateCommand() (setup cli.Command) {
	problems := 0
	bothType := "both"
	setup = cli.Command{
		Name:  "validate",
		Usage: "Validate the MySQL target database and the YAML schema.",
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
			cli.StringFlag{
				Name:  "schema-type",
				Value: bothType,
				Usage: "The target git version",
			},
		},
		Action: func(ctx *cli.Context) error {

			// Setup the management database and configuration settings
			configureManagement(ctx)

			if ctx.IsSet("project") && ctx.IsSet("version") {
				git.Clone(conf.Project)
			}

			schemaType := ctx.String("schema-type")
			if schemaType != "mysql" && schemaType != "yaml" && schemaType != bothType {
				return cli.NewExitError("Validation failed. Unknown schema-type", 1)
			}

			if schemaType == "yaml" || schemaType == bothType {
				// Read the YAML files cloned from the repo
				err := yaml.ReadTables(conf.Options.WorkingPath)
				if util.ErrorCheck(err) {
					return cli.NewExitError("Validation failed. Unable to read YAML Tables", 1)
				}

				// Validate YAML Schema Ids
				problems, err = id.ValidateSchema(yaml.Schema, "YAML Schema")
				if util.ErrorCheck(err) {
					return cli.NewExitError("Validation failed. YAML Errors found", problems)
				}
			}

			if schemaType == "mysql" || schemaType == bothType {
				// Read the MySQL tables from the target database
				err := mysql.ReadTables(conf.Project)
				if util.ErrorCheck(err) {
					return cli.NewExitError("Validation failed. Unable to read MySQL Tables", 1)
				}

				// Validate YAML Schema Ids
				problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema")
				if util.ErrorCheck(err) {
					return cli.NewExitError("Validation failed. Problems with Target Databse detected", problems)
				}
			}

			return cli.NewExitError("Validation completed successfully.", 0)
		},
	}
	return setup
}
