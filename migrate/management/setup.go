package management

import (
	"database/sql"
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/Database"
	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/exec"
	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap

func tablesExist() bool {
	var dbTables []string
	tables := []string{
		"metadata",
		"migration",
		"migration_steps",
		"target_database",
	}

	query := fmt.Sprintf("SHOW TABLES IN management")

	_, err := mgmtDb.Select(&dbTables, query)
	if util.ErrorCheck(err) {
		return false
	}
	return len(dbTables) == len(tables)
}

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

	if !tablesExist() {
		return fmt.Errorf("Unable to continue.  Management database is not initialised")
	}

	// Configure the Database Table packages
	database.Setup(mgmtDb)

	// Check if the target database exists, and if it doesn't, create an entry for it.
	var tdb database.TargetDatabase
	projDB := conf.Project.DB
	tdb, err = database.GetbyProject(conf.Project.Name, projDB.Database, projDB.Environment)
	if util.ErrorCheckf(err, "Target Database entry doesn't exist for Project: [%s]. Creating it", conf.Project.Name) {
		tdb = database.TargetDatabase{
			Project: conf.Project.Name,
			Name:    projDB.Database,
			Env:     projDB.Environment,
		}
		err = tdb.Insert()
	}

	if !util.ErrorCheckf(err, "Couldn't Insert the Target Database for Project: [%s] with Name: [%s]", conf.Project.Name, conf.Project.DB.Database) {
		metadata.Setup(mgmtDb, tdb.DBID)
		migration.Setup(mgmtDb, tdb.DBID)
		exec.Setup(mgmtDb, tdb.DBID, conf.Project.DB.ConnectString())
	}

	return err
}

// BuildSchema Create the tables in the management database
func BuildSchema(conf *config.Config) (err error) {

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

	if !tablesExist() {

		// Configure the Database Table packages
		database.Setup(mgmtDb)

		// Using a placeholder for the TargetDatabase ID as it's not needed for the management database schema creation
		metadata.Setup(mgmtDb, 0)
		migration.Setup(mgmtDb, 0)

		// If the Tables haven't been created, create them now.
		metadata.CreateTables()
		migration.CreateTables()

		err = mgmtDb.CreateTablesIfNotExists()
		if !util.ErrorCheckf(err, "Failed to create tables in the management DB") {
			util.LogInfo("Successfully Created Management database schema.")
		}
	} else {
		err = fmt.Errorf("Management Schema Detected.  Schema creation cancelled.")
	}
	return err
}
