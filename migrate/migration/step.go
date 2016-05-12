package migration

// Step This struct stores the state for a step in a migration
type Step struct {
	SID              int64  `db:"sid,autoincrement,primarykey"`
	MID              int64  `db:"mid"`
	DB               int64  `db:"db"`
	Statement        string `db:"statement"`
	ReverseStatement string `db:"reverse_statement"`
	Output           string `db:"output,size:1024"`
	Status           int    `db:"status"`
}

// Insert Insert the Step into the Management DB
func (s *Step) Insert() {
	mgmtDb.Insert(s)
}

// Update Update the Step in the Management DB
func (s *Step) Update() {
	mgmtDb.Update(s)
}