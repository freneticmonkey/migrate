package cmd

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/serve"
	"github.com/freneticmonkey/migrate/go/util"

	"github.com/urfave/cli"
)

// GetServeCommand Configure the serve command
func GetServeCommand() (srv cli.Command) {
	srv = cli.Command{
		Name:  "serve",
		Usage: "Start a Migration REST server",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "frontend",
				Usage: "REST API only",
			},
			cli.IntFlag{
				Name:  "port",
				Value: 8081,
				Usage: "Server host port",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			var conf config.Config
			// Process command line flags

			// Parse global flags
			parseGlobalFlags(ctx)

			// frontend by default
			frontend := ctx.IsSet("frontend")

			port := ctx.Int("port")

			// Setup the management database and configuration settings
			conf, err = configureManagement()

			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Configuration Load failed. Error: %v", err), 1)
			}

			err = serve.Run(conf, frontend, port)

			if util.ErrorCheck(err) {
				return cli.NewExitError("Server Error", 1)
			}

			return cli.NewExitError("Server shutdown", 0)
		},
	}
	return srv
}
