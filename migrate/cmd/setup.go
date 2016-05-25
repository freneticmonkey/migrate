package cmd

import "github.com/urfave/cli"

// GetSetupCommand Configure the setup command
func GetSetupCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "setup",
		Usage: "Setup the migration environment",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "sandbox",
				Usage: "Recreate the target database from the YAML Schema and insert the metadata",
			},
			cli.BoolFlag{
				Name:  "init-management",
				Usage: "Create the management tables in the management database",
			},
			cli.BoolFlag{
				Name:  "init-existing",
				Usage: "Read the target database and generate a YAML schema including PropertyIds",
			},
		},
		Action: func(c *cli.Context) error {

			return cli.NewExitError("Setup completed successfully.", 0)
		},
	}
	return setup
}
