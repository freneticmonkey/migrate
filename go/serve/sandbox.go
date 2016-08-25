package serve

import (
	"net/http"

	"github.com/gorilla/mux"
)

// registerSandboxEndpoints Register the table functions for the REST API
func registerSandboxEndpoints(r *mux.Router) {
	r.HandleFunc("/api/sandbox/migrate", migrate)
	r.HandleFunc("/api/sandbox/recreate", recreate)
	r.HandleFunc("/api/sandbox/pull-diff", pullDiff)
}

// migrate Get Table by Id
func migrate(w http.ResponseWriter, r *http.Request) {
	var err error

	// vars := mux.Vars(r)

	writeResponse(w, nil, err)
}

// recreate Recreate sandbox database from YAML
func recreate(w http.ResponseWriter, r *http.Request) {
	var err error

	// vars := mux.Vars(r)

	writeResponse(w, nil, err)
}

// pullDiff Serialise any manual MySQL changes to YAML
func pullDiff(w http.ResponseWriter, r *http.Request) {
	var err error

	// vars := mux.Vars(r)

	writeResponse(w, nil, err)
}
