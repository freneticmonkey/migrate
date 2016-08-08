package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
)

// WorkingPathAbs This is determined using the current working directory and
// the value of Config.Options.WorkingPath
var WorkingPathAbs string

func Config(conf config.Config) {

	// Make path absolute
	cwd, err := os.Getwd()
	ErrorCheck(err)

	WorkingPathAbs = filepath.Join(cwd, conf.Options.WorkingPath)
}

func WorkingSubDir(subDir string) string {
	return filepath.Join(WorkingPathAbs, subDir)
}

// Recursively check the path for typetype files and add them to the
// schema list as we're going

func ReadDirRelative(path string, fileType string, files *[]string) (err error) {

	// Make path absolute
	cwd, err := os.Getwd()
	ErrorCheck(err)

	path = filepath.Join(cwd, path)

	return ReadDirAbsolute(path, fileType, files)
}

func ReadDirAbsolute(path string, fileType string, files *[]string) (err error) {

	// Basic path check
	if _, err = os.Stat(path); os.IsNotExist(err) {
		// This path is busted
		return err
	}

	dirInfo, err := ioutil.ReadDir(path)
	ErrorCheck(err)

	for _, fileinfo := range dirInfo {
		path := filepath.Join(path, fileinfo.Name())
		if fileinfo.IsDir() {
			ReadDirAbsolute(path, fileType, files)
		} else {
			// Check if the file is a YAML file
			if strings.ToLower(filepath.Ext(path)) == "."+fileType {
				*files = append(*files, path)
			}
		}
	}

	return err
}

func ReadFile(file string) (data []byte, err error) {

	return ioutil.ReadFile(file)

}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, 0644)
}

// cleanUp is a helper function which empties the target folder
func CleanPath(path string) (err error) {

	// Ensure that the path is within the working folder
	var rel string
	rel, err = filepath.Rel(WorkingPathAbs, path)
	if strings.HasPrefix(rel, "..") {
		return errors.New("Cannot clean paths outside of the working directory: Path: " + rel)
	}
	LogWarn("Cleaning Path: " + path)
	os.RemoveAll(path)

	return err
}

func RecreateFolder(path string) (err error) {

	return err
}
