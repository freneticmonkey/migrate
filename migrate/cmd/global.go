package cmd

import "github.com/urfave/cli"

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
