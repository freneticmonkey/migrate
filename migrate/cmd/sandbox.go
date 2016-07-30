package cmd

import (
	"fmt"
	"time"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/exec"
	"github.com/freneticmonkey/migrate/migrate/id"
	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/mysql"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
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
		},
		Action: func(ctx *cli.Context) (err error) {

			// Parse global flags
			parseGlobalFlags(ctx)

			// Process command line flags
			return sandboxProcessFlags(ctx.Bool("recreate"), ctx.Bool("migrate"), ctx.Bool("dryrun"), ctx.Bool("force"))
		},
	}
	return setup
}

// sandboxProcessFlags Setup the Sandbox operation
func sandboxProcessFlags(recreate, migrate, dryrun, force bool) (err error) {
	var successmsg string

	const YES, NO = "yes", "no"
	action := NO

	if migrate || recreate {

		// Setup the management database and configuration settings
		configureManagement()

		if conf.Project.DB.Environment != "SANDBOX" && !force {
			return cli.NewExitError("Configured database isn't SANDBOX. Halting. If required use the force option.", 1)
		}

		if migrate {
			// If performing a migration

			successmsg, err = sandboxAction(&conf, dryrun, false, "Sandbox Migration")
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
				successmsg, err = sandboxAction(&conf, dryrun, true, "Sandbox Recreation")
				if util.ErrorCheck(err) {
					return cli.NewExitError(err.Error(), 1)
				}
				return cli.NewExitError(successmsg, 0)
			}
		}
		return cli.NewExitError("Sandbox Recreation cancelled.", 0)

	}
	return cli.NewExitError("No known parameters supplied.  Please refer to help for sandbox options.", 1)
}

func sandboxAction(conf *config.Config, dryrun bool, recreate bool, actionTitle string) (successmsg string, err error) {
	util.LogInfo(actionTitle)

	// Kick off a migration to recreate the db

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

	// If a clean migration
	if recreate {
		// Recreate the Sandbox Database
		recreateProjectDatabase(conf, dryrun)
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

func diffSchema(conf *config.Config, actionTitle string, recreate bool) (forwardOps mysql.SQLOperations, backwardOps mysql.SQLOperations, err error) {

	// Read the YAML schema
	err = yaml.ReadTables(conf.Project.LocalSchema.Path)
	if util.ErrorCheck(err) {
		err = fmt.Errorf("%s failed. Unable to read YAML Tables", actionTitle)
	}

	// If schema was found
	if len(yaml.Schema) > 0 {
		// Validate the YAML Schema
		_, err = id.ValidateSchema(yaml.Schema, "YAML Schema")
		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. YAML Validation Errors Detected", actionTitle)
		}

		// Read the MySQL tables from the target database
		err = mysql.ReadTables(conf.Project)
		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. Unable to read MySQL Tables", actionTitle)
		}

		// Don't bother validating the database if we're going to wipe it.
		if !recreate {
			// Validate the MySQL Schema
			_, err = id.ValidateSchema(mysql.Schema, "Target Database Schema")
			if util.ErrorCheck(err) {
				err = fmt.Errorf("%s failed. Target Database Validation Errors Detected", actionTitle)
			}
		}

		// Generate Diffs
		forwardDiff := table.DiffTables(yaml.Schema, mysql.Schema)
		forwardOps = mysql.GenerateAlters(forwardDiff)

		backwardDiff := table.DiffTables(mysql.Schema, yaml.Schema)
		backwardOps = mysql.GenerateAlters(backwardDiff)
	}

	return forwardOps, backwardOps, err
}

func createMigration(conf *config.Config, actionTitle string, dryrun bool, forwardOps mysql.SQLOperations, backwardOps mysql.SQLOperations) (m migration.Migration, err error) {
	if !dryrun {
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
		}

		util.LogInfof("%s: Created Migration with ID: %d", actionTitle, m.MID)

	} else {
		util.LogInfof("(DRYRUN) Skipping Creating Migration")
	}

	return m, err
}

func recreateProjectDatabase(conf *config.Config, dryrun bool) (err error) {
	var output string

	dropCommand := fmt.Sprintf("DROP DATABASE `%s`", conf.Project.DB.Database)
	createCommand := fmt.Sprintf("CREATE DATABASE `%s`", conf.Project.DB.Database)

	util.LogInfo(formatMessage(dryrun, "Sandbox Recreation", "Recreating Database"))
	if !dryrun {
		output, err = exec.ExecuteSQL(dropCommand, false)
		if util.ErrorCheckf(err, "Problem dropping DATABASE for Project: [%s] SQL: [%s] Output: [%s]", conf.Project.Name, dropCommand, output) {
			return cli.NewExitError("Sandbox Recreation failed. Couldn't DROP Project Database", 1)
		}

		output, err = exec.ExecuteSQL(createCommand, false)
		if util.ErrorCheckf(err, "Problem creating DATABASE for Project: [%s] SQL: [%s] Output: [%s]", conf.Project.Name, createCommand, output) {
			return cli.NewExitError("Sandbox Recreation failed. Couldn't Create Project Database", 1)
		}

		// Force a Reconnect to the database because the DB was just recreated
		exec.ConnectProjectDB(true)

	} else {
		util.LogInfof("(DRYRUN) Exec SQL: %s", dropCommand)
		util.LogInfof("(DRYRUN) Exec SQL: %s", createCommand)
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
