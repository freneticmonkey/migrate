package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/config"
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

func RecreateFolder(path string) (err error) {

	return err
}
