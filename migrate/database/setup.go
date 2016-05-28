package database

import "github.com/go-gorp/gorp"

var mgmtDb *gorp.DbMap

// Setup Setup the Database table in the management DB
func Setup(db *gorp.DbMap) {
	mgmtDb = db

	// Configure the Metadata table
	mgmtDb.AddTableWithName(TargetDatabase{}, "target_database").SetKeys(true, "DBID")

}
