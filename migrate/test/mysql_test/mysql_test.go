package mysql_test

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/mysql"
)

var mockDb *sql.DB
var mock sqlmock.Sqlmock

// Configure Gorp with Mock DB
func dbSetup() (gdb *gorp.DbMap, err error) {

	mockDb, mock, err = sqlmock.New()

	if err != nil {
		return nil, err
	}

	gdb = &gorp.DbMap{
		Db: mockDb,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		},
	}

	return gdb, err
}

func dbTearDown() {
	mockDb.Close()
}

// Successful Read
func IgnoreTestRead(t *testing.T) {

	// Mock Database Setup
	db, err := dbSetup()
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed due to mock database setup with error: %v", err))
	}
	defer dbTearDown()

	// Configure metadata
	metadata.Setup(db, 1)

	definition := `
    CREATE TABLE ` + "`test`" + ` (
      ` + "`id`" + ` int(11) NOT NULL,
      ` + "`name`" + ` varchar(64) NOT NULL,
      PRIMARY KEY (` + "`id`" + `),
      KEY ` + "`idx_id_name`" + ` (` + "`id`" + `,` + "name`" + `)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1
    `

	tbl, err := mysql.ParseCreateTable(definition)

	if err != nil {
		t.Error(fmt.Sprintf("Parse Error: %v", err))
	}

	// Validate table properties
	if tbl.Name != "test" {
		t.Error(fmt.Sprintf("Table Name incorrect: Expected: 'test' Found: '%s'", tbl.Name))
	}

	if tbl.CharSet != "latin1" {
		t.Error(fmt.Sprintf("Table CharSet incorrect: Expected: 'latin1' Found: '%s'", tbl.CharSet))
	}

	if tbl.Engine != "InnoDB" {
		t.Error(fmt.Sprintf("Table Engine incorrect: Expected: 'InnoDB' Found: '%s'", tbl.Engine))
	}

	if tbl.ID != "tbl1" {
		t.Error(fmt.Sprintf("Table ID incorrect: Expected: 'tbl1' Found: '%s'", tbl.ID))
	}

	numCol := len(tbl.Columns)
	if numCol != 2 {
		t.Error(fmt.Sprintf("Table has invalid number of columns: Expected: 2 Found: %d", numCol))
	}

	// Validate Column Properties
	col := tbl.Columns[0]

	if col.Name != "id" {
		t.Error(fmt.Sprintf("Column Name incorrect: Expected: 'id' Found: '%s'", col.Name))
	}

	if col.Type != "int" {
		t.Error(fmt.Sprintf("Column Type incorrect: Expected: 'int' Found: '%s'", col.Type))
	}

	if col.Size[0] != 11 {
		t.Error(fmt.Sprintf("Column Size incorrect: Expected: '11' Found: '%d'", col.Size))
	}

	if col.Nullable != false {
		t.Error(fmt.Sprintf("Column Size incorrect: Expected: 'No' Found: '%t'", col.Nullable))
	}

	if col.ID != "col1" {
		t.Error(fmt.Sprintf("Column ID incorrect: Expected: 'col1' Found: '%s'", col.ID))
	}

	// Validate PrimaryKey Properties
	pi := tbl.PrimaryIndex

	if pi.Name != "" {
		t.Error(fmt.Sprintf("PrimaryKey Name incorrect: Expected: '' Found: '%s'", pi.Name))
	}

	if pi.ID != "pi" {
		t.Error(fmt.Sprintf("PrimaryKey ID incorrect: Expected: 'pi' Found: '%s'", pi.ID))
	}

	cols := []string{"name"}
	if !reflect.DeepEqual(pi.Columns, cols) {
		t.Error(fmt.Sprintf("PrimaryKey Columns incorrect: Expected: '%v' Found: '%v'", cols, pi.Columns))
	}

	numInd := len(tbl.SecondaryIndexes)
	if numInd != 1 {
		t.Error(fmt.Sprintf("Table has invalid number of indexes: Expected: 1 Found: %d", numInd))
	}

	si := tbl.SecondaryIndexes[0]

	if si.Name != "idx_id_name" {
		t.Error(fmt.Sprintf("Secondary Index Name incorrect: Expected: 'idx_id_name' Found: '%s'", si.Name))
	}

	if si.ID != "sc1" {
		t.Error(fmt.Sprintf("Secondary Index ID incorrect: Expected: 'sc1' Found: '%s'", si.ID))
	}

	cols = []string{"id", "name"}
	if !reflect.DeepEqual(si.Columns, cols) {
		t.Error(fmt.Sprintf("SecondaryIndexe Columns incorrect: Expected: '%v' Found: '%v'", cols, pi.Columns))
	}

}
