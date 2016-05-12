package migration

import "github.com/go-gorp/gorp"

var mgmtDb *gorp.DbMap

// Setup Setup the migration tables in the management DB
func Setup(db *gorp.DbMap) {
	mgmtDb = db

	// Configure the Metadata table
	mgmtDb.AddTableWithName(Migration{}, "migration").SetKeys(true, "MID")
	mgmtDb.AddTableWithName(Step{}, "migration_steps").SetKeys(true, "SID")

}
