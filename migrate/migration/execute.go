package migration

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// Exec Apply the migration to the target database.  The parmeters can be used to just execute a dryrun, force past
// any validity checks, or disable using pt-online-schema-change.
func Exec(migration *Migration, dryrun bool, force bool, ptodisbled bool) (err error) {

	// has the migration been approved for migration or if it is being forced
	if migration.Status == Approved || force {
		// for each step in the migration
		for _, step := range migration.Steps {
			var md *metadata.Metadata

			util.DebugDump(step)

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

				// execute the migration
				if usePTO {
					err = executePTO(step, md, dryrun)
				} else {
					// otherwise use the regular go sql driver
					err = executeSQL(step, md, dryrun)
				}
			}
		}
	} else {
		err = fmt.Errorf("Migration with id: [%d] has not been approved for migration.  Migration failed.", migration.MID)
	}

	return err
}

func executePTO(step Step, md *metadata.Metadata, dryrun bool) (err error) {
	util.LogInfof("Executing the with Metadata: [%d] migration step: [%d] using pt-online-schema-change", md.MDID, step.SID)
	return err
}

func executeSQL(step Step, md *metadata.Metadata, dryrun bool) (err error) {
	util.DebugDump(md)
	util.LogInfof("Executing the with Metadata: [%d] migration step: [%d] using go sql driver", md.MDID, step.SID)
	return err
}
