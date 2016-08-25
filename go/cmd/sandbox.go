package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/urfave/cli"
)

// TODO: Need to reinsert metadata when inserting new tables.  This is not being done at all presently

// GetSandboxCommand Configure the sandbox command
func GetSandboxCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "sandbox",
		Usage: "Recreate the target database from the YAML Schema and insert the metadata",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "recreate",
				Usage: "Recreate the sandbox schema.",
			},
			cli.BoolFlag{
				Name:  "migrate",
				Usage: "Migrate sandbox database to current schema state",
			},
			cli.BoolFlag{
				Name:  "dryrun",
				Usage: "Perform a dryrun of the migration.",
			},
			cli.BoolFlag{
				Name:  "force",
				Usage: "Extremely Dangerous!!! Force the recreation of schema.",
			},
			cli.StringFlag{
				Name:  "pull-diff",
				Value: "",
				Usage: "Serialise manual MySQL Table alteration to YAML. Use '*' for entire schema.",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			var conf config.Config

			if !ctx.IsSet("recreate") && !ctx.IsSet("migrate") && !ctx.IsSet("pull-diff") {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("Please provide a valid flag", 1)
			}

			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			conf, err = configureManagement()

			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", err), 1)
			}

			// Process command line flags
			return sandboxProcessFlags(conf, ctx.Bool("recreate"), ctx.Bool("migrate"), ctx.Bool("dryrun"), ctx.Bool("force"), ctx.IsSet("pull-diff"), ctx.String("pull-diff"))
		},
	}
	return setup
}

// sandboxProcessFlags Setup the Sandbox operation
func sandboxProcessFlags(conf config.Config, recreate, migrate, dryrun, force, pulldiff bool, pdTable string) (err error) {
	var successmsg string

	const YES, NO = "yes", "no"
	action := ""

	if migrate || recreate {

		if conf.Project.DB.Environment != "SANDBOX" && !force {
			return cli.NewExitError("Configured database isn't SANDBOX. Halting. If required use the force option.", 1)
		}

		if migrate {
			// If performing a migration

			successmsg, err = sandboxAction(conf, dryrun, false, "Sandbox Migration")
			if util.ErrorCheck(err) {
				return cli.NewExitError(err.Error(), 1)
			}
			return cli.NewExitError(successmsg, 0)

		}

		// If recreating the sandbox database from scratch

		// If a dryrun, or being forced, don't prompt
		if dryrun || force {
			action = YES
		}

		if action == "" {
			action, err = util.SelectAction("Are you sure you want to reset your sandbox?", []string{YES, NO})
			if util.ErrorCheck(err) {
				return cli.NewExitError("There was a problem confirming the action.", 1)
			}
		}

		switch action {
		case YES:
			{
				successmsg, err = sandboxAction(conf, dryrun, true, "Sandbox Recreation")
				if util.ErrorCheck(err) {
					return cli.NewExitError(err.Error(), 1)
				}
				return cli.NewExitError(successmsg, 0)
			}
		}
		return cli.NewExitError("Sandbox Recreation cancelled.", 0)

	} else if pulldiff {
		return pullDiff(conf, pdTable)
	}
	return cli.NewExitError("No known parameters supplied.  Please refer to help for sandbox options.", 1)
}

func sandboxAction(conf config.Config, dryrun bool, recreate bool, actionTitle string) (successmsg string, err error) {
	util.LogInfo(actionTitle)

	// Kick off a migration to recreate the db

	// If a clean migration
	if recreate {
		// Recreate the Sandbox Database
		recreateProjectDatabase(conf, dryrun)
	}

	// Check that a local schema exists
	forwardOps, backwardOps, err := diffSchema(conf, actionTitle, recreate)

	if util.ErrorCheck(err) {
		return successmsg, err
	}

	// Create a local migration
	m, err := createMigration(conf, actionTitle, dryrun, forwardOps, backwardOps)

	if util.ErrorCheck(err) {
		return successmsg, err
	}

	// Apply the migration to the sandbox
	err = migrateSandbox(actionTitle, dryrun, &m)
	if util.ErrorCheck(err) {
		return successmsg, err
	}
	dr := ""
	if dryrun {
		dr = "(DRYRUN)"
	}
	successmsg = fmt.Sprintf("%s %s: Migration successfully with ID: %d", dr, actionTitle, m.MID)
	util.LogInfo(successmsg)

	return successmsg, err
}

