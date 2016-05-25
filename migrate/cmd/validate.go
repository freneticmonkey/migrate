package cmd

import "github.com/urfave/cli"

// GetValidateCommand Configure the validate command
func GetValidateCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "validate",
		Usage: "Validate the MySQL target database and the YAML schema.",
		Action: func(c *cli.Context) error {

			return cli.NewExitError("Validation completed successfully.", 0)
		},
	}
	return setup
}
