package serve

import (
	"fmt"
	"net/http"

	"github.com/freneticmonkey/migrate/go/configsetup"
	"github.com/gorilla/mux"
)

// registerHealthEndpoints Register the health functions for the REST API
func registerHealthEndpoints(r *mux.Router) {
	r.HandleFunc("/api/health/", getHealth)
}

// getHealth Get server health
func getHealth(w http.ResponseWriter, r *http.Request) {
	verboseLogging(r)
	health := configsetup.CheckConfig(false)
	if !health.Ok() {
		writeErrorResponse(w, r, "Health Check State: BAD", fmt.Errorf("Health Check State: BAD"), health)
		return
	}
	writeResponse(w, health, nil)
}
