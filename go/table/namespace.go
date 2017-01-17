package table

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/util"
)

// Namespace Stores the namespacing metadata for the Table
type Namespace struct {
	SchemaName    string
	TablePrefix   string
	SchemaPath    string
	GenPath       string
	TableName     string
	TableFilename string
}

// NewNamespace Initialise a new Namespace
func NewNamespace(ns *config.SchemaNamespace, tableName string) (tableNS Namespace) {

	if ns != nil {
		tableName = strings.TrimPrefix(tableName, ns.TablePrefix)
		tableNS = Namespace{
			SchemaName:    ns.Name,
			TablePrefix:   ns.TablePrefix,
			SchemaPath:    ns.SchemaPath,
			GenPath:       ns.GenPath,
			TableName:     tableName,
			TableFilename: tableName,
		}
	} else {
		tableNS = Namespace{
			TableName:     tableName,
			TableFilename: tableName,
		}
	}
	return tableNS

}

// SetTableFilename Sets the expected table filename format ahead of searching for existing files.
func (tn *Namespace) SetTableFilename(fileformat string) {
	tn.TableFilename = strings.Replace(fileformat, "<table>", tn.TableName, 1)
}

// SetExistingFilename Search the files parameter for a file matching the Table
// Will match a file either namespaced, or not.  For example:
//
// For DB table:
// 		animals_dogs
//
// Where SchemaNamespace:
// 		SchemaNamespace {
//			SchemaName:  "Animals",
//			TablePrefix: "animals",
//			SchemaPath:  "animals",
//			TableName:   "dogs",
//		}
//
// Namespaced:
// <root>/<schemapath>/dogs.txt
// Regular:
// <root>/dogs.txt
// Will both match.
func (tn *Namespace) SetExistingFilename(files []string, template config.Template) {

	// tnl := strings.ToLower(tn.TableFilename)
	// tnpath := strings.ToLower(tn.SchemaPath)

	workingPathPieces := strings.Split(util.WorkingPathAbs, fmt.Sprintf("%c", os.PathSeparator))
	workingIndex := len(workingPathPieces) - 1

	for _, file := range files {
		pathPieces := strings.Split(path.Dir(file), fmt.Sprintf("%c", os.PathSeparator))

		// extract the filename without the extension
		f := strings.ToLower(filepath.Base(file))
		fn := strings.TrimSuffix(f, filepath.Ext(f))
		match := false

		// check if the file is in a folder
		if len(pathPieces) > workingIndex+1 {

			// Get the file's path relative to the working path
			fileSubDir := strings.Join(pathPieces[workingIndex+1:], fmt.Sprintf("%c", os.PathSeparator))

			// if the file's path matches the namespace's schema path
			if fileSubDir == tn.GenPath {
				fullNamespace := tn.TablePrefix + tn.TableName
				noNamespace := tn.TableName

				fullNamespace = strings.ToLower(strings.Replace(template.FileFormat, "<table>", fullNamespace, 1))
				noNamespace = strings.ToLower(strings.Replace(template.FileFormat, "<table>", noNamespace, 1))

				// Try to match the file on the full namespaced tablename
				if fn == fullNamespace{
					match = true
				} else if fn == noNamespace {
					// else try to match the file on its un-namespaces name
					match = true
				}
			} else {
				continue
			}
		}

		// If the lowercase filename matches the lowercase tablename
		// if strings.ToLower(fn) == tnl {
		if match {
			// set the filename
			tn.TableFilename = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			// stop searching
			break
		}
	}
}

// GenerateSchemaFilename Generate a filename for the table given its namespace and a file extension
func (tn Namespace) GenerateSchemaFilename(ext string) string {
	path := []string{}
	if tn.SchemaPath != "" {
		path = strings.Split(tn.SchemaPath, fmt.Sprintf("%c", os.PathSeparator))
	}
	if tn.TableFilename != "" {
		file := tn.TableFilename + "." + ext
		if tn.TablePrefix != "" {
			file = tn.TablePrefix + file
		}
		path = append(path, file)
	}
	genpath := filepath.Join(path...)

	return genpath
}

// GenerateGenFilename Generate a filename for the table given its namespace and a file extension
func (tn Namespace) GenerateGenFilename(ext string) string {
	path := []string{}
	if tn.GenPath != "" {
		path = strings.Split(tn.GenPath, fmt.Sprintf("%c", os.PathSeparator))
	}
	if tn.TableFilename == "" {
		file := tn.TableFilename + "." + ext
		if tn.TablePrefix != "" {
			file = tn.TablePrefix + file
		}
		path = append(path, file)
	} else {
		path = append(path, tn.TableFilename + "." + ext)
	}
	genpath := filepath.Join(path...)

	return genpath
}
