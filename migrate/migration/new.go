package migration

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/mysql"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// Param A struct used to define parameters for the Migration struct
type Param struct {
	Project     string
	Version     string
	Timestamp   string
	Description string
	Forwards    mysql.SQLOperations
	Backwards   mysql.SQLOperations
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
		// Migration already created
		alreadyExists, err = VersionExists(p.Version)
		if alreadyExists && err == nil {
			err = fmt.Errorf("Migration with version: [%s] already exists.", p.Version)
			valid = false
		}

		if valid {
			isLatest, err = IsLatest(p.Timestamp)
			if !isLatest && err == nil {
				err = fmt.Errorf("Migration with version: [%s] cannot be created as a newer version already exists.", p.Version)
				valid = false
			}
		}
	} else if err != nil {
		valid = false
	}

	if valid {

		// Insert the Migration and its Steps into the Management DB
		m = Migration{
			DB:                 projectDBID,
			Project:            p.Project,
			Version:            p.Version,
			VersionTimestamp:   p.Timestamp,
			VersionDescription: p.Description,
			Status:             Unapproved,
		}

		for i := 0; i < len(p.Forwards); i++ {
			forward := p.Forwards[i]

			step := Step{
				Forward:  forward.Statement,
				Backward: p.Backwards[i].Statement,
				Status:   Unapproved,
				Op:       forward.Op,
				MDID:     forward.Metadata.MDID,
			}
			m.AddStep(step)
		}
		util.LogWarn("Before insert")
		m.Insert()
	}

	return m, err
}
