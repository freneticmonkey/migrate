package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
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

	// Make path absolute
	cwd, err := os.Getwd()
	ErrorCheck(err)

	// Configure the working path, ensuring the it's lowercase
	WorkingPathAbs = filepath.Join(cwd, strings.ToLower(conf.Options.WorkingPath))
	WorkingPathAbs, err = filepath.Abs(WorkingPathAbs)

	if err != nil {
		LogErrorf("Problem configuring the Working Directory to: %s", WorkingPathAbs)
	}

	// Ensure that the filesystem has been setup correctly
	ConfigFileSystem()

	return fs
}
