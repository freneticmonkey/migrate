package metadata

import "github.com/go-gorp/gorp"

var mgmtDb *gorp.DbMap

// Setup Setup the Metadata table in the management DB
func Setup(db *gorp.DbMap) {
	mgmtDb = db

	// Configure the Metadata table
	mgmtDb.AddTableWithName(Metadata{}, "metadata").SetKeys(true, "MDID")

}
