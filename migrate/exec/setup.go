package exec

import (
	"database/sql"

	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap
var projectDB *gorp.DbMap
var projectConnectionDetails string
var projectDBID int

// Setup Setup the migration tables in the management DB
func Setup(db *gorp.DbMap, projectDatabaseID int, projectConnDetails string) {
	mgmtDb = db
	projectDBID = projectDatabaseID
	projectConnectionDetails = projectConnDetails
}

func connectProjectDB() (result bool, err error) {
	// The connection is already open
	if projectDB != nil {
		result = true
	} else {
		// Open the connection to the project DB
		var db *sql.DB
		db, err = sql.Open("mysql", projectConnectionDetails)
		if !util.ErrorCheckf(err, "Failed to connect to the management DB") {
			projectDB = &gorp.DbMap{
				Db: db,
				Dialect: gorp.MySQLDialect{
					Engine:   "InnoDB",
					Encoding: "UTF8",
				},
			}
			result = true
		}
	}
	return result, err
}
