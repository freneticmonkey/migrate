package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/freneticmonkey/migrate/migrate/util/shell"
)

// ExecOptions A helper struct for parameters when executing a Migration
type ExecOptions struct {
	MID              int64
	Dryrun           bool
	Force            bool
	Rollback         bool
	PTODisabled      bool
	AllowDestructive bool
}

// Exec Apply the migration to the project database.  The parmeters can be used to just execute a dryrun, force past
// any validity checks, or disable using pt-online-schema-change.
func Exec(options ExecOptions) (err error) {

	mid := options.MID
	dryrun := options.Dryrun
	force := options.Force
	rollback := options.Rollback
	ptodisbled := options.PTODisabled
	allowDestructive := options.AllowDestructive

	var statement string
	var output string
	var success bool
	var m *Migration

	m, err = Load(mid)

	if !util.ErrorCheckf(err, "Couldn't load Migration: [%d] from the Management DB", mid) {

		// has the migration been approved for migration or if it is being forced
		// Assumes that this migration hasn't already been applied since the Load statement above
		if m.Status == Approved || force {

			// Validate the migration
			var isLatest bool
			var migrationRunning bool
			var lm Migration
			var inProgressID int64
			var failReason string

			// By default assume that this isn't the latest migration
			isLatest = false
			// Clearly an invalid Migration ID
			inProgressID = -1
			// By default we assume that another migration is running until proven otherwise
			migrationRunning = true

			// If we aren't knowingly applying an older state (rollback)
			if !rollback {
				// Ensure that this migration is the latest migration known to the DB
				lm, err = GetLatest()
				if err != nil {
					failReason = fmt.Sprintf("Couldn't get latest Migration from DB: ERROR: %v", err)
				} else {
					if lm.MID == mid {
						isLatest = true
					} else {
						failReason = fmt.Sprintf("Migration: [%d] has been automatically depreciated by a Migration request with a newer schema from Git", mid)

						// Mark the migration as depreciated so that it won't be run again.
						m.Status = Depreciated
						m.Update()
					}
				}
			}

			// Ensure that another migation isn't already in progress
			inProgressID, err = InProgressID()
			if err != nil {
				failReason = fmt.Sprintf("Couldn't determine if any Migrations were InProgress from DB: ERROR: %v", err)
			} else {
				if inProgressID == 0 {
					migrationRunning = false
				} else {
					failReason = fmt.Sprintf("Migration: [%d] cannot be run because another Migration: [%d] is already running", mid, inProgressID)
				}
			}

			if isLatest && !migrationRunning {
				// for each step in the migration
				for _, step := range m.Steps {

					var md *metadata.Metadata
					isDestructive := (step.Op != table.Add)

					// check if ptodisabled is true
					usePTO := !ptodisbled

					// Check if create or drop table.
					md, err = metadata.Load(step.MDID)

					if !util.ErrorCheckf(err, "The Metadata: [%d] for Step: [%d] couldn't be loaded from the Management DB", step.MDID, step.SID) {

						// if PTO can be used, and this migration is changing a table and
						// the modification is either a CREATE OR DROP TABLE.
						if usePTO && md.IsTable() && (step.Op == table.Add || step.Op == table.Del) {
							// if so, use the regular go sql driver to execute the migration
							usePTO = false
						}

						// If the Step has been approved to be applied
						if step.Status == Approved {

							success = false
							statement = step.Forward

							if dryrun {

								if !allowDestructive && isDestructive {
									util.LogAttentionf("(DRYRUN) Skipping Migration Step: [%d]: Unapproved destructive change", step.SID)
								} else {
									// execute a dryrun of the migration step
									if usePTO {
										output, err = executePTO(statement, dryrun)
									} else {
										// otherwise use the regular go sql driver
										output, err = executeSQL(statement, dryrun)
									}
									util.LogAttentionf("(DRYRUN) Migration Step: [%d]\n%s", step.SID, output)
								}

								// Dryrun successful
								success = true

							} else {
								// If the change is destructive and it hasn't been approved, skip it
								if !allowDestructive && isDestructive {
									step.Output = fmt.Sprintf("Skipping Destructive Migration Step: [%d]: Unapproved destructive change", step.SID)
									step.Status = Skipped

								} else {

									// Indicate that the step is going to be applied
									step.Status = InProgress
									step.Update()

									// execute the migration
									if usePTO {
										output, err = executePTO(statement, dryrun)
									} else {
										// otherwise use the regular go sql driver
										output, err = executeSQL(statement, dryrun)
									}

									if !util.ErrorCheckf(err, "Migration Step: [%d] Apply Failed with ERROR: ", output) {
										// Record the result into the step table
										step.Output = output

										if force {
											step.Status = Forced
										} else {
											step.Status = Complete
										}

										// Message that the migration step was successful
										success = true

									} else {

										// Record the failure into the DB
										step.Output = fmt.Sprintf("Failed with Error: %v", err)
										step.Status = Failed
									}
								}

								// Record the result of the migration
								step.Update()
							}
						} else {
							util.LogWarnf("Migration Step: [%d] isn't approved to be applied. Skipping.", step.SID)

							// A skipped step is still successful
							success = true
						}
					}

					// If unsuccessful, halt the migration
					if !success {
						util.LogWarn("Migration Step Failed.  Halting migration.")
						break
					}
				}
			} else {
				err = fmt.Errorf("Migration with ID: [%d] failed validation. Reason: %s", m.MID, failReason)
			}
		} else {
			err = fmt.Errorf("Migration with id: [%d] has not been approved for migration.  Migration failed.", m.MID)
		}
	}

	return err
}

func executePTO(statement string, dryrun bool) (output string, err error) {

	params := []string{
		fmt.Sprintf("D=%s", "test"),
		fmt.Sprintf("t=%s", "test"),
		fmt.Sprintf("--alter \"%s\"", statement),
		"--critical-load \"Threads_running=500\"",
		"--execute",
	}

	if dryrun {
		output = fmt.Sprintf("PTO: [pt-online-schema-change %s]", strings.Join(params, " "))
	} else {
		output, err = shell.Run("pt-online-schema-change", "pto: ", params)
	}

	return output, err
}

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
