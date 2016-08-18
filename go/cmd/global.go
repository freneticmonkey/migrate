package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
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

	if ctx.GlobalIsSet("config-url") {
		configURL = ctx.GlobalString("config-url")
		util.LogInfof("Detected config-url: %s", configURL)
	}

	if ctx.GlobalIsSet("verbose") {
		verbose = ctx.GlobalBool("verbose")
		util.LogInfof("Detected verbose: %t", verbose)
	}
	util.SetVerbose(verbose)
}

// configureManagement Read the command line parameters,
// load configuration and setup the mananagement database
func configureManagement() (targetConfig config.Config, err error) {

	util.ConfigFileSystem()

	// Load Configuration
	targetConfig, err = loadConfig(configURL, configFile)

	if err == nil {
		// Set Configuration
		err = setConfig(targetConfig)
	}

	return targetConfig, err
}

// loadConfig Load a configuration from URL and fallback to filepath if URL is not supplied.
// If the URL fails to return a valid configration an error is returned.
func loadConfig(configURL, configFile string) (targetConfig config.Config, err error) {

	var configSource string

	// If the ConfigURL is set and it's a http URL
	if strings.HasPrefix(configURL, "http") {
		var response *http.Response

		// Download the configuration
		response, err = http.Get(configURL)

		// If the request was successfull
		if err == nil {
			// Read the response body
			var data []byte
			defer response.Body.Close()
			data, err = util.ReadAll(response.Body)

			if !util.ErrorCheckf(err, "Problem reading the response for the config-url request") {
				// Unmarshal the YAML config
				err = yaml.ReadData(data, &targetConfig)
				configSource = configURL
			}
		}

	} else {
		// Assume that it's a local file
		err = yaml.ReadFile(configFile, &targetConfig)
		configSource = configFile
	}

	if util.ErrorCheckf(err, "Configuration read failed for: %s", configSource) {
		return targetConfig, fmt.Errorf("Unable to read configuration from: [%s]", configSource)
	}

	util.LogInfo("Successfully read configuration from: " + configSource)

	return targetConfig, err
}

// setConfig Initialise using the Config parameter
func setConfig(targetConfig config.Config) (err error) {

	// Initialise any utility configuration
	util.Config(targetConfig)

	// Configure access to the management DB
	err = management.Setup(targetConfig)

	if err != nil {
		return fmt.Errorf("Unable configure management database. Error: %v", err)
	}

	return nil
}
