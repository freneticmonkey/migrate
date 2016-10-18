package serve

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/freneticmonkey/migrate/go/database"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/gorilla/mux"
)

// registerDatabaseEndpoints Register the database functions for the REST API
func registerDatabaseEndpoints(r *mux.Router) {
	r.HandleFunc("/api/database/{id}", getDatabase)
}

// getDatabase Temporary REST test function
func getDatabase(w http.ResponseWriter, r *http.Request) {
	verboseLogging(r)
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Invalid Database Id: %d", id), err, nil)
		return
	}

	db, err := database.Load(int64(id))

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to load Database Object from Database Id: %d", id), err, nil)
		return
	}
	writeResponse(w, db, err)
}
