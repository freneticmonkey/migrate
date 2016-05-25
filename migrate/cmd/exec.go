package cmd

import "github.com/urfave/cli"

// GetExecCommand Execute the migration indicated by the id flag
func GetExecCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "exec",
		Usage: "Migrations created by the create are executed by this subcommand. Migrations are identified by an id.",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "id",
				Value: -1,
				Usage: "The id of the migration to execute",
			},
			cli.BoolFlag{
				Name:  "dryrun",
				Usage: "Execute a dryrun of the migration",
			},
			cli.BoolFlag{
				Name:  "rollback",
				Usage: "Allow for a rollback migration to be executed.",
			},
		},
		Action: func(c *cli.Context) error {

			return cli.NewExitError("Migration Exec completed successfully.", 0)
		},
	}
	return setup
}
