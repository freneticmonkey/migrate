package migration

import (
	"github.com/freneticmonkey/migrate/migrate/metadata"
	"github.com/freneticmonkey/migrate/migrate/table"
	"github.com/freneticmonkey/migrate/migrate/util"
)

// Step This struct stores the state for a step in a migration
type Step struct {
	SID      int64  `db:"sid,autoincrement,primarykey"`
	MID      int64  `db:"mid"`
	Op       int    `db:"op"`
	MDID     int64  `db:"mdid"`
	Name     string `db:"name"`
	Forward  string `db:"forward"`
	Backward string `db:"backward"`
	Output   string `db:"output,size:1024"`
	Status   int    `db:"status"`
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

// UpdateMetadata Use the Step info to update the database
func (s *Step) UpdateMetadata() (err error) {
	var m *metadata.Metadata

	if s.Status == Forced || s.Status == Complete {

		m, err = metadata.Load(s.MDID)

		if util.ErrorCheckf(err, "Failed to load Metadata from the database") {
			return err
		}
		switch s.Op {

		case table.Add:
			// Mark exists
			m.Exists = true
			m.Update()
		case table.Mod:
			// If a rename has occurred, be sure to update the new name in the Metadata
			if m.Name != s.Name {
				m.Name = s.Name
				m.Update()
			}
		case table.Del:
			// If the operation is removing something, delete the associated Metadata
			m.Delete()
		}
	}

	return err
}
