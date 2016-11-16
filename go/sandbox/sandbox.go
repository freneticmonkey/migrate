package sandbox

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
)

// Action Handle cli requests to manipulate the sandbox
func Action(conf config.Config, dryrun bool, recreate bool, actionTitle string) (successmsg string, err error) {
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
		Rollback:         false,
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
	err = yaml.ReadTables(conf)
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
			return forwardOps, backwardOps, err
		}

		// Read the MySQL tables from the target database
		err = mysql.ReadTables(conf)
		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. Unable to read MySQL Tables", actionTitle)
			return forwardOps, backwardOps, err
		}

		// Don't bother validating the database if we're going to wipe it.
		if !recreate {
			// Validate the MySQL Schema
			_, err = id.ValidateSchema(mysql.Schema, "Target Database Schema", true)
			if util.ErrorCheck(err) {
				err = fmt.Errorf("%s failed. Target Database Validation Errors Detected", actionTitle)
				return forwardOps, backwardOps, err
			}
		}

		_, err = id.ValidatePropertyIDs(yaml.Schema, mysql.Schema, true)
		if util.ErrorCheck(err) {
			err = fmt.Errorf("%s failed. YAML Validation Errors Detected", actionTitle)
			return forwardOps, backwardOps, err
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
				return fmt.Errorf("Sandbox Recreation failed. Couldn't DROP ALL TABLES for Project Database")
			}

			err = metadata.DeleteAllTargetDBMetadata()
			if util.ErrorCheckf(err, "Problem deleting all Metadata from Management Database") {
				return fmt.Errorf("Sandbox Recreation failed. Couldn't delete all Management Database")
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

// PullDiff Pull changes from the MySQL Schema back to the YAML.
func PullDiff(conf config.Config, tableName string) (result string, err error) {

	if len(tableName) == 0 {
		return "", fmt.Errorf("Pull-diff Error.  No table name supplied.  Use '*' to pull all tables.")
	}

	metadata.UseCache(true)

	// Read the MySQL tables from the target database
	err = mysql.ReadTables(conf)
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

	if len(mysql.Schema) == 0 {
		return "", fmt.Errorf("Generate: MySQL Table not found: %s", tableName)
	}

	// Serialise the MySQL Schema
	path := util.WorkingSubDir(strings.ToLower(conf.Project.Name))

	util.VerboseOverrideSet(true)
	util.LogInfof("Detected %d Tables. Converting to YAML.", len(mysql.Schema))
	util.LogInfof("Writing to Path: %s", path)
	util.VerboseOverrideRestore()

	exists, err := util.DirExists(path)

	if err != nil {
		return "", fmt.Errorf("Couldn't create project folder: " + path)
	}

	if !exists {
		util.Mkdir(path, 0755)
	}

	// TODO: Build list of Tables that will be effected by the pull-diff command.

	for i := 0; i < len(mysql.Schema); i++ {
		tbl := &mysql.Schema[i]

		// Generate PropertyIds for new fields in MySQL
		tbl.GeneratePropertyIDs()

		// Generate YAML from the Tables and write to the working folder
		err = yaml.WriteTable(path, *tbl)

		if err != nil {
			return "", fmt.Errorf("Existing Database Setup FAILED.  Unable to create YAML Table: %s due to error: %v", path, err)
		}

		// Insert the metadata for any new fields into MySQL
		err = tbl.InsertMetadata()
		if err != nil {
			return "", fmt.Errorf("Existing Database Setup FAILED.  Unable to insert metdata for Table: %s due to error: %v", tbl.Name, err)
		}
		util.LogInfof("Updating Table Registration for migrations: %s", mysql.Schema[i].Name)
	}

	result = "Successfully ran pull diff"

	return result, err
}

func formatMessage(dryrun bool, context string, message string, info ...interface{}) string {
	message = fmt.Sprintf("%s: %s", context, fmt.Sprintf(message, info...))
	if dryrun {
		message = "(DRYRUN) " + message
	}
	return message
}
