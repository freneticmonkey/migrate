package sandbox

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
)

// Generate Serialise the YAML Table(s) using a configuration defined template
func Generate(conf config.Config, tableName string) (err error) {
	var data []byte
	var tmpl *template.Template
	var f *os.File
	var exists bool

	tmpOptions := conf.Options.Template

	if tmpOptions.Path == "" || tmpOptions.File == "" || tmpOptions.Ext == "" {
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

	tmpl, err = template.New("t").Parse(tblTmpl)
	if err != nil {
		return err
	}

	err = yaml.ReadTables(strings.ToLower(conf.Project.Name))
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

	// Generate each table
	for _, tbl := range yaml.Schema {
		// Build a filename and path for the table file
		tblFilename := fmt.Sprintf("%s.%s", strings.ToLower(tbl.Name), tmpOptions.Ext)

		// Create a file
		f, err = os.Create(filepath.Join(genPath, tblFilename))
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

	return err
}

// GenerateTable Serialise a table using the configured template
func GenerateTable(conf config.Config, tbl table.Table) (err error) {
	var data []byte
	var tmpl *template.Template
	var f *os.File
	var exists bool

	tmpOptions := conf.Options.Template

	if tmpOptions.Path == "" || tmpOptions.File == "" || tmpOptions.Ext == "" {
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

	tmpl, err = template.New("t").Parse(tblTmpl)
	if err != nil {
		return err
	}

	// Build a filename and path for the table file
	tblFilename := fmt.Sprintf("%s.%s", strings.ToLower(tbl.Name), tmpOptions.Ext)

	// Create a file
	f, err = os.Create(filepath.Join(genPath, tblFilename))
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

	return err
}
