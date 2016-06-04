package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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

func configureManagement(ctx *cli.Context) (err error) {

	configURL := "config.yml"

	if ctx.GlobalIsSet("config-url") {
		configURL = ctx.GlobalString("config-url")
		util.LogInfof("Detected config-url: %s", configURL)
	}

	if strings.HasPrefix(configURL, "http") {
		var response *http.Response

		// Download the configuration
		response, err = http.Get(configURL)

		// If the request was successfull
		if err == nil {
			// Read the response body
			var data []byte
			defer response.Body.Close()
			data, err = ioutil.ReadAll(response.Body)

			if !util.ErrorCheckf(err, "Problem reading the response for the config-url request") {
				// Unmarshal the YAML config
				err = yaml.ReadData(data, &conf)
			}
		}

	} else {
		// Assume that it's a local file
		err = yaml.ReadFile(configURL, &conf)
	}

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
