package test

import "testing"

type ProjectDB struct {
	MockDB
}

func CreateProjectDB(context string, t *testing.T) (p ProjectDB, err error) {
	db, mock, err := createMockDB()
	if err != nil {
		t.Errorf("%s: Setup Project DB Failed with Error: %v", context, err)
		return p, err
	}
	p = ProjectDB{MockDB{db, mock, "project"}}

	return p, nil
}

func (m *ProjectDB) ShowTables(results []DBRow, expectEmpty bool) {

	query := DBQueryMock{
		Query:   "show tables",
		Columns: []string{"table"},
	}

	if !expectEmpty {
		query.Rows = results
	}
	m.ExpectQuery(query)
}

func (m *ProjectDB) ShowCreateTable(name string, createStatement string) {
	query := DBQueryMock{
		Columns: []string{
			"name",
			"create_table",
		},
		Rows: []DBRow{
			{
				name,
				createStatement,
			},
		},
	}
	query.FormatQuery("show create table %s", name)

	m.ExpectQuery(query)
}
