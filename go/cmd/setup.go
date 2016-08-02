package cmd

import (
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
		Action: func(ctx *cli.Context) error {

			// Parse global flags
			parseGlobalFlags(ctx)

			if ctx.IsSet("management") {

				// Read configuration and access the management database
				conf, err := configureManagement()

				if err != nil {
					return cli.NewExitError("Configuration Load failed.", 1)
				}

				// If the management database access throws an error,
				// it's because the schema needs to be created
				if util.ErrorCheck(err) {
					// Build the schema for the management database
					management.BuildSchema(conf)

					return cli.NewExitError("Management Database Setup completed successfully.", 0)
				}
				return cli.NewExitError("Management Database Setup Failed: Couldn't create Management Database Schema.", 1)
			}

			if ctx.IsSet("existing") {

				const YES, NO = "yes", "no"
				action := NO

				// Read configuration and access the management database
				conf, err := configureManagement()

				if err != nil {
					return cli.NewExitError("Configuration Load failed.", 1)
				}

				// Read the MySQL Database and generate Tables
				mysql.Setup(conf)
				err = mysql.ReadTables()
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

				if !util.ErrorCheckf(err, "There was a determining how to proceed. Cancelling setup.") {
					if action == YES {

						path := util.WorkingSubDir(conf.Project.Name)

						// Generate PropertyIds for all Database properties
						for i := 0; i < len(mysql.Schema); i++ {
							tbl := &mysql.Schema[i]
							tbl.GeneratePropertyIDs()

							// Generate YAML from the Tables and write to the working folder
							yaml.WriteTable(path, *tbl)

							tbl.InsertMetadata()
							util.LogInfof("Registering Table for migrations: %s", mysql.Schema[i].Name)
						}
					}
				}
			}

			return cli.NewExitError("Setup completed successfully.", 0)
		},
	}
	return setup
}
