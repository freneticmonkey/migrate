package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/metadata"
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
			cli.BoolFlag{
				Name:  "check-config",
				Usage: "Check environment and configuration",
			},
		},
		Action: func(ctx *cli.Context) *cli.ExitError {

			if !ctx.IsSet("management") && !ctx.IsSet("existing") && !ctx.IsSet("check-config") {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("Please specify a setup action", 1)
			}

			// Parse global flags
			parseGlobalFlags(ctx)

			if ctx.IsSet("management") {
				util.ConfigFileSystem()
				// Load Configuration only
				conf, configError := loadConfig(configURL, configFile)

				if configError != nil {
					return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", configError), 1)
				}

				// Build the schema for the management database
				err := management.BuildSchema(conf)

				if util.ErrorCheck(err) {
					return cli.NewExitError(fmt.Sprintf("Management Database Setup: Building tables FAILED with error: %v", err), 1)
				}
				return cli.NewExitError("Management Database Setup completed successfully.", 0)

				// return cli.NewExitError("Management Database Setup: Management DB is already setup", 1)

			} else if ctx.IsSet("existing") {
				// Read configuration and access the management database
				conf, configError := configureManagement()

				if configError != nil {
					return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", configError), 1)
				}
				return setupExistingDB(conf)

			} else if ctx.IsSet("check-config") {
				return checkConfig()
			}

			return cli.NewExitError("No action performed.", 0)
		},
	}
	return setup
}

func setupExistingDB(conf config.Config) *cli.ExitError {

	const YES, NO = "yes", "no"
	action := NO
	util.VerboseOverrideSet(true)
	util.LogInfo("Starting Setup from Existing DB")
	util.VerboseOverrideRestore()

	metadata.UseCache(true)

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

			util.VerboseOverrideSet(true)
			util.LogInfof("Detected %d Tables. Converting to YAML.", len(mysql.Schema))
			util.LogInfof("Writing to Path: %s", path)
			util.VerboseOverrideRestore()

			exists, err := util.DirExists(path)

			if err != nil {
				return cli.NewExitError("Couldn't create project folder: "+path, 1)
			}

			if !exists {
				util.Mkdir(path, 0755)
			}

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

			util.VerboseOverrideSet(true)
			util.LogOkf("Processed %d Tables", len(mysql.Schema))
			util.LogOkf("Generated YAML definitions in path: %s", path)
			return cli.NewExitError("Existing Database Setup Completed", 0)

		}
	}

	return cli.NewExitError("Management Database Setup Failed: Invalid option.", 1)
}

func checkConfig() *cli.ExitError {
	// Check Configuration
	valid := true

	// We need to see the output here
	util.SetVerbose(true)

	// Configuration Load
	util.ConfigFileSystem()

	// Load Configuration only
	conf, configError := loadConfig(configURL, configFile)

	if configError != nil {
		return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", configError), 1)
	}

	// Validate Configuration
	if conf.Options.WorkingPath == "" {
		util.LogError("Working Path: MISSING")
		valid = false
	} else {
		util.LogOk("Working Path: OK")
	}

	// Managment DB
	mgmtDBOk := true
	if conf.Options.Management.DB.Username == "" {
		util.LogError("Management DB Username: MISSING")
		mgmtDBOk = false
	}
	if conf.Options.Management.DB.Password == "" {
		util.LogError("Management DB Password: MISSING")
		mgmtDBOk = false
	}
	if conf.Options.Management.DB.Ip == "" {
		util.LogError("Management DB Ip: MISSING")
		mgmtDBOk = false
	}
	if conf.Options.Management.DB.Port == 0 {
		util.LogError("Management DB Port: MISSING")
		mgmtDBOk = false
	}
	if conf.Options.Management.DB.Database == "" {
		util.LogError("Management DB Database Name: MISSING")
		mgmtDBOk = false
	}

	// Check Management DB access
	if mgmtDBOk {
		util.LogOk("Management DB Configuration: OK")

		mgmtDB, err := sql.Open("mysql", conf.Options.Management.DB.ConnectString())

		if err != nil {
			util.LogErrorf("Management DB Connection: Couldn't Connect: %v", err)
			valid = false
		} else {
			mgmtDB.Close()
			util.LogOk("Management DB Connection: SUCCESS")
		}

	} else {
		valid = false
	}

	// Project
	if conf.Project.Name == "" {
		util.LogError("Project Name: MISSING")
		valid = false
	} else {
		util.LogOk("Project Name: OK")
	}

	// Target DB
	targetDBOk := true
	if conf.Project.DB.Username == "" {
		util.LogError("Target DB Username: MISSING")
		targetDBOk = false
	}
	if conf.Project.DB.Password == "" {
		util.LogError("Target DB Password: MISSING")
		targetDBOk = false
	}
	if conf.Project.DB.Ip == "" {
		util.LogError("Target DB Ip: MISSING")
		targetDBOk = false
	}
	if conf.Project.DB.Port == 0 {
		util.LogError("Target DB Port: MISSING")
		targetDBOk = false
	}
	if conf.Project.DB.Database == "" {
		util.LogError("Target DB Database Name: MISSING")
		targetDBOk = false
	}
	if conf.Project.DB.Environment == "" {
		util.LogError("Target DB Environment Name: MISSING")
		targetDBOk = false
	}

	// Check Target DB access
	if targetDBOk {
		util.LogOk("Target DB Configuration: OK")

		targetDB, err := sql.Open("mysql", conf.Project.DB.ConnectString())

		if err != nil {
			util.LogErrorf("Target DB Connection: Couldn't Connect: %v", err)
			valid = false
		} else {
			targetDB.Close()
			util.LogOk("Target DB Connection: SUCCESS")
		}
	} else {
		valid = false
	}

	// Check Git Configuration
	gitConfig := true
	if conf.Project.Schema.Url == "" {
		util.LogError("Schema Repo URL: MISSING")
		gitConfig = false
	}
	// All other Git Schema options are optional :)
	if gitConfig {
		util.LogOk("Project Schema Repo Config: OK")
	} else {
		valid = false
	}

	buildCheck := func(cmd string) []string {
		return []string{
			"-v",
			cmd,
			">/dev/null",
			"2>&1",
			"||",
			"{ echo >&2 \"I require foo but it's not installed.  Aborting.\"; exit 1; }",
		}
	}

	// Check for git install
	out, e := util.GetShell().Run("command", buildCheck("git")...)

	if e != nil {
		util.LogErrorf("Checking for Git: %v", e)
		valid = false
	} else {
		util.LogOk("Checking for Git: OK " + out)
	}

	// Check for pt-online-schema-change
	out, e = util.GetShell().Run("command", buildCheck("pt-online-schema-change")...)

	if e != nil {
		util.LogWarnf("Checking for PTO: FAILED: %v", e)
		valid = false
	} else {
		util.LogOk("Checking for PTO: OK Found: " + out)
	}

	if valid {
		return cli.NewExitError("Configuration Test Successful", 0)
	}
	return cli.NewExitError("Configuration Test Failed with Errors.", 1)
}
