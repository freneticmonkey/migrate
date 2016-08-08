package yaml

import (
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"

	"gopkg.in/yaml.v2"
)

func ReadFile(file string, out interface{}) (err error) {

	data, err := util.ReadFile(file)

	util.ErrorCheck(err)

	err = ReadData(data, out)

	return err

}

func ReadData(data []byte, out interface{}) (err error) {

	err = yaml.Unmarshal(data, out)

	util.ErrorCheck(err)

	return err

}

func WriteFile(file string, tbl table.Table) (err error) {

	tbl.RemoveNamespace()

	filedata, err := yaml.Marshal(tbl)

	util.ErrorCheck(err)

	err = util.WriteFile(file, filedata, 0644)

	util.ErrorCheck(err)

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

	pk := t.PrimaryIndex
	t.PrimaryIndex.Metadata = metadata.Metadata{
		PropertyID: pk.ID,
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
