package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/management"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
	"github.com/urfave/cli"
)

var conf config.Config

// GetGlobalFlags Configures the global flags used by all subcommands
func GetGlobalFlags() (flags []cli.Flag) {

	flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-url",
			Value: "config.yml",
			Usage: "URL for remote configuration.",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable verbose logging output",
		},
	}

	return flags
}

func configureManagement(ctx *cli.Context) error {

	configURL := "config.yml"
	if ctx.IsSet("config-url") {
		configURL = ctx.String("config-url")
	}
	err := yaml.ReadFile(configURL, &conf)
	if !util.ErrorCheckf(err, "Configuration read failed for: %s", configURL) {
		util.LogInfo("Configuration Read Success: " + configURL)

		// Initialise any utility configuration
		util.Config(conf)

		// Configure access to the management DB
		err = management.Setup(conf)
		if util.ErrorCheck(err) {
			return fmt.Errorf("Unable configure management database")
		}

	} else {
		return fmt.Errorf("Unable to read configuration: [%s]", configURL)
	}
	return nil
}
