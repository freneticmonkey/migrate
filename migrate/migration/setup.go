package migration

import "github.com/go-gorp/gorp"

var mgmtDb *gorp.DbMap
var targetDBID int

// Setup Setup the migration tables in the management DB
func Setup(db *gorp.DbMap, targetDatabaseID int) {
	mgmtDb = db
	targetDBID = targetDatabaseID

	// Configure the Metadata table
	mgmtDb.AddTableWithName(Migration{}, "migration").SetKeys(true, "MID")
	mgmtDb.AddTableWithName(Step{}, "migration_steps").SetKeys(true, "SID")

}
