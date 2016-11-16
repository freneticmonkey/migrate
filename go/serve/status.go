package serve

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/gorilla/mux"
)

type migrationStatus struct {
	MID    int64 `json:"mid"`
	Status int   `json:"status"`
}

type stepStatus struct {
	SID    int64 `json:"sid"`
	Status int   `json:"status"`
}

type editStatus struct {
	Migrations []migrationStatus `json:"migrations"`
	Steps      []stepStatus      `json:"steps"`
	VettedBy   string            `json:"vetted_by"`
}

// registerStatusEndpoints Register the migration functions for the REST API
func registerStatusEndpoints(r *mux.Router) {
	r.HandleFunc("/api/status/edit/", setStatus)
}

// setStatus Update Migration and associated step status'
func setStatus(w http.ResponseWriter, r *http.Request) {
	status := editStatus{}
	migrationIds := []int64{}
	migrations := []migration.Migration{}
	var err error

	verboseLogging(r)

	stepIds := []int64{}
	steps := []migration.Step{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&status)

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to Read Migration"), err, nil)
		return
	}

	// Verify that a valid string has been supplied for the vetter
	if status.VettedBy == "" {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to update status.  No vetter supplied"), nil, nil)
		return
	}

	// Build a list of Migration Ids
	for _, m := range status.Migrations {
		migrationIds = append(migrationIds, m.MID)
	}

	migrations, err = migration.LoadMigrationsList(migrationIds)

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to load Migrations from POST Ids"), err, nil)
		return
	}

	// Build a list of Step Ids
	for _, s := range status.Steps {
		stepIds = append(stepIds, s.SID)
	}
	steps, err = migration.LoadStepsList(stepIds)

	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("Unable to load Steps from POST Ids"), err, nil)
		return
	}

	// Update the migration steps with the sent status for each of the steps and store to the database
	for _, step := range status.Steps {
		for _, dbstep := range steps {
			if dbstep.SID == step.SID {
				dbstep.Status = step.Status
				dbstep.VettedBy = status.VettedBy
				err = dbstep.Update()

				if util.ErrorCheck(err) {
					writeErrorResponse(w, r, fmt.Sprintf("Unable to update Step with ID: %d", dbstep.SID), err, nil)
					return
				}
			}
		}
	}

	// Update the migration steps with the sent status for each of the steps and store to the database
	for _, migrationStatus := range status.Migrations {
		for _, migration := range migrations {
			if migration.MID == migrationStatus.MID {
				migration.Status = migrationStatus.Status
				migration.VettedBy = status.VettedBy

				err = migration.Update()

				if util.ErrorCheck(err) {
					writeErrorResponse(w, r, fmt.Sprintf("Unable to update Migration with ID: %d", migration.MID), err, nil)
					return
				}
			}
		}
	}

	writeResponse(w, "ok", err)
}
