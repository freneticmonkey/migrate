package cmd

import (
	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/urfave/cli"
)

// GetSandboxCommand Configure the sandbox command
func GetSandboxCommand(conf *config.Config) (setup cli.Command) {
	setup = cli.Command{
		Name:  "sandbox",
		Usage: "Recreate the target database from the YAML Schema and insert the metadata",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force",
				Usage: "Force sandbox recreation.",
			},
		},
		Action: func(ctx *cli.Context) error {
			action, err := util.SelectAction("Are you sure you want to reset your sandbox?", []string{"yes", "no"})
			if util.ErrorCheck(err) {
				return cli.NewExitError("There was a problem confirming the action.", 1)
			}
			switch action {
			case "yes":
				{
					util.LogInfo("resetting sandbox")
				}
			case "no":
				{
					util.LogInfo("NOT resetting sandbox")
				}
			}
			if action == "yes" {
			}
			return cli.NewExitError("Setup completed successfully.", 0)
		},
	}
	return setup
}
