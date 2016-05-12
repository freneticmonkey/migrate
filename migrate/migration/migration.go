package migration

// Migration This struct stores the top migration properties.
type Migration struct {
	MID                int64  `db:"mid,autoincrement,primarykey"`
	DB                 int64  `db:"db"`
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
func (m *Migration) Insert() {
	mgmtDb.Insert(m)

	for _, step := range m.Steps {
		step.Insert()
	}
}

// Update Update the Migration in the Management DB
func (m *Migration) Update() {
	mgmtDb.Update(m)

	for _, step := range m.Steps {
		step.Update()
	}
}
