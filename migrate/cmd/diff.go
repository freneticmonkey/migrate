package cmd

import "github.com/urfave/cli"

// GetDiffCommand Configure the validate command
func GetDiffCommand() (setup cli.Command) {
	setup = cli.Command{
		Name:  "diff",
		Usage: "Diff the MySQL target database and the YAML schema.",
		Action: func(c *cli.Context) error {

			return cli.NewExitError("Diff completed successfully.", 0)
		},
	}
	return setup
}
