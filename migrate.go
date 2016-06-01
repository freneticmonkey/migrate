package main

import (
	"flag"
	"os"

	"github.com/urfave/cli"

	"github.com/freneticmonkey/migrate/migrate/cmd"
	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/management"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/yaml"
)

var conf config.Config

var flags struct {
	config string
}

//
func processFlags() {
	flags.config = "config.yml"
	flag.Parse()
}

// Read Configuration
func readConfig() {
	err := yaml.ReadFile(flags.config, &conf)
	util.ErrorCheck(err)
	util.LogInfo("Configuration Read Success: " + flags.config)

	// Initialise any utility configuration
	util.Config(conf)
}

func main() {

	app := cli.NewApp()
	app.Name = "migrate"
	app.Usage = "Migrate MySQL databases using a YAML defined target schema"
	app.Author = "Scott Porter"
	app.Copyright = "MIT"
	app.Email = "scottporter@neuroticstudios.com"
	app.Version = "0.0.1"

	// Configure the app

	app.Flags = cmd.GetGlobalFlags()

	app.Commands = []cli.Command{
		cmd.GetSetupCommand(),
		cmd.GetSandboxCommand(&conf),
		cmd.GetDiffCommand(&conf),
		cmd.GetValidateCommand(&conf),
		cmd.GetCreateCommand(&conf),
		cmd.GetExecCommand(),
	}
	app.Before = func(ctx *cli.Context) error {

		configURL := ctx.String("config-url")
		err := yaml.ReadFile(configURL, &conf)
		if !util.ErrorCheckf(err, "Configuration read failed for: %s", configURL) {
			util.LogInfo("Configuration Read Success: " + configURL)

			// Initialise any utility configuration
			util.Config(conf)

			// Configure access to the management DB
			management.Setup(conf)

		} else {
			return cli.NewExitError("Unable to read configuration: "+configURL, 1)
		}
		return nil
	}

	app.Run(os.Args)
}
