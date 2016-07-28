package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/exec"
	"github.com/freneticmonkey/migrate/migrate/util"
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
			cli.BoolFlag{
				Name:  "dryrun",
				Usage: "Peform a dry run of the migration",
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
		Action: func(ctx *cli.Context) error {
			var mid int64

			if ctx.IsSet("id") && ctx.Int("id") > 0 {
				mid = int64(ctx.Int("id"))
			} else {
				return cli.NewExitError("Exec failed. No Migration Id set", 1)
			}

			dryrun := ctx.Bool("dryrun")
			rollback := ctx.Bool("rollback")
			PTODisabled := ctx.Bool("pto-disabled")
			allowDestructive := ctx.Bool("allow-destructive")

			err := exec.Exec(exec.Options{
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
