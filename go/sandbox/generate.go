package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/spf13/afero"
)

func stringSlice(input string, from, to int) string {
	return input[from:to]
}

func removeHungarian(input string) string {
	return strings.TrimLeft(input, "abcdefghijklnmopqrstuvxyz")
}

func underscoreDelimit(input string) string {
	pattern, err := regexp.Compile("(.)([A-Z][a-z])")
	if err != nil {
		util.LogErrorf("Regex Compile Error: %v", err)
	}
	return pattern.ReplaceAllString(input, "${1}_$2")
}

func scanGeneratedFiles(path string) (files []string, err error) {

	visit := func(path string, f os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	}

	err = filepath.Walk(path, visit)

	return files, err
}

// Generate Serialise the YAML Table(s) using a configuration defined template
func Generate(conf config.Config, tableName string) (err error) {

	err = yaml.ReadTables(conf)
	if util.ErrorCheck(err) {
		return fmt.Errorf("Generate failed. Unable to read YAML Tables")
	}

	// Filter by tableName in the YAML Schema
	if tableName != "*" {
		tgtTbl := []table.Table{}

		for _, tbl := range yaml.Schema {
			if tbl.Name == tableName {
				tgtTbl = append(tgtTbl, tbl)
				break
			}
		}
		// Reduce the YAML schema to the single target table
		yaml.Schema = tgtTbl
	}

	if len(yaml.Schema) == 0 {
		return fmt.Errorf("Generate: YAML Table not found: %s", tableName)
	}

	return writeTables(conf, yaml.Schema)
}

// GenerateTable Serialise a table using the configured template
func GenerateTable(conf config.Config, tbl table.Table) (err error) {

	return writeTables(conf, []table.Table{tbl})
}

func writeTables(conf config.Config, tables []table.Table) (err error) {
	var data []byte
	var tmpl *template.Template
	var exists bool
	var files []string

	util.SetVerbose(true)

	// Define the custom template functions
	funcMap := template.FuncMap{
		"content":           contentSlice,
		"removeHungarian":   removeHungarian,
		"toUpper":           strings.ToUpper,
		"trimSuffix":        strings.TrimSuffix,
		"replace":           strings.Replace,
		"title":             strings.Title,
		"contains":          strings.Contains,
		"slice":             stringSlice,
		"underscoreDelimit": underscoreDelimit,
	}

	for _, tmpOptions := range conf.Options.Generation.Templates {
		util.LogInfof("Generating Schema for Template: %s", tmpOptions.Name)

		if tmpOptions.Path == "" || tmpOptions.File == "" || tmpOptions.FileFormat == "" {
			return fmt.Errorf("Generate Table: Badly configured template options")
		}

		templateFile := util.WorkingSubDir(tmpOptions.File)

		// Generation path
		genPath := util.WorkingSubDir(tmpOptions.Path)

		// Ensure that the path exists
		exists, err = util.DirExists(genPath)
		if err != nil {
			return err
		}

		if !exists {
			err = util.Mkdir(genPath, 0755)
			if err != nil {
				return err
			}
		}

		data, err = util.ReadFile(templateFile)
		if err != nil {
			return err
		}
		tblTmpl := string(data)

		// Configure the table template
		tmpl, err = template.New(tmpOptions.Name).Funcs(funcMap).Parse(tblTmpl)
		if err != nil {
			return err
		}

		files, err = scanGeneratedFiles(genPath)

		// Generate each table
		for _, tbl := range tables {
			var f afero.File
			// Build a filename and path for the table file

			// Set the file format for the output file
			tbl.Namespace.SetTableFilename(tmpOptions.FileFormat)
			// Search for an existing file to adopt its naming case
			tbl.Namespace.SetExistingFilename(files)
			// Generate the final filename
			tblFilename := tbl.Namespace.GenerateFilename(tmpOptions.Ext)
			// Relative to the working directory
			tblFilename = filepath.Join(genPath, tblFilename)

			// Create Directory if not exists
			dir := filepath.Dir(tblFilename)
			exists, err = util.DirExists(dir)
			if err != nil {
				return err
			}
			if !exists {
				util.Mkdir(dir, 0755)
			}

			// If the file exists, read it and build any content sections
			exists, err = util.FileExists(tblFilename)
			if exists {
				data, err = util.ReadFile(tblFilename)
				if err != nil {
					return fmt.Errorf("Generate Table: Problem reading target file for replacement. Error: %v", err)
				}
				// Read the current file for content sections
				err = parseContent(string(data))
			}

			// Create a file
			f, err = util.Create(tblFilename)
			if err != nil {
				return err
			}

			// Generate the contents from the template
			err = tmpl.Execute(f, tbl)
			if err != nil {
				f.Close()
				return err
			}
			f.Close()
		}

	}

	return err

}
