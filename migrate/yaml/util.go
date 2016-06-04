package yaml

import (
	"io/ioutil"

	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"

	"gopkg.in/yaml.v2"
)

func ReadFile(file string, out interface{}) (err error) {

	data, err := ioutil.ReadFile(file)

	util.ErrorCheck(err)

	err = yaml.Unmarshal(data, out)

	util.ErrorCheck(err)

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

	err = ioutil.WriteFile(file, filedata, 0644)

	util.ErrorCheck(err)

	return err
}
