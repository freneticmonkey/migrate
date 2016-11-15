package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/configsetup"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/git"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/urfave/cli"
)

// GetExecCommand Execute the migration indicated by the id flag
func GetExecCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "exec",
		Usage: "Apply a migration.",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "id",
				Value: 0,
				Usage: "The id of the migration to be applied",
			},
			cli.StringFlag{
				Name:  "gitinfo",
				Value: "",
				Usage: "Provide a git info file for the version of the target migration",
			},
			cli.BoolFlag{
				Name:  "print",
				Usage: "Print the Migration and its steps",
			},
			cli.BoolFlag{
				Name:  "dryrun",
				Usage: "Perform a dry run of the migration",
			},
			cli.BoolFlag{
				Name:  "rollback",
				Usage: "Allows for a rollback to be created",
			},
			cli.BoolFlag{
				Name:  "pto-disabled",
				Usage: "Execute the migration without using pt-online-schema-change.",
			},
			cli.BoolFlag{
				Name:  "allow-destructive",
				Usage: "Explictly allow rename and delete migration actions",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			var mid int64
			var info string
			var version string
			var ts string
			var m *migration.Migration

			// Parse global flags
			parseGlobalFlags(ctx)

			// Setup the management database and configuration settings
			_, err = configsetup.ConfigureManagement()

			// Check for gitinfo flag
			if ctx.IsSet("gitinfo") {
				info = ctx.String("gitinfo")
			}


			// Parse gitinfo and setup if valid
			if info != "" {
				version, info, ts, err = git.GetVersionDetailsFile(info)

				if err != nil {
					return cli.NewExitError("Migration failed. Unable to parse gitinfo file", 1)
				}

				util.LogInfof("Detected gitinfo file. Parsed:\nVersion: %s\nTime: %s\nInfo:\n>>>\n%s\n<<<", version, ts, info)

				m, err = migration.LoadVersion(version)

				if err != nil {
					return cli.NewExitError(fmt.Sprintf("Migration failed. Unable to find migration in database matching Version: [%s]", version), 1)
				}

				mid = m.MID

			} else {

				// Else check the command line flags for a migration id
				if ctx.IsSet("id") && ctx.Int("id") > 0 {
					mid = int64(ctx.Int("id"))
				} else {
					cli.ShowSubcommandHelp(ctx)
					return cli.NewExitError("Migration failed. Unable to execute a migration without a Migration Id", 1)
				}
			}


			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", err), 1)
			}

			if ctx.IsSet("print") {

				err = migration.Print(mid)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("Couldn't print migration: [%d] Error: %v", mid, err), 1)
				}

				return cli.NewExitError("Migration printed.", 0)
			}

			dryrun := ctx.Bool("dryrun")
			rollback := ctx.Bool("rollback")
			PTODisabled := ctx.Bool("pto-disabled")
			allowDestructive := ctx.Bool("allow-destructive")

			err = exec.Exec(exec.Options{
				MID:              mid,
				Dryrun:           dryrun,
				Rollback:         rollback,
				PTODisabled:      PTODisabled,
				AllowDestructive: allowDestructive,
			})

			if util.ErrorCheck(err) {
				errmsg := fmt.Sprintf("Execute failed. Unable to execute new Migration with ID: [%d].", mid)
				return cli.NewExitError(errmsg, 1)
			}

			success := fmt.Sprintf("Migration successfully with ID: %d", mid)
			return cli.NewExitError(success, 0)
		},
	}
	return setup
}
