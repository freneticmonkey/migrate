package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/go-gorp/gorp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

type SQLParseCTTest struct {
	CTStatement []string
	Expected    table.Table
	ExpectFail  bool
	Description string
}

func (t SQLParseCTTest) Statement() string {
	return strings.Join(t.CTStatement, "\n")
}

var parseCreateTableTests = []SQLParseCTTest{
	{
		CTStatement: []string{
			"CREATE TABLE `test` (",
			"`id` int(11) NOT NULL, ",
			"`name` varchar(64) NOT NULL, ",
			"PRIMARY KEY (`id`), ",
			"KEY `idx_id_name` (`id`,`name`)",
			") ENGINE=InnoDB DEFAULT CHARSET=latin1",
		},
		Expected: table.Table{
			Name:    "test",
			Engine:  "InnoDB",
			CharSet: "latin1",
			Columns: []table.Column{
				{
					Name: "id",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						Name:   "id",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "name",
					Type: "varchar",
					Size: []int{64},
					Metadata: metadata.Metadata{
						Name:   "name",
						Type:   "Column",
						Exists: true,
					},
				},
			},
			PrimaryIndex: table.Index{
				Name:      "PrimaryKey",
				Columns:   []string{"id"},
				IsPrimary: true,
				Metadata: metadata.Metadata{
					Name:   "PrimaryKey",
					Type:   "PrimaryKey",
					Exists: true,
				},
			},
			SecondaryIndexes: []table.Index{
				{
					Name: "idx_id_name",
					Columns: []string{
						"id",
						"name",
					},
					Metadata: metadata.Metadata{
						Name:   "idx_id_name",
						Type:   "Index",
						Exists: true,
					},
				},
			},
			Filename: "DB",
			Metadata: metadata.Metadata{
				Name:   "test",
				Type:   "Table",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Create Table: Basic Parse",
	},
}

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

func TestParseCreateTable(t *testing.T) {

	// Mock Database Setup
	db, err := dbSetup()
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed due to mock database setup with error: %v", err))
	}
	defer dbTearDown()

	// Configure metadata
	metadata.Setup(db, 1)

	for _, test := range parseCreateTableTests {

		query := fmt.Sprintf("SELECT count(*) from metadata WHERE name=\"%s\" and type=\"Table\"", test.Expected.Name)
		query = regexp.QuoteMeta(query)

		mock.ExpectQuery(query).
			WillReturnRows(sqlmock.NewRows([]string{
				"count",
			}).AddRow(0))

		result, err := ParseCreateTable(test.Statement())

		if err != nil || !reflect.DeepEqual(result, test.Expected) {

			// we make sure that all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Metadata was not queried for table name:%s Error: %s", test.Expected.Name, err)
			}

			t.Errorf("%s FAILED.", test.Description)
			if err != nil {
				util.LogWarnf("%s FAILED with error: %v", test.Description, err)
			}
			util.LogWarnf("%s FAILED.", test.Description)

		}
	}
}
