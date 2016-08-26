package cmd

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/git"
	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/urfave/cli"
)

// GetDiffCommand Configure the validate command
func GetDiffCommand() (setup cli.Command) {
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
			cli.StringFlag{
				Name:  "table",
				Value: "",
				Usage: "Name of the target table to diff",
			},
		},
		Action: func(ctx *cli.Context) error {

			var version string
			var project string
			var table string

			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			conf, err := configureManagement()

			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", err), 1)
			}

			// Override the project settings with the command line flags
			if ctx.IsSet("version") {
				version = ctx.String("version")
			}

			if ctx.IsSet("project") {
				project = ctx.String("project")
			}

			if ctx.IsSet("table") {
				table = ctx.String("table")
			}

			return diff(version, project, table, conf)
		},
	}
	return setup
}

func diff(project, version, tableName string, conf config.Config) *cli.ExitError {

	var forwardDiff table.Differences
	var problems id.ValidationErrors
	var err error

	targetTableFound := false

	// Enable Metadata cache as we're not going to be making changes to it
	metadata.UseCache(true)

	// Override the project settings with the command line flags
	if version != "" {
		conf.Project.Schema.Version = version
	}

	if project != "" {
		conf.Project.Name = project
		git.Clone(conf.Project)
	} else {
		util.LogInfo("No project specified.  Comparing the current state of the YAML schema in working path.")
	}

	// Read the YAML files cloned from the repo
	err = yaml.ReadTables(strings.ToLower(conf.Project.Name))
	if util.ErrorCheck(err) {
		return cli.NewExitError("Diff failed. Unable to read YAML Tables", 1)
	}
	problems, err = id.ValidateSchema(yaml.Schema, "YAML Schema", true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. YAML Errors found", problems.Count())
	}

	// Filter by tableName in the YAML Schema
	if tableName != "" {
		tgtTbl := []table.Table{}

		for _, tbl := range yaml.Schema {
			if tbl.Name == tableName {
				tgtTbl = append(tgtTbl, tbl)
				targetTableFound = true
				break
			}
		}
		// Reduce the YAML schema to the single target table
		yaml.Schema = tgtTbl
	}

	// Read the MySQL tables from the target database
	err = mysql.ReadTables()
	if util.ErrorCheck(err) {
		return cli.NewExitError("Diff failed. Unable to read MySQL Tables", 1)
	}
	problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema", true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. Problems with Target Database Detected", problems.Count())
	}

	// Filter by tableName in the MySQL Schema
	if tableName != "" {
		tgtTbl := []table.Table{}

		for _, tbl := range mysql.Schema {
			if tbl.Name == tableName {
				tgtTbl = append(tgtTbl, tbl)
				targetTableFound = true
				break
			}
		}
		// Reduce the YAML schema to the single target table
		mysql.Schema = tgtTbl
	}

	// If a Table Name was specified, and a table wasn't found
	if !targetTableFound && tableName != "" {
		return cli.NewExitError(fmt.Sprintf("Diff failed for Table: %s. No found in YAML or MySQL Schemas", tableName), 1)
	}

	forwardDiff, err = table.DiffTables(yaml.Schema, mysql.Schema, true)
	if util.ErrorCheck(err) {
		return cli.NewExitError("Validation failed. Problems determining differences", 1)
	}
	util.VerboseOverrideSet(true)
	mysql.GenerateAlters(forwardDiff)
	util.VerboseOverrideRestore()

	completeMessage := "Diff completed successfully."

	if len(forwardDiff.Slice) > 0 {
		completeMessage += fmt.Sprintf(" %d differences found.", len(forwardDiff.Slice))
	} else {
		completeMessage += " No differences found."
	}

	return cli.NewExitError(completeMessage, 0)
}
