package migration

// Step This struct stores the state for a step in a migration
type Step struct {
	SID      int64  `db:"sid,autoincrement,primarykey"`
	MID      int64  `db:"mid"`
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
