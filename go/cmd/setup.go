package cmd

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
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
		Action: func(ctx *cli.Context) *cli.ExitError {

			if !ctx.IsSet("management") && !ctx.IsSet("existing") {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("Please specify a setup action", 1)
			}

			// Parse global flags
			parseGlobalFlags(ctx)

			// Read configuration and access the management database
			conf, configError := configureManagement()
			if configError != nil {
				return cli.NewExitError("Configuration Load failed.", 1)
			}

			if ctx.IsSet("management") {

				if configError != nil {
					// Build the schema for the management database
					err := management.BuildSchema(conf)

					if util.ErrorCheck(err) {
						return cli.NewExitError(fmt.Sprintf("Management Database Setup: Building tables FAILED with error: %v", err), 1)
					}
					return cli.NewExitError("Management Database Setup completed successfully.", 0)
				}
				return cli.NewExitError("Management Database Setup: Building tables FAILED because the Management DB is already setup", 1)
			} else if ctx.IsSet("existing") {

				if configError != nil {
					return cli.NewExitError("Configuration Load failed.", 1)
				}

				return setupExistingDB(conf)
			}

			return cli.NewExitError("No action performed.", 0)
		},
	}
	return setup
}

func setupExistingDB(conf config.Config) *cli.ExitError {

	const YES, NO = "yes", "no"
	action := NO

	// Read the MySQL Database and generate Tables
	err := mysql.ReadTables()
	if util.ErrorCheck(err) {
		return cli.NewExitError("Setup Existing failed. Unable to read MySQL Tables", 1)
	}

	tables := []string{}
	for _, tbl := range mysql.Schema {
		if tbl.Metadata.MDID == 0 {
			tables = append(tables, tbl.Name)
		}
	}

	// actionMsg := "Found the following unmanaged tables in the project database:\n"
	// actionMsg += strings.Join(tables, "\n")
	// actionMsg += fmt.Sprintf("\nDo you want to register these tables for migrations?")

	// action, err = util.SelectAction(actionMsg, []string{YES, NO})
	action = YES

	if !util.ErrorCheckf(err, "There was a error while determining how to proceed. Cancelling setup.") {
		if action == YES {

			path := util.WorkingSubDir(strings.ToLower(conf.Project.Name))

			// Generate PropertyIds for all Database properties
			for i := 0; i < len(mysql.Schema); i++ {
				tbl := &mysql.Schema[i]
				tbl.GeneratePropertyIDs()

				// Generate YAML from the Tables and write to the working folder
				err = yaml.WriteTable(path, *tbl)

				if err != nil {
					return cli.NewExitError(fmt.Sprintf("Existing Database Setup FAILED.  Unable to create YAML Table: %s due to error: %v", path, err), 1)
				}

				err = tbl.InsertMetadata()
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("Existing Database Setup FAILED.  Unable to insert metdata for Table: %s due to error: %v", tbl.Name, err), 1)
				}
				util.LogInfof("Registering Table for migrations: %s", mysql.Schema[i].Name)
			}

			return cli.NewExitError("Existing Database Setup Completed. Generated YAML definitions in path: "+path, 0)

		}
	}

	return cli.NewExitError("Management Database Setup Failed: Invalid option.", 1)

}
