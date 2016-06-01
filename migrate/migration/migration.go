package migration

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/util"
)

// Migration This struct stores the top migration properties.
type Migration struct {
	MID                int64  `db:"mid,autoincrement,primarykey"`
	DB                 int    `db:"db"`
	Project            string `db:"project"`
	Version            string `db:"version"`
	VersionTimestamp   string `db:"version_timestamp"`
	VersionDescription string `db:"version_description,size:512"`
	Status             int    `db:"status"`
	Timestamp          string `db:"timestamp"`

	Steps   []Step `db:"-"`
	Sandbox bool   `db:"-"`
}

// AddStep Add a Step to the migration
func (m *Migration) AddStep(step Step) {
	m.Steps = append(m.Steps, step)
}

// Insert Insert the Migration into the Management DB
func (m *Migration) Insert() (err error) {

	// If not in the sandbox
	if !m.Sandbox {
		err = mgmtDb.Insert(m)

		if !util.ErrorCheckf(err, "Inserting Migration into the DB failed for Project: [%s] with Version: [%s]", m.Project, m.Version) {
			for i := 0; i < len(m.Steps); i++ {
				m.Steps[i].MID = m.MID
				err = m.Steps[i].Insert()
				if util.ErrorCheckf(err, "Inserting Migration Step into the DB failed for Project: [%s] with Version: [%s]", m.Project, m.Version) {
					break
				}
			}
		}
	}

	return err
}

// Update Update the Migration in the Management DB
func (m *Migration) Update() (err error) {

	// If not in the Sandbox
	if !m.Sandbox {
		_, err = mgmtDb.Update(m)

		for i := 0; i < len(m.Steps); i++ {
			err = m.Steps[i].Update()
			if !util.ErrorCheckf(err, "Updating Migration Step into the DB failed for Project: [%s] with Version: [%s]", m.Project, m.Version) {
				break
			}
		}
	}

	return err
}

// Load Load a migation from the DB using the Migration ID primary key
func Load(mid int64) (m *Migration, err error) {
	obj, err := mgmtDb.Get(Migration{}, mid)

	if obj == nil {
		err = fmt.Errorf("Migration: [%d] not found in the DB", mid)
	}

	if err == nil {
		m = obj.(*Migration)

		var steps []Step
		query := fmt.Sprintf("select * from migration_steps WHERE mid = %d", mid)
		_, err = mgmtDb.Select(&steps, query)
		if err == nil {
			m.Steps = steps
		}
	}
	return m, err
}
