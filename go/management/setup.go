package management

import (
	"database/sql"
	"fmt"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/database"
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
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
	if err != nil {
		return false
	}
	return len(dbTables) == len(tables)
}

// SetManagementDB Used to set a configured gorp.DbMap so that Unit Tests
// can control management database access
func SetManagementDB(db *gorp.DbMap) {
	mgmtDb = db
}

func getManagementDB(conf config.Config) *gorp.DbMap {
	mgmt := conf.Options.Management

	db, err := sql.Open("mysql", mgmt.DB.ConnectString())
	if util.ErrorCheckf(err, "Failed to connect to the management DB") {
		return nil
	}

	return &gorp.DbMap{
		Db: db,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		},
	}
}

// Setup Setup the database access to the Management DB
func Setup(conf config.Config) (err error) {

	// If the management DB hasn't already been setup, set it up now
	// The idea is that Unit Tests can set it up before this function is called.
	if mgmtDb == nil {
		mgmtDb = getManagementDB(conf)
	}

	if !tablesExist() {
		return fmt.Errorf("Cannot connect to Management database")
	}

	// Configure the Database Table packages
	database.Setup(mgmtDb)

	// Check if the target database exists, and if it doesn't, create an entry for it.
	var tdb database.TargetDatabase
	projDB := conf.Project.DB
	tdb, err = database.GetbyProject(conf.Project.Name, projDB.Database, projDB.Environment)

	if err != nil {
		// if util.ErrorCheckf(err, "Target Database entry doesn't exist for Project: [%s]. Creating it", conf.Project.Name) {
		util.LogWarnf("Target Database entry doesn't exist for Project: [%s]", conf.Project.Name)
		util.LogGreen("Creating it")
		tdb = database.TargetDatabase{
			Project: conf.Project.Name,
			Name:    projDB.Database,
			Env:     projDB.Environment,
		}
		err = tdb.Insert()

		util.ErrorCheckf(err, "Couldn't Insert the Target Database for Project: [%s] with Name: [%s]", conf.Project.Name, conf.Project.DB.Database)
	}

	if err == nil {
		yaml.Setup(conf)
		mysql.Setup(conf)
		metadata.Setup(mgmtDb, tdb.DBID)
		migration.Setup(mgmtDb, tdb.DBID)
		exec.Setup(mgmtDb, tdb.DBID, conf.Project.DB.ConnectString())
		util.LogInfo("Connected to Management DB")
	}

	return err
}

// BuildSchema Create the tables in the management database
func BuildSchema(conf config.Config) (err error) {

	// If the management DB hasn't already been setup, set it up now
	// The idea is that Unit Tests can set it up before this function is called.
	if mgmtDb == nil {
		mgmtDb = getManagementDB(conf)
	}

	if mgmtDb != nil && !tablesExist() {

		// Configure the Database Table packages
		database.Setup(mgmtDb)

		// Using a placeholder for the TargetDatabase ID as it's not needed for the management database schema creation
		metadata.Setup(mgmtDb, 0)
		migration.Setup(mgmtDb, 0)

		// If the Tables haven't been created, create them now.
		_, err = metadata.CreateTables()
		if util.ErrorCheckf(err, "Failed to create Metadata table in the management DB") {
			return err
		}
		_, err = migration.CreateTables()
		if util.ErrorCheckf(err, "Failed to create Migration and Migration Steps tables in the management DB") {
			return err
		}

		_, err = database.CreateTables()
		if util.ErrorCheckf(err, "Failed to create Database table in the management DB") {
			return err
		}

		util.LogInfo("Successfully Created Management database schema.")

	} else {
		err = fmt.Errorf("Existing Management Schema Detected.  Schema creation cancelled.")
	}
	return err
}
