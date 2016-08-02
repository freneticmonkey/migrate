package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/util/shell"
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
		"--format=%cI",
	})

	// 2006-01-02T15:04:05-07:00 		 - Example Git Time
	// Mon Jan 2 15:04:05 -0700 MST 2006 - Go Time format baseline
	// 2006-01-02T15:04:05Z07:00 		 - RFC339 Format (DB Format)
	// RFC3339 Format is used to store and compare DB version timestamps
	tm, err = time.Parse("2006-01-02T15:04:05-07:00", timestamp)
	formattedTime := tm.UTC().Format(mysql.TimeFormat)

	return formattedTime, err
}

// GetVersion Reads the Git repository in the working sub directory (project)
// and returns current commit version
func GetVersion(project string) (version string, err error) {
	path := util.WorkingSubDir(project)
	version, err = gitCmd(path, []string{
		"show",
		"-s",
		"--format=%H",
	})

	return version, err
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
func Clone(project config.Project) (err error) {
	path := util.WorkingSubDir(project.Name)

	// Cleanup the working path before doing any work
	util.CleanPath(path)

	// The steps to checkout

	// mkdir <working_dir>
	if _, err = os.Stat(path); os.IsNotExist(err) {
		util.LogInfof("Creating Path %s", path)
		err = os.Mkdir(path, 0755)
		util.ErrorCheckf(err, "Could not create git working folder: "+path)
	}

	// git init
	gitCmd(path, []string{"init"})

	// git remote add -f origin <url>
	gitCmd(path, []string{
		"remote",
		"add",
		"-f",
		"origin",
		project.Schema.Url,
	})

	// If folders have been specified for the repo
	if len(project.Schema.Folders) > 0 {

		// git config core.sparseCheckout true
		gitCmd(path, []string{
			"config",
			"core.sparseCheckout",
			"true",
		})

		var repoFolders []string

		// for each of the configured folders
		for _, folder := range project.Schema.Folders {
			// echo <repo_path>/*> .git/info/sparse-checkout
			repoFolders = append(repoFolders, folder+"/*")

		}
		filedata := strings.Join(repoFolders, "\n")

		// Write the folders into the git sparse-checkout info file
		sparseFile := fmt.Sprintf("%s/.git/info/sparse-checkout", path)
		err = ioutil.WriteFile(sparseFile, []byte(filedata), 0644)
	}

	// Build the checkout command
	params := []string{
		"checkout",
	}

	// If a version was specified append it to the checkout command
	if len(project.Schema.Version) > 0 {
		// git checkout <version>
		params = append(params, project.Schema.Version)
	} else {
		params = append(params, "master")
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
	}

	for _, piece := range cmd {
		params = append(params, piece)
	}
	util.LogInfof("Running git command: git %s", strings.Join(params, " "))
	out, err = shell.Run("git", "git:", params)
	util.ErrorCheckf(err, out)
	return out, err
}
