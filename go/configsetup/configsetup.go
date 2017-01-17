package configsetup

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
)

var configFile string
var configURL string
var intConfig config.Config
var configCreated bool

// SetConfigFile Set the file used for local configuration
func SetConfigFile(cFile string) {
	configFile = cFile
}

// SetConfigURL Set the URL used for remote configuration
func SetConfigURL(cURL string) {
	configURL = cURL
}

// ConfigureManagement Load configuration and setup the mananagement database
func ConfigureManagement() (targetConfig config.Config, err error) {

	util.ConfigFileSystem()

	// Load Configuration
	targetConfig, err = LoadConfig(configURL, configFile)

	if err == nil {
		// Set Configuration
		// Initialise any utility configuration
		util.Config(targetConfig)

		// Configure access to the management DB
		err = management.Setup(targetConfig)

		if err != nil {
			err = fmt.Errorf("Unable configure management database. Error: %v", err)
		}

		intConfig = targetConfig
		configCreated = true
	}

	return targetConfig, err
}

// LoadConfig Load a configuration from URL and fallback to filepath if URL is not supplied.
// If the URL fails to return a valid configration an error is returned.
func LoadConfig(configURL, configFile string) (targetConfig config.Config, err error) {

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
				err = yaml.ReadData(configURL, data, &targetConfig)
				configSource = configURL
				targetConfig.ConfigURL = configURL
			}
		}

	} else {
		// Assume that it's a local file
		err = yaml.ReadFile(configFile, &targetConfig)
		configSource = configFile
		targetConfig.ConfigFile = configFile
	}

	if util.ErrorCheckf(err, "Configuration read failed for: %s", configSource) {
		return targetConfig, fmt.Errorf("Unable to read configuration from: [%s]", configSource)
	}

	util.LogInfo("Successfully read configuration from: " + configSource)

	return targetConfig, err
}

// CheckConfig Check the current configuration for issues
func CheckConfig(log bool) (checks Health) {
	var conf config.Config
	var configError error

	// Get a Health object
	checks = GetHealth()

	// We need to see the output here
	util.SetVerbose(true)

	// Configuration Load
	util.ConfigFileSystem()

	// Load Configuration only
	if !configCreated {
		conf, configError = LoadConfig(configURL, configFile)

		if configError != nil {
			checks.AddFail(fmt.Sprintf("Configuration Load failed. Error: %v", configError))
		}
	} else {
		conf = intConfig
	}

	if checks.Ok() {
		// Validate Configuration
		if conf.Options.WorkingPath == "" {
			checks.AddFail("Working Path: MISSING")
		} else {
			checks.AddPass("Working Path: OK")
		}

		// Managment DB
		mgmtDBOk := true
		if conf.Options.Management.DB.Username == "" {
			checks.AddFail("Management DB Username: MISSING")
			mgmtDBOk = false
		}
		if conf.Options.Management.DB.Password == "" {
			checks.AddFail("Management DB Password: MISSING")
			mgmtDBOk = false
		}
		if conf.Options.Management.DB.Ip == "" {
			checks.AddFail("Management DB Ip: MISSING")
			mgmtDBOk = false
		}
		if conf.Options.Management.DB.Port == 0 {
			checks.AddFail("Management DB Port: MISSING")
			mgmtDBOk = false
		}
		if conf.Options.Management.DB.Database == "" {
			checks.AddFail("Management DB Database Name: MISSING")
			mgmtDBOk = false
		}

		// Check Management DB access
		if mgmtDBOk {
			checks.AddPass("Management DB Configuration: OK")

			mgmtDB, err := sql.Open("mysql", conf.Options.Management.DB.ConnectString())

			if err != nil {
				checks.AddFail(fmt.Sprintf("Management DB Connection: Couldn't Connect: %v", err))
			} else {
				mgmtDB.Close()
				checks.AddPass("Management DB Connection: SUCCESS")
			}

		} else {
		}

		// Project
		if conf.Project.Name == "" {
			checks.AddFail("Project Name: MISSING")
		} else {
			checks.AddPass("Project Name: OK")
		}

		// Target DB
		targetDBOk := true
		if conf.Project.DB.Username == "" {
			checks.AddFail("Target DB Username: MISSING")
			targetDBOk = false
		}
		if conf.Project.DB.Password == "" {
			checks.AddFail("Target DB Password: MISSING")
			targetDBOk = false
		}
		if conf.Project.DB.Ip == "" {
			checks.AddFail("Target DB Ip: MISSING")
			targetDBOk = false
		}
		if conf.Project.DB.Port == 0 {
			checks.AddFail("Target DB Port: MISSING")
			targetDBOk = false
		}
		if conf.Project.DB.Database == "" {
			checks.AddFail("Target DB Database Name: MISSING")
			targetDBOk = false
		}
		if conf.Project.DB.Environment == "" {
			checks.AddFail("Target DB Environment Name: MISSING")
			targetDBOk = false
		}

		// Check Target DB access
		if targetDBOk {
			checks.AddPass("Target DB Configuration: OK")

			targetDB, err := sql.Open("mysql", conf.Project.DB.ConnectString())

			if err != nil {
				checks.AddFail(fmt.Sprintf("Target DB Connection: Couldn't Connect: %v", err))
			} else {
				targetDB.Close()
				checks.AddPass("Target DB Connection: SUCCESS")
			}
		} else {
			checks.AddFail("Target DB Configuration: BAD")
		}

		// Check Git Configuration
		gitConfig := true
		if conf.Project.Git.Url == "" {
			checks.AddFail("Git Repo URL: MISSING")
			gitConfig = false
		}
		// All other Git Schema options are optional :)
		if gitConfig {
			checks.AddPass("Project Git Repo Config: OK")
		} else {
			checks.AddFail("Git Configuration: BAD")
		}

		buildCheck := func(cmd string) []string {
			return []string{
				"-v",
				cmd,
				">/dev/null",
				"2>&1",
				"||",
				"{ echo >&2 \"I require foo but it's not installed.  Aborting.\"; exit 1; }",
			}
		}

		// Check for git install
		out, e := util.GetShell().Run("command", buildCheck("git")...)

		if e != nil {
			checks.AddFail(fmt.Sprintf("Checking for Git: %v", e))
		} else {
			checks.AddPass("Checking for Git: OK " + out)
		}

		// Check for pt-online-schema-change
		out, e = util.GetShell().Run("command", buildCheck("pt-online-schema-change")...)

		if e != nil {
			checks.AddFail(fmt.Sprintf("Checking for PTO: FAILED: %v", e))
		} else {
			checks.AddPass("Checking for PTO: OK Found: " + out)
		}
	}

	if log {
		checks.Display()
	}

	return checks

}
