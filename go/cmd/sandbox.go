package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/sandbox"
	"github.com/freneticmonkey/migrate/go/util"
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

			successmsg, err = sandbox.Action(conf, dryrun, false, "Sandbox Migration")
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
				successmsg, err = sandbox.Action(conf, dryrun, true, "Sandbox Recreation")
				if util.ErrorCheck(err) {
					return cli.NewExitError(err.Error(), 1)
				}
				return cli.NewExitError(successmsg, 0)
			}
		}
		return cli.NewExitError("Sandbox Recreation cancelled.", 0)

	} else if pulldiff {

		action, e := util.SelectAction("Are you sure you pull changes from MySQL to the YAML Schema? (This will override any unmigrated changes to the YAML)", []string{YES, NO})

		if util.ErrorCheck(e) {
			return cli.NewExitError("There was a problem confirming the pull-diff action.", 1)

		} else if action != YES {
			return cli.NewExitError("Pull-diff cancelled.", 0)
		}
		successmsg, err = sandbox.PullDiff(conf, pdTable)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Pull-diff FaILED: Error: %v", err), 1)
		}
		return cli.NewExitError("Pull-diff completed successfully.", 0)
	}
	return cli.NewExitError("No known parameters supplied.  Please refer to help for sandbox options.", 1)
}
