package yaml

import (
	"path/filepath"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"

	"gopkg.in/yaml.v2"
)

func ReadFile(file string, out interface{}) (err error) {

	data, err := util.ReadFile(file)

	if !util.ErrorCheckf(err, "Error Reading File: %s Error: %v", file, err) {
		err = ReadData(file, data, out)
	}

	return err

}

func ReadData(file string, data []byte, out interface{}) (err error) {
	err = yaml.Unmarshal(data, out)

	util.ErrorCheckf(err, "Error Unmarshalling File: %s Error: %v", file, err)

	return err
}

func WriteFile(file string, tbl table.Table) (err error) {
	var exists bool

	filedata, err := yaml.Marshal(tbl)
	if err != nil {
		return err
	}

	// Create Directory if not exists
	dir := filepath.Dir(file)
	exists, err = util.DirExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		util.Mkdir(dir, 0755)
	}

	if !util.ErrorCheck(err) {
		err = util.WriteFile(file, filedata, 0644)
	}

	return err
}

// Postprocess the loaded YAML table for it's Metadata
func processMetadata(t *table.Table) {
	t.Metadata = metadata.Metadata{
		PropertyID: t.ID,
		Name:       t.Name,
		Type:       "Table",
	}

	for i, col := range t.Columns {
		t.Columns[i].Metadata = metadata.Metadata{
			PropertyID: col.ID,
			ParentID:   t.ID,
			Name:       col.Name,
			Type:       "Column",
		}
	}

	// Set the name and id for the PrimaryIndex as it can only ever be the same value
	t.PrimaryIndex.Metadata = metadata.Metadata{
		PropertyID: "primarykey",
		ParentID:   t.ID,
		Name:       "PrimaryKey",
		Type:       "Index",
	}

	for i, index := range t.SecondaryIndexes {
		t.SecondaryIndexes[i].Metadata = metadata.Metadata{
			PropertyID: index.ID,
			ParentID:   t.ID,
			Name:       index.Name,
			Type:       "Index",
		}
	}
}
