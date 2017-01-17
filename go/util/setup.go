package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/robertkowalski/graylog-golang"
	"github.com/spf13/afero"
)

// WorkingPathAbs This is determined using the current working directory and
// the value of Config.Options.WorkingPath
var WorkingPathAbs string

var isTesting bool
var fsConfigured bool

var fs afero.Fs
var sh ShellRunner

// GetShell Get the shell command object
func GetShell() ShellRunner {
	return sh
}

// SetConfigTesting Enable unit test file and command subsystems
func SetConfigTesting() {
	isTesting = true
}

// configGrayLog Configure the Gray Log Driver
func configGrayLog(conf config.GrayLog) {
	gelfDriver = gelf.New(gelf.Config{
		GraylogPort:     conf.Port,
		GraylogHostname: conf.Hostname,
		Connection:      conf.Connection,
		MaxChunkSizeWan: conf.MaxChunkSizeWan,
		MaxChunkSizeLan: conf.MaxChunkSizeLan,
	})

	headers := []string{}

	for _, param := range conf.Parameters {
		headers = append(headers, fmt.Sprintf("\t\"%s\":\"%s\"", param.Name, param.Value))
	}

	gelfMessageFormat = `{
		` + strings.Join(headers, ",\n") + `,
		"message":"%s"
	}`
}

func ConfigFileSystem() {
	if !fsConfigured {
		// Configure the file system depending on whether we are running unit tests
		if !isTesting {
			fs = afero.NewOsFs()
			sh = &ShellExecutor{}
		} else {
			fs = afero.NewMemMapFs()
			sh = &MockShellExecutor{}
		}
		fsConfigured = true
	}

}

func ShutdownFileSystem() {
	if isTesting {
		fs = nil
		fsConfigured = false
	}
}

// Config Configure the utility subsystems depending on testing
func Config(conf config.Config) afero.Fs {
	var err error

	// If Graylog configured
	if conf.Options.GrayLog.Hostname != "" {
		configGrayLog(conf.Options.GrayLog)
	}

	rootPath := ""

	if conf.ConfigURL == "" {
		// The Working Path is Relative to the config file
		rootPath = filepath.Dir(conf.ConfigFile)

	} else {
		// Config URL is active, or defaulting to config.yml being loaded from CWD.

		rootPath, err = os.Getwd()
		if err != nil {
			LogErrorf("Problem extracting the CWD while configuring the Working Directory")
		}
	}

	// Configure the working path, ensuring the it's lowercase
	WorkingPathAbs = filepath.Join(rootPath, strings.ToLower(conf.Options.WorkingPath))
	WorkingPathAbs, err = filepath.Abs(WorkingPathAbs)

	LogInfof("Set Working Path: %s", WorkingPathAbs)

	if err != nil {
		LogErrorf("Problem configuring the Working Directory to: %s", WorkingPathAbs)
	}

	// Ensure that the filesystem has been setup correctly
	ConfigFileSystem()

	return fs
}
