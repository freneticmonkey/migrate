package exec

import (
	"database/sql"

	"github.com/freneticmonkey/migrate/go/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap
var projectDB *gorp.DbMap
var projectConnectionDetails string
var projectDBID int

// SetProjectDB Used to set a configured gorp.DbMap so that Unit Tests
// can control project database access
func SetProjectDB(pdb *gorp.DbMap) {
	projectDB = pdb
}

// Setup Setup the migration tables in the management DB
func Setup(db *gorp.DbMap, projectDatabaseID int, projectConnDetails string) {
	mgmtDb = db
	projectDBID = projectDatabaseID
	projectConnectionDetails = projectConnDetails
}

// ConnectProjectDB Setup the Database connection to the project database.
// If the reconnect parameter is true, then a reconnect will be forced.
// This is used when recreating the project database.
func ConnectProjectDB(reconnect bool) (result bool, err error) {

	if reconnect && projectDB != nil {
		projectDB.Db.Close()
		projectDB = nil
	}

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
