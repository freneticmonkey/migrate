package serve

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/gorilla/mux"
)

// MigrationList Helper type used to return a List of Migrations
type MigrationList struct {
	Migrations []migration.Migration `json:"migrations"`
	Start      int64                 `json:"start"`
	End        int64                 `json:"end"`
	Total      int64                 `json:"total"`
}

// RegisterMigrationEndpoints Register the migration functions for the REST API
func RegisterMigrationEndpoints(r *mux.Router) {
	r.HandleFunc("/api/migration/{id}", GetMigration)
	r.HandleFunc("/api/migration/list/{start}/{count}", ListMigrations)
}

// GetMigration Get Migration by Id
func GetMigration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Invalid Migration Id: %d", id), err)
		return
	}

	m, err := migration.Load(int64(id))

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Unable to load Migration Id: %d", id), err)
		return
	}
	WriteResponse(w, m, err)
}

// ListMigrations List all migrations migrations
func ListMigrations(w http.ResponseWriter, r *http.Request) {
	var start int64
	var count int64
	var end int64
	var total int64
	var err error
	var migrations []migration.Migration

	// Setting default parameters
	start = 0
	count = 10

	vars := mux.Vars(r)

	for key, value := range vars {
		switch key {

		case "start":
			start, err = strconv.ParseInt(value, 10, 64)
		case "count":
			count, err = strconv.ParseInt(value, 10, 64)
		}
	}

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Invalid Migration Start Id: %d", start), err)
		return
	}

	migrations, end, total, err = migration.LoadList(start, count)

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Unable to retrieve Migrations from: %d for count: %d", start, count), err)
		return
	}

	payload := MigrationList{
		Migrations: migrations,
		Start:      start,
		End:        end,
		Total:      total,
	}

	WriteResponse(w, payload, err)
}
