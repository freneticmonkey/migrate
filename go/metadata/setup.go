package metadata

import (
	"fmt"

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

// CreateTables If tables need to be created, management.Setup will call here
// first
func CreateTables() {

}

// configured Internal Helper function for checking database validity
func configured() error {
	if mgmtDb != nil && mgmtDb.Db != nil {
		return nil
	}
	return fmt.Errorf("Metadata: Database not configured.")
}
