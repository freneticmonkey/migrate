package management

import (
	"database/sql"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap

// Setup Setup the database access to the Management DB
func Setup(mgmt config.Management) (err error) {
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
	metadata.Setup(mgmtDb)
	migration.Setup(mgmtDb)

	// If the Tables haven't been created, create them now.
	err = mgmtDb.CreateTablesIfNotExists()
	util.ErrorCheckf(err, "Failed to create tables in the management DB")

	return err
}
