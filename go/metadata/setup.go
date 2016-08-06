package metadata

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap
var targetDBID int

// Setup Setup the Metadata table in the management DB
func Setup(db *gorp.DbMap, targetDatabaseID int) {
	mgmtDb = db
	targetDBID = targetDatabaseID

	// Configure the Metadata table
	mgmtDb.AddTableWithName(Metadata{}, "metadata").SetKeys(true, "MDID")

}

// CreateTables Create the Metadata table
func CreateTables() (result bool, err error) {

	createTable := []string{
		"CREATE TABLE `metadata` (",
		"  `mdid` bigint(20) NOT NULL AUTO_INCREMENT,",
		"  `db` int(11) DEFAULT NULL,",
		"  `property_id` varchar(255) DEFAULT NULL,",
		"  `parent_id` varchar(255) DEFAULT NULL,",
		"  `type` varchar(255) DEFAULT NULL,",
		"  `name` varchar(255) DEFAULT NULL,",
		"  `exists` tinyint(1) DEFAULT NULL,",
		"  PRIMARY KEY (`mdid`)",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}
	statement := strings.Join(createTable, "")

	// Execute the migration
	_, err = mgmtDb.Exec(statement)

	result = util.ErrorCheckf(err, "Problem creating Database table in the management DB")

	return result, err
}

// configured Internal Helper function for checking database validity
func configured() error {
	if mgmtDb != nil && mgmtDb.Db != nil {
		return nil
	}
	return fmt.Errorf("Metadata: Database not configured.")
}
