package serve

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/gorilla/mux"
)

// migrationList Helper type used to return a List of Migrations
type migrationList struct {
	Migrations []migration.Migration `json:"migrations"`
	Start      int64                 `json:"start"`
	End        int64                 `json:"end"`
	Total      int64                 `json:"total"`
}

// registerMigrationEndpoints Register the migration functions for the REST API
func registerMigrationEndpoints(r *mux.Router) {
	r.HandleFunc("/api/migration/{id}", getMigration)
	r.HandleFunc("/api/migration/list/{start}/{count}", listMigrations)
}

// getMigration Get Migration by Id
func getMigration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Invalid Migration Id: %d", id), err, nil)
		return
	}

	m, err := migration.Load(int64(id))

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to load Migration Id: %d", id), err, nil)
		return
	}
	writeResponse(w, m, err)
}

// listMigrations List all migrations
func listMigrations(w http.ResponseWriter, r *http.Request) {
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
		writeErrorResponse(w, r, fmt.Sprintf("Invalid Migration Start Id: %d", start), err, nil)
		return
	}

	migrations, end, total, err = migration.LoadList(start, count)

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to retrieve Migrations from: %d for count: %d", start, count), err, nil)
		return
	}

	payload := migrationList{
		Migrations: migrations,
		Start:      start,
		End:        end,
		Total:      total,
	}

	writeResponse(w, payload, err)
}
