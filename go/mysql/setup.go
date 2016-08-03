package mysql

import (
	"database/sql"

	"github.com/freneticmonkey/migrate/go/config"
)

var projectDB *sql.DB
var projectDBConn string
var projectDBID int

// SetProjectDB Directly set the Project DB Connection.  For unit testing.
func SetProjectDB(pdb *sql.DB) {
	projectDB = pdb
}

// Setup Configure the Project DB Connection
func Setup(conf config.Config, dbID int) {
	projectDBConn = conf.Project.DB.ConnectString()
	projectDBID = dbID
}

func connectProjectDB() (pdb *sql.DB, err error) {

	if projectDB == nil {
		projectDB, err = sql.Open("mysql", projectDBConn)
	}
	pdb = projectDB

	return pdb, err
}
