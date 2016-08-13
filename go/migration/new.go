package migration

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/util"
)

// Param A struct used to define parameters for the setup and creation of the Migration struct
type Param struct {
	Project     string
	Version     string
	Timestamp   string
	Description string
	Forwards    mysql.SQLOperations
	Backwards   mysql.SQLOperations
	Rollback    bool
	Sandbox     bool
}

// New Migration constructor which also creates Steps and add everything
// to the database
func New(p Param) (m Migration, err error) {

	// Validate
	var alreadyExists bool
	var isLatest bool
	var valid bool
	var existing bool

	valid = true

	// If there are existing Migrations, validate this migration
	if existing, err = HasMigrations(); existing {

		// If the migration isn't flagged as a sandbox migration
		if !p.Sandbox {

			// Migration already created
			alreadyExists, err = VersionExists(p.Version)
			if alreadyExists && err == nil {
				err = fmt.Errorf("Migration with version: [%s] already exists.", p.Version)
				valid = false
			}

			// Migration too old
			if valid {
				if !p.Rollback {
					isLatest, err = IsLatest(p.Timestamp)
					if !isLatest && err == nil {
						err = fmt.Errorf("Migration with version: [%s] cannot be created as a newer version already exists.", p.Version)
						valid = false
					}
				} else {
					util.LogWarnf("Creation of Rollback Migration Detected! Be sure you want to apply these changes. Project: [%s] Version: [%s] Time (UTC): [%s]", p.Project, p.Version, p.Timestamp)
				}
			}

			// Migration doesn't do anything
			if valid {
				if len(p.Forwards) == 0 {
					valid = false
					err = fmt.Errorf("Empty Migration detected. Cannot continue. Project: [%s] Version: [%s] Time (UTC): [%s]", p.Project, p.Version, p.Timestamp)
				}
			}

		} else {
			util.LogWarnf("Sandbox Migration Detected. Skipping validation")
		}

	} else if err != nil {
		valid = false
	}

	if valid {

		// Ensure that the migration actually has steps (not an empty migration)
		if len(p.Forwards) > 0 {
			// Insert the Migration and its Steps into the Management DB
			m = Migration{
				DB:                 projectDBID,
				Project:            p.Project,
				Version:            p.Version,
				VersionTimestamp:   p.Timestamp,
				VersionDescription: p.Description,
				Status:             Unapproved,
				Sandbox:            p.Sandbox,
			}

			for i := 0; i < len(p.Forwards); i++ {
				forward := p.Forwards[i]

				// Insert the metadata
				err = forward.Metadata.OnCreate()
				if util.ErrorCheckf(err, "Failed to insert Metadata for Migration.") {
					return m, err
				}

				step := Step{
					Forward:  forward.Statement,
					Backward: p.Backwards[i].Statement,
					Status:   Unapproved,
					Op:       forward.Op,
					MDID:     forward.Metadata.MDID,
					Name:     forward.Name,
				}
				m.AddStep(step)
			}
			m.Insert()
		} else {
			return m, fmt.Errorf("Migration creation failed.  No operations detected for Project: [%s] Version: [%s]", p.Project, p.Version)
		}
	}

	return m, err
}
