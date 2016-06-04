package cmd

import (
	"github.com/freneticmonkey/migrate/migrate/management"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/urfave/cli"
)

// GetSetupCommand Configure the setup command
func GetSetupCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "setup",
		Usage: "Setup the migration environment",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "management",
				Usage: "Create the management tables in the management database",
			},
			cli.BoolFlag{
				Name:  "existing",
				Usage: "Read the target database and generate a YAML schema including PropertyIds",
			},
		},
		Action: func(ctx *cli.Context) error {

			if ctx.IsSet("management") {

				// Read configuration and access the management database
				err := configureManagement(ctx)

				// If the management database access throws an error,
				// it's because the schema needs to be created
				if util.ErrorCheck(err) {
					// Build the schema for the management database
					management.BuildSchema(&conf)

					return cli.NewExitError("Management Database Setup completed successfully.", 0)
				}
				return cli.NewExitError("Management Database Setup Failed: Couldn't create Management Database Schema.", 1)
			}

			if ctx.IsSet("existing") {
				// TODO:
				// Read the MySQL Database and generate Tables
				// Generate PropertyIds for all Database properties
				// Generate YAML from the Tables
				// Insert Metadata into Management metadata table
				// Write YAML to working folder
			}

			return cli.NewExitError("Setup completed successfully.", 0)
		},
	}
	return setup
}
