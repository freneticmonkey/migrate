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

	{
		CTStatement: []string{
			"CREATE TABLE `devicetierparameter` (",
			"`parameter_id` int(11) NOT NULL AUTO_INCREMENT,",
			"`game_id` int(11) NOT NULL,",
			"`name` varchar(255) NOT NULL,",
			"`type` int(11) NOT NULL,",
			"`default_value` varchar(255) NOT NULL,",
			"`order` int(11) NOT NULL,",
			"PRIMARY KEY (`parameter_id`),",
			"UNIQUE KEY `idx_game_id_name` (`game_id`,`name`),",
			"KEY `idx_game_id_parameter_id` (`game_id`,`parameter_id`)",
			") ENGINE=InnoDB AUTO_INCREMENT=1014",
		},
		Expected: table.Table{
			Name:    "devicetierparameter",
			Engine:  "InnoDB",
			AutoInc: 1014,
			Columns: []table.Column{
				{
					Name:    "parameter_id",
					Type:    "int",
					Size:    []int{11},
					AutoInc: true,
					Metadata: metadata.Metadata{
						Name:   "parameter_id",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "game_id",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						Name:   "game_id",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "name",
					Type: "varchar",
					Size: []int{255},
					Metadata: metadata.Metadata{
						Name:   "name",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "type",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						Name:   "type",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "default_value",
					Type: "varchar",
					Size: []int{255},
					Metadata: metadata.Metadata{
						Name:   "default_value",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "order",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						Name:   "order",
						Type:   "Column",
						Exists: true,
					},
				},
			},
			PrimaryIndex: table.Index{
				Name:      "PrimaryKey",
				Columns:   []string{"parameter_id"},
				IsPrimary: true,
				Metadata: metadata.Metadata{
					Name:   "PrimaryKey",
					Type:   "PrimaryKey",
					Exists: true,
				},
			},
			SecondaryIndexes: []table.Index{
				{
					Name:     "idx_game_id_name",
					IsUnique: true,
					Columns: []string{
						"game_id",
						"name",
					},
					Metadata: metadata.Metadata{
						Name:   "idx_game_id_name",
						Type:   "Index",
						Exists: true,
					},
				},
				{
					Name: "idx_game_id_parameter_id",
					Columns: []string{
						"game_id",
						"parameter_id",
					},
					Metadata: metadata.Metadata{
						Name:   "idx_game_id_parameter_id",
						Type:   "Index",
						Exists: true,
					},
				},
			},
			Filename: "DB",
			Metadata: metadata.Metadata{
				Name:   "devicetierparameter",
				Type:   "Table",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Create Table: Device Tier Parse",
	},

	{
		CTStatement: []string{
			"CREATE TABLE `storeproductfile` (",
			"`file_id` int(11) NOT NULL AUTO_INCREMENT,",
			"`game_id` int(11) NOT NULL,",
			"`file` longblob NOT NULL,",
			"`order` int(11) NOT NULL,",
			"PRIMARY KEY (`file_id`)",
			") ENGINE=InnoDB",
		},
		Expected: table.Table{
			Name:   "storeproductfile",
			Engine: "InnoDB",
			Columns: []table.Column{
				{
					Name:    "file_id",
					Type:    "int",
					Size:    []int{11},
					AutoInc: true,
					Metadata: metadata.Metadata{
						Name:   "file_id",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "game_id",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						Name:   "game_id",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "file",
					Type: "longblob",
					Metadata: metadata.Metadata{
						Name:   "file",
						Type:   "Column",
						Exists: true,
					},
				},
				{
					Name: "order",
					Type: "int",
					Size: []int{11},
					Metadata: metadata.Metadata{
						Name:   "order",
						Type:   "Column",
						Exists: true,
					},
				},
			},
			PrimaryIndex: table.Index{
				Name:      "PrimaryKey",
				Columns:   []string{"file_id"},
				IsPrimary: true,
				Metadata: metadata.Metadata{
					Name:   "PrimaryKey",
					Type:   "PrimaryKey",
					Exists: true,
				},
			},
			Filename: "DB",
			Metadata: metadata.Metadata{
				Name:   "storeproductfile",
				Type:   "Table",
				Exists: true,
			},
		},
		ExpectFail:  false,
		Description: "Create Table: StoreProducts Parse",
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

			context := ""
			if err != nil {
				util.LogWarnf("%s FAILED with error: %v", test.Description, err)
				context = "Errors while parsing CREATE TABLE statement"
			} else {
				util.LogWarnf("%s FAILED.", test.Description)
				context = "Parsed Table doesn't match"
				util.DebugDumpDiff(test.Expected, result)
				// util.LogAttention("Expected")
				// util.DebugDump(test.Expected)
				// util.LogWarn("Result")
				// util.DebugDump(result)
			}

			t.Errorf("%s FAILED. %s", test.Description, context)

		}
	}
}
