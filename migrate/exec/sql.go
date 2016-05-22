package exec

import (
	"database/sql"
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/util"
)

func executeSQL(statement string, dryrun bool) (output string, err error) {
	var ready bool
	var result sql.Result
	var rowsAffected int64

	// Ensure that the project DB connection is open
	ready, err = connectProjectDB()

	if dryrun {
		output = fmt.Sprintf("SQL: [%s]", statement)

	} else {
		// If the connection is ok
		if ready && !util.ErrorCheckf(err, "Migration Failed to open Project DB") {

			// Execute the migration
			result, err = projectDB.Exec(statement)
			if !util.ErrorCheckf(err, "Migration Step: ALTER TABLE Failed: [%v]", err) {

				// Record the result into the step table
				rowsAffected, err = result.RowsAffected()
				output = fmt.Sprintf("Row(s) Affected: %d", rowsAffected)

			} else {
				output = fmt.Sprintf("Failed with Error: %v", err)
			}
		}
	}

	return output, err
}
