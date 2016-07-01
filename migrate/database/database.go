package database

import (
	"fmt"

	"github.com/freneticmonkey/migrate/migrate/util"
)

// Database Environments
const (
	SANDBOX = iota
	DEV
	STAGE
	MLT
	LT
	PROD
)

// EnvNames Database Environment Names for debugging purposes
var EnvNames = []string{
	"SANDBOX",
	"DEV",
	"STAGE",
	"MLT",
	"LT",
	"PROD",
}

// TargetDatabase TargetDatabase stores a list of target databases which allow for a single
// management database to store migrations, metedata, and steps for multiple
// target databases
type TargetDatabase struct {
	DBID    int    `db:"dbid, autoincrement, primarykey" json:"dbid"`
	Project string `db:"project" json:"project"`
	Name    string `db:"name" json:"name"`
	Env     string `db:"env" json:"env"`
}

// Insert Insert the Database into the Management DB
func (d *TargetDatabase) Insert() error {
	return mgmtDb.Insert(d)
}

// Load Load a migation from the DB using the Migration ID primary key
func Load(dbid int64) (db *TargetDatabase, err error) {
	obj, err := mgmtDb.Get(TargetDatabase{}, dbid)

	if obj == nil {
		err = fmt.Errorf("Database: [%d] not found in the DB", dbid)
	}
	db = obj.(*TargetDatabase)

	return db, err
}

// GetbyProject Get a target database from the management db using the names of
// the project, database, and environment
func GetbyProject(project string, name string, env string) (db TargetDatabase, err error) {
	err = mgmtDb.SelectOne(&db, fmt.Sprintf("SELECT * FROM target_database WHERE project=\"%s\" AND name=\"%s\" AND env=\"%s\"", project, name, env))
	util.ErrorCheckf(err, "Failed to find Target Database for Project: [%s] with Name: [%s] and Env: [%s]", project, name, env)

	return db, err
}
