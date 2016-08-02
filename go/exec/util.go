package exec

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/util"
)

// InProgressID Returns the ID of a migration in the DB whose current status
// is InProgress.  If no Migration is running 0 is returned.
func InProgressID() (inProgressID int64, err error) {
	var migrations []migration.Migration
	query := fmt.Sprintf("select * from migration WHERE status = %d", migration.InProgress)
	_, err = mgmtDb.Select(&migrations, query)
	if !util.ErrorCheckf(err, "Unable to get InProgress Migrations from Management DB") {
		if len(migrations) > 0 {
			inProgressID = migrations[0].MID
		}
	}
	return inProgressID, err
}
