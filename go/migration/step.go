package migration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
)

// Step This struct stores the state for a step in a migration
type Step struct {
	SID      int64  `db:"sid,autoincrement,primarykey" json:"sid"`
	MID      int64  `db:"mid" json:"mid"`
	Op       int    `db:"op" json:"op"`
	MDID     int64  `db:"mdid" json:"mdid"`
	Name     string `db:"name" json:"name"`
	Forward  string `db:"forward" json:"forward"`
	Backward string `db:"backward" json:"backward"`
	Output   string `db:"output,size:1024" json:"output"`
	Status   int    `db:"status" json:"status"`
	VettedBy string `db:"vetted_by" json:"vetted_by"`
}

// Insert Insert the Step into the Management DB
func (s *Step) Insert() error {
	return mgmtDb.Insert(s)
}

// Update Update the Step in the Management DB
func (s *Step) Update() (err error) {
	_, err = mgmtDb.Update(s)
	return err
}

// LoadStepsList Populate a slice of Steps using the Step Ids contained within sids
func LoadStepsList(sids []int64) (s []Step, err error) {
	var strIds []string
	for _, sid := range sids {
		strIds = append(strIds, strconv.FormatInt(sid, 10))
	}
	jstrIds := strings.Join(strIds, ",")

	query := fmt.Sprintf("select * from migration_steps WHERE sid IN (%s)", jstrIds)
	_, err = mgmtDb.Select(&s, query)

	if util.ErrorCheckf(err, "There was a problem retrieving Steps with Ids: [%s]", jstrIds) {
		return s, err
	}

	return s, err
}

// UpdateMetadata Use the Step info to update the database
func (s *Step) UpdateMetadata() (err error) {
	var m *metadata.Metadata

	if s.Status == Forced || s.Status == Complete || s.Status == Rollback {

		m, err = metadata.Load(s.MDID)
		if util.ErrorCheckf(err, "Failed to load Metadata from the database") {
			return err
		}

		switch s.Op {

		case table.Add:
			// Mark exists
			m.Exists = true

			// if rollback apply the reverse
			if s.Status == Rollback {
				m.Exists = false
			}

			err = m.Update()

			if m.IsTable() {
				// Update all table fields to exist
				metadata.SetTableExists(m.Name)
			}

		case table.Mod:
			// If a rename has occurred, be sure to update the new name in the Metadata
			if m.Name != s.Name {
				m.Name = s.Name
				err = m.Update()
			}
		case table.Del:

			// Mark not existant
			m.Exists = false

			// if rollback apply the reverse
			if s.Status == Rollback {
				m.Exists = true
			}

			err = m.Update()

			if m.IsTable() {
				// Delete all table fields
				metadata.MarkNonExistAllTableMetadata(m.Name)
			}
		}
	}

	return err
}

// ToDBRow Used to convert the Migration into a unit test DBRow
func (s Step) ToDBRow() test.DBRow {
	return test.DBRow{
		s.SID,
		s.MID,
		s.Op,
		s.MDID,
		s.Name,
		s.Forward,
		s.Backward,
		s.Output,
		s.Status,
		s.VettedBy,
	}
}
