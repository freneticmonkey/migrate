package mysql

import (
	"database/sql"

	"github.com/freneticmonkey/migrate/migrate/config"
)

var projectDB *sql.DB
var projectDBConn string

func SetProjectDB(pdb *sql.DB) {
	projectDB = pdb
}

func Setup(conf config.Config) {
	projectDBConn = conf.Project.DB.ConnectString()
}

func connectProjectDB() (pdb *sql.DB, err error) {

	if projectDB == nil {
		projectDB, err = sql.Open("mysql", projectDBConn)
	}

	return pdb, err
}
