package database

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/util"
)

// TargetDatabase TargetDatabase stores a list of target databases which allow for a single
// management database to store migrations, metedata, and steps for multiple
// target databases
type TargetDatabase struct {
	DBID    int    `db:"dbid, autoincrement, primarykey"`
	Project string `db:"project"`
	Name    string `db:"name"`
}

// Insert Insert the Database into the Management DB
func (d *TargetDatabase) Insert() error {
	return mgmtDb.Insert(d)
}

// GetbyProject Get a target database from the management db using the names of
// the project and the database
func GetbyProject(project string, name string) (db TargetDatabase, err error) {
	err = mgmtDb.SelectOne(&db, fmt.Sprintf("SELECT * FROM target_database WHERE project=\"%s\" AND name=\"%s\"", project, name))
	util.ErrorCheckf(err, "Failed to find Target Database for Project: [%s] with Name: [%s]", project, name)

	return db, err
}
