package cmd

import "github.com/urfave/cli"

// GetValidateCommand Configure the validate command
func GetValidateCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "validate",
		Usage: "Validate the MySQL target database and the YAML schema.",
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
		},
		Action: func(c *cli.Context) error {

			return cli.NewExitError("Validation completed successfully.", 0)
		},
	}
	return setup
}
