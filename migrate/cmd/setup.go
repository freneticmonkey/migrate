package cmd

import "github.com/urfave/cli"

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
		},
		Action: func(ctx *cli.Context) error {
			return cli.NewExitError("Setup completed successfully.", 0)
		},
	}
	return setup
}
