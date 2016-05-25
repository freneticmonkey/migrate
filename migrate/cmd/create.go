package cmd

import "github.com/urfave/cli"

// GetCreateCommand Create a new migration for the target project at the version indicated by hash
func GetCreateCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "create",
		Usage: "This subcommand is used to create a migration and register it with the management database.",
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
			cli.BoolFlag{
				Name:  "rollback",
				Usage: "Allows for a rollback to be created",
			},
			cli.BoolFlag{
				Name:  "force-sandbox",
				Usage: "Immediately apply the new migration to the target database. Will only function in the sandbox environment.",
			},
		},
		Action: func(c *cli.Context) error {

			return cli.NewExitError("Migration Create completed successfully.", 0)
		},
	}
	return setup
}
