package exec

import (
	"database/sql"
	"fmt"

	"github.com/freneticmonkey/migrate/go/util"
)

// ExecuteSQL Execute SQL in the Project DB
func ExecuteSQL(statement string, dryrun bool) (output string, err error) {
	var ready bool
	var result sql.Result
	var rowsAffected int64

	// Ensure that the project DB connection is open
	ready, err = ConnectProjectDB(false)

	if dryrun {
		output = fmt.Sprintf("SQL: [%s]", statement)

	} else {
		// If the connection is ok
		if ready && !util.ErrorCheckf(err, "Failed to open connection to Project DB") {

			// Execute the migration
			util.LogAlertf("SQL: Executing Migration: [%s]", statement)
			result, err = projectDB.Exec(statement)
			if !util.ErrorCheck(err) {

				// Record the result into the step table
				rowsAffected, err = result.RowsAffected()
				output = fmt.Sprintf("Row(s) Affected: %d", rowsAffected)

			} else {
				output = fmt.Sprintf("SQL Exec Failed with Error: %v", err)
			}
		}
	}

	return output, err
}
