package database

import (
	"strings"

	"github.com/freneticmonkey/migrate/go/util"
	"github.com/go-gorp/gorp"
)

var mgmtDb *gorp.DbMap

// Setup Setup the Database table in the management DB
func Setup(db *gorp.DbMap) {
	mgmtDb = db

	// Configure the Metadata table
	mgmtDb.AddTableWithName(TargetDatabase{}, "target_database").SetKeys(true, "DBID")

}

// CreateTables Create the Database table
func CreateTables() (result bool, err error) {

	createTable := []string{
		"CREATE TABLE `target_database` (",
		" `dbid` int(11) NOT NULL AUTO_INCREMENT,",
		" `project` varchar(255) DEFAULT NULL,",
		" `name` varchar(255) DEFAULT NULL,",
		" `env` varchar(255) DEFAULT NULL,",
		" PRIMARY KEY (`dbid`)",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}
	statement := strings.Join(createTable, "\n")

	// Execute the migration
	_, err = mgmtDb.Exec(statement)

	result = util.ErrorCheckf(err, "Problem creating Database table in the management DB")

	return result, err
}
