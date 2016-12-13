package util

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func WorkingSubDir(subDir string) string {
	return filepath.Join(WorkingPathAbs, subDir)
}

// Recursively check the path for typetype files and add them to the
// schema list as we're going
func ReadDirRelative(path string, fileType string, recursive bool, files *[]string) (err error) {

	// Make path absolute to the working folder
	path = filepath.Join(WorkingPathAbs, path)

	return ReadDirAbsolute(path, fileType, recursive, files)
}

func ReadDirAbsolute(path string, fileType string, recursive bool, files *[]string) (err error) {
	// Basic path check
	if _, err = fs.Stat(path); os.IsNotExist(err) {
		// This path is busted
		return err
	}
	dirInfo, err := afero.ReadDir(fs, path)
	if err != nil {
		return err
	}

	// Find files in this directory, and queue subdirectories
	subdirs := []string{}

	for _, fileinfo := range dirInfo {
		fp := filepath.Join(path, fileinfo.Name())

		if fileinfo.IsDir() {
			subdirs = append(subdirs, fp)
		} else {
			// Check if the file is a YAML file
			if strings.ToLower(filepath.Ext(fp)) == "."+fileType {
				alreadyFound := false

				// Verify file isn't already in the path first
				for _, f := range *files {
					if f == fp {
						alreadyFound = true
						break
					}
				}
				if !alreadyFound {
					*files = append(*files, fp)
				}
			}
		}
	}

	if recursive {
		// Process any subdirectories
		for _, dir := range subdirs {
			ReadDirAbsolute(dir, fileType, recursive, files)
		}
	}

	return err
}

func Stat(path string) (os.FileInfo, error) {
	return fs.Stat(path)
}

func FileExists(path string) (bool, error) {
	return afero.Exists(fs, path)
}

func DirExists(path string) (bool, error) {
	return afero.DirExists(fs, path)
}

func ReadAll(r io.Reader) ([]byte, error) {
	return afero.ReadAll(r)
}

func ReadFile(filename string) (data []byte, err error) {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(WorkingPathAbs, filename)
	}
	return afero.ReadFile(fs, filename)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(WorkingPathAbs, filename)
	}
	return afero.WriteFile(fs, filename, data, 0644)
}

func Open(filename string) (afero.File, error) {
	return fs.Open(filename)
}

func Create(filename string) (afero.File, error) {
	return fs.Create(filename)
}

func DeleteFile(filename string) error {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(WorkingPathAbs, filename)
	}
	return fs.Remove(filename)
}

func Mkdir(path string, perm os.FileMode) error {
	return fs.Mkdir(path, perm)
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

	fs.RemoveAll(path)

	return err
}

func RecreateFolder(path string) (err error) {

	return err
}
