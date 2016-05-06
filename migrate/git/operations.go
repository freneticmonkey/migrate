package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/util/shell"
)

// GetVersionTime Reads the Git repository in the working sub directory (project)
// and returns the committer timestamp for (version) in RFC3339 format which
// should be compatible with MySQL timestamps.
func GetVersionTime(project string, version string) (timestamp string, err error) {
	var tm time.Time
	path := util.WorkingSubDir(project)
	timestamp, err = gitCmd(path, []string{
		"show",
		"-s",
		"format=%cI",
	})

	// 2016-05-06T08:28:17+10:00 - Example Git Time
	// Mon Jan 2 15:04:05 -0700 MST 2006 - Go Time format baseline
	tm, err = time.Parse("2006-01-02T15:04:05-07:00", timestamp)
	formattedTime := tm.UTC().Format(time.RFC3339)

	return formattedTime, err
}

// GetVersionDetails Reads the Git repository in the working sub directory (project)
// and returns the commit message, author details, and timestamp of the commit
func GetVersionDetails(project string, version string) (details string, err error) {
	path := util.WorkingSubDir(project)
	details, err = gitCmd(path, []string{
		"show",
		"-s",
		"--pretty=medium",
	})

	return details, err
}

// Clone performs a check out into a new (project) sub folder underneath WorkingPath
// If the project configuration specifies sub folders within the project
// repository, then a sparse checkout is performed for only the specified folders
func Clone(project string, schemaConf config.Schema) (err error) {
	path := util.WorkingSubDir(project)

	// Cleanup the working path before doing any work
	cleanUp(path)

	// The steps to checkout

	// mkdir <working_dir>
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 644)
		util.ErrorCheck(err)
	}

	// cd <working_dir>
	err = os.Chdir(path)
	util.ErrorCheck(err)

	// git init
	gitCmd(path, []string{"init"})

	// git remote add -f origin <url>
	gitCmd(path, []string{
		"remote",
		"add",
		"-f",
		"origin",
		schemaConf.Url,
	})

	// If folders have been specified for the repo
	if len(schemaConf.Folders) > 0 {

		// git config core.sparseCheckout true
		gitCmd(path, []string{
			"config",
			"core.sparseCheckout",
			"true",
		})

		var repoFolders []string

		// for each of the configured folders
		for _, folder := range schemaConf.Folders {
			// echo <repo_path>/*> .git/info/sparse-checkout
			repoFolders = append(repoFolders, folder+"/*")

		}
		filedata := strings.Join(repoFolders, "\n")

		// Write the folders into the git spare-checkout info file
		sparseFile := fmt.Sprintf("%s/.git/info/sparse-checkout", path)
		err = ioutil.WriteFile(sparseFile, []byte(filedata), 0644)
	}

	// Build the checkout command
	params := []string{
		"checkout",
	}

	// If a version was specified append it to the checkout command
	if len(schemaConf.Version) > 0 {
		// git checkout <version>
		params = append(params, schemaConf.Version)
	}

	// run the checkout
	gitCmd(path, params)

	return err
}

// gitCmd is a helper function for executing git commands within the project
// WorkingPath folder
func gitCmd(path string, cmd []string) (out string, err error) {
	params := []string{
		"-C",
		path,
		"init",
	}

	for _, piece := range cmd {
		params = append(params, piece)
	}
	out, err = shell.Run("git", "git:", params)
	util.ErrorCheckf(err, out)
	return out, err
}

// cleanUp is a helper function which empties the target folder
func cleanUp(path string) (err error) {
	// Build a path to the working folder
	// wd, err := os.Getwd()
	//
	// wp := fmt.Sprintf("%s/*", filepath.Join(wd, path) )
	//
	// LogWarn(wp)
	// os.RemoveAll(fmt.Sprintf("%s/*", wp))

	return err

}
