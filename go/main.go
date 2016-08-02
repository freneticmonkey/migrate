package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/freneticmonkey/migrate/go/cmd"
)

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
		cmd.GetSandboxCommand(),
		cmd.GetDiffCommand(),
		cmd.GetValidateCommand(),
		cmd.GetCreateCommand(),
		cmd.GetExecCommand(),
		cmd.GetServeCommand(),
	}

	app.Run(os.Args)
}
