package cmd

import (
	"github.com/freneticmonkey/migrate/go/git"
	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/urfave/cli"
)

// GetValidateCommand Configure the validate command
func GetValidateCommand() (setup cli.Command) {
	problems := 0
	bothType := "both"
	setup = cli.Command{
		Name:  "validate",
		Usage: "Validate the MySQL target database and the YAML schema can be successfully parsed.",
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
				Usage: "Which schema to validate: yaml, mysql",
			},
		},
		Action: func(ctx *cli.Context) error {
			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			conf, err := configureManagement()

			if err != nil {
				return cli.NewExitError("Configuration Load failed.", 1)
			}

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
				err := mysql.ReadTables()
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
