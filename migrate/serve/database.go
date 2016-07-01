package serve

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/freneticmonkey/migrate/migrate/database"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/gorilla/mux"
)

// RegisterDatabaseEndpoints Register the database functions for the REST API
func RegisterDatabaseEndpoints(r *mux.Router) {
	r.HandleFunc("/api/database/{id}", GetDatabase)
}

// GetDatabase Temporary REST test function
func GetDatabase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Invalid Database Id: %d", id), err)
		return
	}

	db, err := database.Load(int64(id))

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Unable to load Database Object from Database Id: %d", id), err)
		return
	}
	WriteResponse(w, db, err)
}
