package migration

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/freneticmonkey/migrate/go/util"
)

// Migration This struct stores the top migration properties.
type Migration struct {
	MID                int64  `db:"mid,autoincrement,primarykey" json:"mid"`
	DB                 int    `db:"db" json:"db"`
	Project            string `db:"project" json:"project"`
	Version            string `db:"version" json:"version"`
	VersionTimestamp   string `db:"version_timestamp" json:"version_timestamp"`
	VersionDescription string `db:"version_description,size:512" json:"version_description"`
	Status             int    `db:"status" json:"status"`
	Timestamp          string `db:"timestamp" json:"timestamp"`

	Steps   []Step `db:"-" json:"steps"`
	Sandbox bool   `db:"-" json:"-"`
}

// AddStep Add a Step to the migration
func (m *Migration) AddStep(step Step) {
	m.Steps = append(m.Steps, step)
}

// Insert Insert the Migration into the Management DB
func (m *Migration) Insert() (err error) {

	// If not in the sandbox
	// if !m.Sandbox {
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
	// }

	return err
}

// Update Update the Migration in the Management DB
func (m *Migration) Update() (err error) {

	_, err = mgmtDb.Update(m)

	if err == nil {
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

	var mig Migration
	if err = configured(); err != nil {
		return m, err
	}
	query := fmt.Sprintf("SELECT * FROM `migration` WHERE mid=%d", mid)
	err = mgmtDb.SelectOne(&mig, query)

	if err == nil {
		m = &mig
	} else if err == sql.ErrNoRows {
		err = fmt.Errorf("Migration: [%d] not found in the DB", mid)
	}

	// obj, err := mgmtDb.Get(Migration{}, mid)

	// if obj == nil && err == sql.ErrNoRows {
	// 	err = fmt.Errorf("Migration: [%d] not found in the DB", mid)
	// } else {
	// 	fmt.Printf("Sad :( %v)", err)
	// }

	if err == nil {
		// m = obj.(*Migration)

		var steps []Step
		query := fmt.Sprintf("SELECT * FROM `migration_steps` WHERE mid=%d", mid)
		_, err = mgmtDb.Select(&steps, query)
		if err == nil {
			m.Steps = steps
		}
	}
	return m, err
}

// LoadList Build a slice of Migrations.  count has a maximum size of 50
func LoadList(start int64, count int64) (migrations []Migration, end int64, total int64, err error) {

	// Restrict count to 50
	if count > 50 {
		count = 50
	}

	// Retrieve the number of migrations listed
	total, err = mgmtDb.SelectInt("select count(*) from migration")

	if !util.ErrorCheck(err) {

		// Retrive the slice of migrations
		query := fmt.Sprintf("select * from migration WHERE mid >= %d LIMIT %d", start, count)
		_, err = mgmtDb.Select(&migrations, query)

		if !util.ErrorCheck(err) {
			// If there wasn't any issues retrieving the Migrations, calculate the Migration Id of the end of the slice
			end = migrations[len(migrations)-1].MID
		}
	}

	return migrations, end, total, err
}

// LoadMigrationsList Populate a slice of Migrations using the Migration Ids contained within mids
func LoadMigrationsList(mids []int64) (m []Migration, err error) {
	var strIds []string
	for _, mid := range mids {
		strIds = append(strIds, strconv.FormatInt(mid, 10))
	}
	jstrIds := strings.Join(strIds, ",")

	if len(strIds) > 0 {
		query := fmt.Sprintf("select * from migration WHERE mid IN (%s)", jstrIds)
		_, err = mgmtDb.Select(&m, query)

		if util.ErrorCheckf(err, "There was a problem retrieving Migrations with Ids: [%s]", jstrIds) {
			return m, err
		}
	} else {
		err = fmt.Errorf("No Migration IDs detected. IDs: [%s]", jstrIds)
	}

	return m, err
}
