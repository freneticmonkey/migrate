package cmd

import (
	"github.com/freneticmonkey/migrate/go/configsetup"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/urfave/cli"
)

// var conf config.Config
var configURL string
var configFile string

// GetGlobalFlags Configures the global flags used by all subcommands
func GetGlobalFlags() (flags []cli.Flag) {

	flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-url",
			Value: "",
			Usage: "URL for remote configuration.  If supplied config-file is ignored.",
		},
		cli.StringFlag{
			Name:  "config-file",
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

func parseGlobalFlags(ctx *cli.Context) {
	// Verbose output for now
	verbose := false
	configFile = ctx.GlobalString("config-file")

	if ctx.GlobalIsSet("config-file") {
		util.LogInfof("Detected config-file: %s", configFile)
	}
	configsetup.SetConfigFile(configFile)

	if ctx.GlobalIsSet("config-url") {
		configURL = ctx.GlobalString("config-url")
		util.LogInfof("Detected config-url: %s", configURL)
		configsetup.SetConfigURL(configURL)
	}

	if ctx.GlobalIsSet("verbose") {
		verbose = ctx.GlobalBool("verbose")
		util.LogInfof("Detected verbose: %t", verbose)
	}
	util.SetVerbose(verbose)
}
