package migration

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/mysql"
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

	Steps []Step `db:"-"`
}

// AddStep Add a Step to the migration
func (m *Migration) AddStep(step Step) {
	m.Steps = append(m.Steps, step)
}

// Insert Insert the Migration into the Management DB
func (m *Migration) Insert() (err error) {
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
	return err
}

// Update Update the Migration in the Management DB
func (m *Migration) Update() (err error) {
	_, err = mgmtDb.Update(m)

	for i := 0; i < len(m.Steps); i++ {
		err = m.Steps[i].Update()
		if !util.ErrorCheckf(err, "Updating Migration Step into the DB failed for Project: [%s] with Version: [%s]", m.Project, m.Version) {
			break
		}
	}
	return err
}

// Load Load a migation from the DB using the Migration ID primary key
func Load(mid int64) (m *Migration, err error) {
	obj, err := mgmtDb.Get(Migration{}, mid)
	if err == nil {
		m = obj.(*Migration)
	}
	return m, err
}

// GetLatest Return the git timestamp latest Migration from the DB
func GetLatest() (m Migration, err error) {
	var migrations []Migration
	_, err = mgmtDb.Select(&migrations, "select * from migration ORDER BY version_timestamp DESC LIMIT 1")
	if !util.ErrorCheckf(err, "Unable to get latest Migration from Management DB") {
		if len(migrations) > 0 {
			m = migrations[0]
		}
	}
	return m, err
}

// InProgressID Returns the ID of a migration in the DB whose current status
// is InProgress.  If no Migration is running 0 is returned.
func InProgressID() (inProgressID int64, err error) {
	var migrations []Migration
	query := fmt.Sprintf("select * from migration WHERE status = %d", InProgress)
	_, err = mgmtDb.Select(&migrations, query)
	if !util.ErrorCheckf(err, "Unable to get InProgress Migrations from Management DB") {
		if len(migrations) > 0 {
			inProgressID = migrations[0].MID
		}
	}
	return inProgressID, err
}

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
func New(p Param) Migration {
	m := Migration{
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

	// for _, s := range m.Steps {
	// 	util.DebugDump(s)
	// }

	return m
}
