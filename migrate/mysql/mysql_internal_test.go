package mysql

import (
	"testing"

	"github.com/freneticmonkey/migrate/migrate/table"
)

var tblPropertyID = "testtbl"
var tblName = "test"

var colTests = []struct {
	ColStr   string       // Column definition to parse
	Expected table.Column // Expected Column defintion
}{
	{
		ColStr: "`name` varchar(64) NOT NULL",
		Expected: table.Column{
			Name:     "name",
			Type:     "varchar",
			Size:     64,
			Nullable: false,
			AutoInc:  false,
		},
	},
	{
		ColStr: "`age` int(11) NOT NULL",
		Expected: table.Column{
			Name:     "age",
			Type:     "int",
			Size:     11,
			Nullable: false,
			AutoInc:  false,
		},
	},
}

func TestColumnParse(t *testing.T) {
	var err error
	var colResult table.Column

	for _, colTest := range colTests {

		colResult, err = buildColumn(colTest.ColStr, tblPropertyID, tblName)

		if err != nil {
			t.Errorf("MySQL Column Parse Failed for column: '%s' with Error: '%s'", colTest.ColStr, err)
		} else {
			if hasDiff, diff := table.Compare(tblName, "TestColumn", colResult, colTest.Expected); hasDiff {
				t.Errorf("MySQL Column Parse Failed with Diff: '%s'", diff.Print())
			}
			// if !reflect.DeepEqual(colResult, colTest.Expected) {
			//
			// }
		}
	}

}
