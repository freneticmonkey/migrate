package cmd

import (
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
			// Process command line flags

			// Parse global flags
			parseGlobalFlags(ctx)

			// frontend by default
			frontend := ctx.IsSet("frontend")

			port := ctx.Int("port")

			// Setup the management database and configuration settings
			_, err = configureManagement()

			if err != nil {
				return cli.NewExitError("Configuration Load failed.", 1)
			}

			err = serve.Run(frontend, port)

			if util.ErrorCheck(err) {
				return cli.NewExitError("Server Error", 1)
			}

			return cli.NewExitError("Server shutdown", 0)
		},
	}
	return srv
}