func migrateSandbox(actionTitle string, dryrun bool, m *migration.Migration) (err error) {
	util.LogInfof(formatMessage(dryrun, actionTitle, "Applying Schema"))

	util.LogInfo(actionTitle)

	exec.Exec(exec.Options{
		Dryrun:           dryrun,
		Force:            true,
		Rollback:         true,
		PTODisabled:      true,
		AllowDestructive: true,
		Sandbox:          true,
		Migration:        m,
	})
	if util.ErrorCheck(err) {
		err = fmt.Errorf("%s: Execute failed. Unable to execute sandbox Migration with ID: [%d]", actionTitle, m.MID)
	}

	return err
}

func diffSchema(conf config.Config, actionTitle string, recreate bool) (forwardOps mysql.SQLOperations, backwardOps mysql.SQLOperations, err error) {
	util.LogInfo(formatMessage(false, "Sandbox Diff Schema", "Comparing MySQL and YAML Schemas"))

	var forwardDiff table.Differences
	var backwardDiff table.Differences
	// Read the YAML schema
	err = yaml.ReadTables(conf.Project.LocalSchema.Path)
	if util.ErrorCheck(err) {
		err = fmt.Errorf("%s failed. Unable to read YAML Tables", actionTitle)
	}

	// If schema was found
	if len(yaml.Schema) > 0 {

		metadata.UseCache(true)

		// Validate the YAML Schema
		_, err = id.ValidateSchema(yaml.Schema, "YAML Schema", true)
		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. YAML Validation Errors Detected", actionTitle)
		}

		// Read the MySQL tables from the target database
		err = mysql.ReadTables()
		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. Unable to read MySQL Tables", actionTitle)
		}

		// Don't bother validating the database if we're going to wipe it.
		if !recreate {
			// Validate the MySQL Schema
			_, err = id.ValidateSchema(mysql.Schema, "Target Database Schema", true)
			if util.ErrorCheck(err) {
				err = fmt.Errorf("%s failed. Target Database Validation Errors Detected", actionTitle)
			}
		}

		// Generate Diffs
		forwardDiff, err = table.DiffTables(yaml.Schema, mysql.Schema, false)
		if util.ErrorCheckf(err, "Diff Failed while generating forward migration") {
			return forwardOps, backwardOps, err
		}
		forwardOps = mysql.GenerateAlters(forwardDiff)

		backwardDiff, err = table.DiffTables(mysql.Schema, yaml.Schema, false)
		if util.ErrorCheckf(err, "Diff Failed while generating backward migration") {
			return forwardOps, backwardOps, err
		}
		backwardOps = mysql.GenerateAlters(backwardDiff)
	}

	return forwardOps, backwardOps, err
}

func createMigration(conf config.Config, actionTitle string, dryrun bool, forwardOps mysql.SQLOperations, backwardOps mysql.SQLOperations) (m migration.Migration, err error) {
	if !dryrun {
		util.LogInfo(formatMessage(dryrun, "Sandbox Create Migration", "Inserting new Migration into the DB"))
		// Create a temporary migration.  If there a way we can avoid this?
		m, err = migration.New(migration.Param{
			Project: conf.Project.Name,
			Version: conf.Project.Schema.Version,
			// Use the current state of the local Git repo (Don't do a git checkout )
			// Migration Database doesn't need to have any git info in it because this feature is for testing
			// migrations without having checked them in
			Timestamp:   time.Now().UTC().Format(mysql.TimeFormat),
			Description: actionTitle,
			Forwards:    forwardOps,
			Backwards:   backwardOps,
			Sandbox:     true,
		})

		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. Unable to create new Migration in the management database", actionTitle)
		} else {
			util.LogInfof("%s: Created Migration with ID: %d", actionTitle, m.MID)
		}

	} else {
		util.LogInfof("(DRYRUN) Skipping Creating Migration")
	}

	return m, err
}

