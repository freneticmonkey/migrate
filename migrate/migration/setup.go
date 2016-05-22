package migration

import (
	"strings"

	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap
var projectDBID int

// Setup Setup the migration tables in the management DB
func Setup(db *gorp.DbMap, projectDatabaseID int) {
	mgmtDb = db
	projectDBID = projectDatabaseID

	// Configure the Metadata table
	table := mgmtDb.AddTableWithName(Migration{}, "migration").SetKeys(true, "MID")
	table.ColMap("Timestamp").SetTransient(true)
	mgmtDb.AddTableWithName(Step{}, "migration_steps").SetKeys(true, "SID")

}

// CreateTables If tables need to be created, management.Setup will call here
// first
func CreateTables() {

	CreateMigrationTable()
}

// CreateMigrationTable Create the table for the Migration table as it needs some
// specific handling for the time related columns than go-gorp can current handle.
func CreateMigrationTable() (result bool, err error) {

	createTable := []string{
		"CREATE TABLE IF NOT EXISTS `migration` (",
		"`mid` bigint(20) NOT NULL AUTO_INCREMENT,",
		"`db` int(11) NOT NULL,",
		"`project` varchar(255) NOT NULL,",
		"`version` varchar(255) NOT NULL,",
		"`version_timestamp` DATETIME NOT NULL,",
		"`version_description` text,",
		"`status` int(11) NOT NULL,",
		"`timestamp` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,",
		"PRIMARY KEY (`mid`)",
		") ENGINE=InnoDB AUTO_INCREMENT=63 DEFAULT CHARSET=utf8",
	}
	statement := strings.Join(createTable, "\n")

	// Execute the migration
	_, err = mgmtDb.Exec(statement)

	result = util.ErrorCheckf(err, "Problem creating Migrations table in the management DB")

	return result, err
}
