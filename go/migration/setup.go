package migration

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap
var projectDBID int

// Setup Setup the migration tables in the management DB
func Setup(db *gorp.DbMap, projectDatabaseID int) {
	mgmtDb = db
	projectDBID = projectDatabaseID

	if mgmtDb != nil {
		// Configure the Metadata table
		table := mgmtDb.AddTableWithName(Migration{}, "migration").SetKeys(true, "MID")
		table.ColMap("Timestamp").SetTransient(true)
		mgmtDb.AddTableWithName(Step{}, "migration_steps").SetKeys(true, "SID")
	}
}

// CreateTables Create the table for the Migration table as it needs some
// specific handling for the time related columns than go-gorp can current handle.
func CreateTables() (result bool, err error) {
	result = false

	createTable := []string{
		"CREATE TABLE `migration` (",
		"  `mid` bigint(20) NOT NULL AUTO_INCREMENT,",
		"  `db` int(11) NOT NULL,",
		"  `project` varchar(255) NOT NULL,",
		"  `version` varchar(255) NOT NULL,",
		"  `version_timestamp` datetime NOT NULL,",
		"  `version_description` text,",
		"  `status` int(11) NOT NULL,",
		"  `timestamp` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,",
		"  PRIMARY KEY (`mid`)",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}
	statement := strings.Join(createTable, "\n")

	// Execute the migration
	_, err = mgmtDb.Exec(statement)

	if !util.ErrorCheckf(err, "Problem creating Migrations table in the management DB") {

		// Now Create the migration steps table
		createTable = []string{
			"CREATE TABLE `migration_steps` (",
			"  `sid` bigint(20) NOT NULL AUTO_INCREMENT,",
			"  `mid` bigint(20) DEFAULT NULL,",
			"  `op` int(11) DEFAULT NULL,",
			"  `mdid` bigint(20) DEFAULT NULL,",
			"  `name` varchar(255) DEFAULT NULL,",
			"  `forward` varchar(255) DEFAULT NULL,",
			"  `backward` varchar(255) DEFAULT NULL,",
			"  `output` text,",
			"  `status` int(11) DEFAULT NULL,",
			"  PRIMARY KEY (`sid`)",
			") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		}
		statement := strings.Join(createTable, "\n")

		// Execute the migration
		_, err = mgmtDb.Exec(statement)
	}

	return result, err
}

// configured Internal Helper function for checking database validity
func configured() error {
	if mgmtDb != nil && mgmtDb.Db != nil && projectDBID > 0 {
		return nil
	}
	return fmt.Errorf("Metadata: Database not configured.")
}