func recreateProjectDatabase(conf config.Config, dryrun bool) (err error) {
	var output string
	var tables []string

	tables, err = mysql.ReadTableNames()

	util.LogInfo(formatMessage(dryrun, "Sandbox Recreation", "Recreating Database"))

	if len(tables) > 0 {
		dropTables := fmt.Sprintf("DROP TABLE `%s`", strings.Join(tables, "`,`"))

		util.LogInfo(formatMessage(dryrun, "Sandbox Recreation", "Recreating Database"))
		if !dryrun {
			output, err = exec.ExecuteSQL(dropTables, false)
			if util.ErrorCheckf(err, "Problem dropping ALL TABLES for Project: [%s] SQL: [%s] Output: [%s]", conf.Project.Name, dropTables, output) {
				return cli.NewExitError("Sandbox Recreation failed. Couldn't DROP ALL TABLES for Project Database", 1)
			}

			// Force a Reconnect to the database because the DB was just recreated
			// exec.ConnectProjectDB(true)

			// Emptying the MySQL Schema

			mysql.Schema = []table.Table{}
		} else {
			util.LogInfof("(DRYRUN) Exec SQL: %s", dropTables)
		}
	}

	return err
}

func pullDiff(conf config.Config, tableName string) (err error) {

	YES := "yes"
	NO := "no"

	if len(tableName) == 0 {
		return cli.NewExitError("Pull-diff Error.  No table name supplied.  Use '*' to pull all tables.", 1)
	}

	action, e := util.SelectAction("Are you sure you pull changes from MySQL to the YAML Schema? (This will override any unmigrated changes to the YAML)", []string{YES, NO})

	if util.ErrorCheck(e) {
		return cli.NewExitError("There was a problem confirming the action.", 1)

	} else if action != YES {
		return cli.NewExitError("Pull-diff cancelled.", 0)
	}

	metadata.UseCache(true)

	// Read the MySQL tables from the target database
	err = mysql.ReadTables()
	if util.ErrorCheck(err) {
		err = fmt.Errorf("Pull-Diff failed. Unable to read MySQL Tables")
	}

	// Filter by tableName in the MySQL Schema
	if tableName != "*" {
		tgtTbl := []table.Table{}

		for _, tbl := range mysql.Schema {
			if tbl.Name == tableName {
				tgtTbl = append(tgtTbl, tbl)
				break
			}
		}
		// Reduce the YAML schema to the single target table
		mysql.Schema = tgtTbl
	}

	// Serialise the MySQL Schema
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

	for i := 0; i < len(mysql.Schema); i++ {
		tbl := &mysql.Schema[i]

		// Generate PropertyIds for new fields in MySQL
		tbl.GeneratePropertyIDs()

		// Generate YAML from the Tables and write to the working folder
		err = yaml.WriteTable(path, *tbl)

		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Existing Database Setup FAILED.  Unable to create YAML Table: %s due to error: %v", path, err), 1)
		}

		// Insert the metadata for any new fields into MySQL
		err = tbl.InsertMetadata()
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Existing Database Setup FAILED.  Unable to insert metdata for Table: %s due to error: %v", tbl.Name, err), 1)
		}
		util.LogInfof("Updating Table Registration for migrations: %s", mysql.Schema[i].Name)
	}

	return err
}

func formatMessage(dryrun bool, context string, message string, info ...interface{}) string {
	message = fmt.Sprintf("%s: %s", context, fmt.Sprintf(message, info...))
	if dryrun {
		message = "(DRYRUN) " + message
	}
	return message
}
