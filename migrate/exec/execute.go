package exec

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// Options A helper struct for parameters when executing a Migration
type Options struct {
	MID              int64
	Dryrun           bool
	Force            bool
	Rollback         bool
	PTODisabled      bool
	AllowDestructive bool
	Migration        *migration.Migration
	Sandbox          bool
}

// Exec Apply the migration to the project database.  The parmeters can be used to just execute a dryrun, force past
// any validity checks, or disable using pt-online-schema-change.
func Exec(options Options) (err error) {

	mid := options.MID
	dryrun := options.Dryrun
	force := options.Force
	rollback := options.Rollback
	ptodisbled := options.PTODisabled
	allowDestructive := options.AllowDestructive
	m := options.Migration

	var statement string
	var output string
	var success bool

	// If a Migration ID was supplied in the Migration Options, then attempt to load from the DB
	if mid > 0 {
		m, err = migration.Load(mid)
		if util.ErrorCheckf(err, "Couldn't load Migration: [%d] from the Management DB", mid) {
			return err
		}
	}

	// TODO: Update the Migration state at the end of the Migration!!!!

	// has the migration been approved for migration or if it is being forced
	// Assumes that this migration hasn't already been applied since the Load statement above
	if m.Status == migration.Approved || force {

		// Validate the migration
		var isLatest bool
		var migrationRunning bool
		var lm migration.Migration
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
			lm, err = migration.GetLatest()
			if err != nil {
				failReason = fmt.Sprintf("Couldn't get latest Migration from DB: ERROR: %v", err)
			} else {
				if lm.MID == mid {
					isLatest = true
				} else {
					failReason = fmt.Sprintf("Migration: [%d] has been automatically depreciated by a Migration request with a newer schema from Git", mid)

					// Mark the migration as depreciated so that it won't be run again.
					m.Status = migration.Depreciated
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

		// If this migration is the latest and another migration is not running, OR we're in the sandbox
		if isLatest && !migrationRunning || options.Sandbox {

			// Flag the migration as running
			if !dryrun {
				m.Status = migration.InProgress
				m.Update()
			}

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
					if step.Status == migration.Approved || options.Sandbox {

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
									output, err = ExecuteSQL(statement, dryrun)
									util.ErrorCheckf(err, "Migration Step: ALTER TABLE Failed: [%v]", err)

								}
								util.LogAttentionf("(DRYRUN) Migration Step: [%d]\n%s", step.SID, output)
							}

							// Dryrun successful
							success = true

						} else {
							// If the change is destructive and it hasn't been approved, skip it
							if !allowDestructive && isDestructive {
								step.Output = fmt.Sprintf("Skipping Destructive Migration Step: [%d]: Unapproved destructive change", step.SID)
								step.Status = migration.Skipped

							} else {

								// Indicate that the step is going to be applied
								step.Status = migration.InProgress
								step.Update()

								// execute the migration
								if usePTO {
									output, err = executePTO(statement, dryrun)
								} else {
									// otherwise use the regular go sql driver
									output, err = ExecuteSQL(statement, dryrun)
									util.ErrorCheckf(err, "Migration Step: ALTER TABLE Failed: [%v]", err)
								}

								if !util.ErrorCheckf(err, "Migration Step: [%d] Apply Failed with ERROR: ", output) {
									// Record the result into the step table
									step.Output = output

									if force {
										step.Status = migration.Forced
									} else {
										step.Status = migration.Complete
									}

									// Message that the migration step was successful
									success = true

								} else {

									// Record the step failure into the DB
									failReason = fmt.Sprintf("Failed with Error: %v", err)
									step.Output = failReason
									step.Status = migration.Failed
									step.Update()

									failReason = fmt.Sprintf("Step: [%d] ", step.SID) + failReason

									// Record the Migration as failed into the DB
									m.Status = migration.Failed
									m.Update()

									// Format an error message
									err = fmt.Errorf("Migration with ID: [%d] failed during apply. Reason: %s", m.MID, failReason)

									success = false
									break
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

				// Finished Migration Step
			}

			// Store success in the database
			if success {
				if force {
					m.Status = migration.Forced
				} else {
					m.Status = migration.Complete
				}
				m.Update()

				util.LogInfof("Migration with ID: [%d] completed successfully.", m.MID)
			}

		} else {
			err = fmt.Errorf("Migration with ID: [%d] failed validation. Reason: %s", m.MID, failReason)
		}
	} else {
		err = fmt.Errorf("Migration with id: [%d] has not been approved for migration.  Migration failed.", m.MID)
	}

	return err
}