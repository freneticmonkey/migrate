package table

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
)

// Namespace Stores the namespacing metadata for the Table
type Namespace struct {
	SchemaName    string
	TablePrefix   string
	Path          string
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
			Path:          ns.Path,
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
//			Path:        "animals",
//			TableName:   "dogs",
//		}
//
// Namespaced:
// <root>/<path>/dogs.txt
// Regular:
// <root>/dogs.txt
// Will both match.
func (tn *Namespace) SetExistingFilename(files []string) {

	tnl := strings.ToLower(tn.TableFilename)
	tnpath := strings.ToLower(tn.Path)
	for _, file := range files {
		// check if the file is in a folder
		pathPieces := strings.Split(path.Dir(file), fmt.Sprintf("%c", os.PathSeparator))

		if len(pathPieces) > 1 {
			dir := strings.ToLower(pathPieces[len(pathPieces)-1])

			// If the folder of the file doesn't match the SchemaNamespace short name,
			// ignore it.
			if dir != tnpath {
				continue
			}
			// If the folder is correct, keep checking the file
		}
		// extract the filename without the extension
		f := strings.ToLower(filepath.Base(file))
		fn := strings.TrimSuffix(f, filepath.Ext(f))

		// If the lowercase filename matches the lowercase tablename
		if strings.ToLower(fn) == tnl {
			// set the filename
			tn.TableFilename = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			// stop searching
			break
		}
	}
}

// GenerateFilename Generate a filename for the table given its namespace and a file extension
func (tn Namespace) GenerateFilename(ext string) string {
	path := []string{}
	if tn.Path != "" {
		path = strings.Split(tn.Path, fmt.Sprintf("%c", os.PathSeparator))
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
