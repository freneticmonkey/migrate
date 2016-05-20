package management

import (
	"database/sql"

	"github.com/freneticmonkey/migrate/migrate/Database"
	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap

// Setup Setup the database access to the Management DB
func Setup(conf config.Config) (err error) {
	mgmt := conf.Management

	var db *sql.DB
	db, err = sql.Open("mysql", mgmt.DB.ConnectString())
	util.ErrorCheckf(err, "Failed to connect to the management DB")

	mgmtDb = &gorp.DbMap{
		Db: db,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		},
	}

	// Configure the Database Table packages
	database.Setup(mgmtDb)

	// Check if the target database exists, and if it doesn't, create an entry for it.
	var tdb database.TargetDatabase
	tdb, err = database.GetbyProject(conf.Project.Name, conf.Project.DB.Database)
	if util.ErrorCheckf(err, "Target Database entry doesn't exist for Project: [%s]. Creating it", conf.Project.Name) {
		err = mgmtDb.CreateTablesIfNotExists()
		tdb = database.TargetDatabase{
			Project: conf.Project.Name,
			Name:    conf.Project.DB.Database,
		}
		err = tdb.Insert()
	}

	if !util.ErrorCheckf(err, "Couldn't Insert the Target Database for Project: [%s] with Name: [%s]", conf.Project.Name, conf.Project.DB.Database) {
		metadata.Setup(mgmtDb, tdb.DBID)
		migration.Setup(mgmtDb, tdb.DBID, conf.Project.DB.ConnectString())

		// If the Tables haven't been created, create them now.
		err = mgmtDb.CreateTablesIfNotExists()
		util.ErrorCheckf(err, "Failed to create tables in the management DB")
	}

	return err
}
