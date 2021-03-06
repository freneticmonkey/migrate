package exec

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
)

// Options A helper struct for parameters when executing a Migration
type Options struct {
	MID               int64
	Dryrun            bool
	ForceCI           bool
	Rollback          bool
	PTODisabled       bool
	AllowDestructive  bool
	Migration         *migration.Migration
	Sandbox           bool
	StepConfirm    	  bool
}

// Exec Apply the migration to the project database.  The parmeters can be used to just execute a dryrun, force past
// any validity checks, or disable using pt-online-schema-change.
func Exec(options Options) (err error) {

	mid := options.MID
	dryrun := options.Dryrun
	force := options.ForceCI
	rollback := options.Rollback
	sandbox := options.Sandbox
	ptodisbled := options.PTODisabled
	allowDestructive := options.AllowDestructive
	stepConfirm := options.StepConfirm

	m := options.Migration

	var statement string
	var output string
	var success bool
	var action string

	// If we are in the sandbox, rollback migrations are allowed.
	if sandbox {
		rollback = true
	}


	// If a Migration ID was supplied in the Migration Options, then attempt to load from the DB
	if mid > 0 {
		m, err = migration.Load(mid)
		if util.ErrorCheckf(err, "Couldn't load Migration: [%d] from the Management DB", mid) {
			return err
		}
	} else if !(options.Sandbox && m != nil) {
		return fmt.Errorf("Migration failed.  Invalid Migration Id: [%d]", mid)
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

		// If we aren't knowingly applying an older state (rollback) and we aren't in the sandbox
		if !rollback && !sandbox {
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

		// Check the migration for destructive changes, and verify if they are allowed.
		unapprovedDestructive := false
		destructiveChanges := []string{}
		for _, step := range m.Steps {
			// If Destructive
			if step.Op != table.Add {
				destructiveChanges = append(destructiveChanges, step.Forward)

				// If not destruction not approved - fail
				if !options.AllowDestructive {
					unapprovedDestructive = true
					failReason = fmt.Sprintf("Migration: [%d] cannot be applied because it contains destructive change(s): [%s] without use of the --allow-destructive flag", mid, step.Forward)
					break
				}
			}
		}

		// if not forced, and  prompt for destructive approval
		if !force && len(destructiveChanges) > 0 && !unapprovedDestructive {
			util.LogWarn("The following DESTRUCTIVE changes have been detected.")
			util.LogAttentionf("\t%s", strings.Join(destructiveChanges, "\n\t"))
			action, err = util.SelectAction("Do you wish to continue? (y/n)", []string{"y", "n"})

			// Fail if not approved, or there was some kind of error reading input
			if action != "y" || err != nil {
				failReason = fmt.Sprintf("Migration: [%d] cannot be applied because it contains destructive change(s).", mid)
				unapprovedDestructive = true
			}
		}

		// We assume that everything is ok by default
		migrationCanExecute := true

		// If the migration isn't a rollback and we're not in the sandbox, ensure that it's the latest migration
		if !rollback && !sandbox && !isLatest {

			// if it's the sandbox we can ignore this fail state
			if !options.Sandbox {
				// If not, can't run
				migrationCanExecute = false
				failReason = fmt.Sprintf("Migration is too old to apply.  Use --rollback to force")
			}
		}

		// If there's a problem with destructive changes
		if unapprovedDestructive {
			migrationCanExecute = false
		}

		// If there's another migration already running
		if migrationRunning {
			migrationCanExecute = false
		}

		// If this migration can execute, then start applying it
		if migrationCanExecute {

			// Flag the migration as running
			if !dryrun && !m.Sandbox {
				m.Status = migration.InProgress
				err = m.Update()

				// If there was a problem updating
				if err != nil {
					return err
				}
			}

			// for each step in the migration
			for i := 0; i < len(m.Steps); i++ {
				step := m.Steps[i]

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

						// If we are performing a rollback and we're not in the sandbox, use the backward migration
						if rollback && !sandbox {
							statement = step.Backward
						}

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
								m.Steps[i].Output = fmt.Sprintf("Skipping Destructive Migration Step: [%d]: Unapproved destructive change", step.SID)
								m.Steps[i].Status = migration.Skipped

							} else {
								skipped := false

								// If the alter statements are going to be manually confirmed
								if stepConfirm {
									action := "no"

									action, err = util.SelectAction(fmt.Sprintf("SQL: [%s]\nApply ALTER?", statement), []string{"yes", "no"})
									if util.ErrorCheck(err) {
										util.LogError("There was a problem confirming the action.")
									}

									if action == "no" {
										skipped = false
									}
								}

								if skipped {

									// Indicate that the step is going to be applied
									m.Steps[i].Status = migration.Skipped
									m.Steps[i].Output = fmt.Sprintf("Skipping Migration Step: [%d]: Manually skipped ", step.SID)
									err = m.Steps[i].Update()
									if util.ErrorCheck(err) {
										return err
									}

								} else {

									// Indicate that the step is going to be applied
									m.Steps[i].Status = migration.InProgress
									err = m.Steps[i].Update()
									if util.ErrorCheck(err) {
										return err
									}

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
										m.Steps[i].Output = output

										if force {
											m.Steps[i].Status = migration.ForcedCI
										} else if rollback {
											m.Steps[i].Status = migration.Rollback
										} else {
											m.Steps[i].Status = migration.Complete
										}

										// Message that the migration step was successful
										success = true

									} else {

										// Record the step failure into the DB
										failReason = fmt.Sprintf("Failed with Error: %v", err)
										m.Steps[i].Output = failReason
										m.Steps[i].Status = migration.Failed
										err = step.Update()

										if err != nil {
											return err
										}

										failReason = fmt.Sprintf("Step: [%d] ", step.SID) + failReason

										// Record the Migration as failed into the DB
										m.Status = migration.Failed
										err = m.Update()

										if err != nil {
											return err
										}

										// Format an error message
										err = fmt.Errorf("Migration with ID: [%d] failed during apply. Reason: %s", m.MID, failReason)

										success = false
										break
									}
								}
							}

							// Record the result of the migration
							err = m.Steps[i].Update()

							// Die immediately because there's some kind of DB connectivity issue
							if err != nil {
								return err
							}

							// If necessary, update the Metadata in the database
							err = m.Steps[i].UpdateMetadata()

							// Die immediately because there's some kind of DB connectivity issue
							if err != nil {
								return err
							}

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
				if !dryrun {
					if force {
						m.Status = migration.ForcedCI
					} else if rollback {
						m.Status = migration.Rollback
					} else {
						m.Status = migration.Complete
					}
					err = m.Update()
					if err != nil {
						return err
					}
					util.LogInfof("Migration with ID: [%d] and Description: [%s] completed successfully with status: [%s]", m.MID, m.VersionDescription, migration.StatusString[m.Status])
				} else {
					util.LogInfof("(DRYRUN) Migration with ID: [%d] and Description: [%s] completed successfully with status: [%s]", m.MID, m.VersionDescription, migration.StatusString[m.Status])
				}

			}

		} else {
			err = fmt.Errorf("Migration with ID: [%d] and Description: [%s] failed validation. Reason: %s", m.MID, m.VersionDescription, failReason)
		}
	} else {
		err = fmt.Errorf("Migration with id: [%d] and Description: [%s] has not been approved for migration.  Migration failed.", m.MID, m.VersionDescription)
	}

	return err
}
