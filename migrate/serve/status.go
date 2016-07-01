package serve

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/freneticmonkey/migrate/migrate/migration"
	"github.com/freneticmonkey/migrate/migrate/util"
	"github.com/gorilla/mux"
)

type MigrationStatus struct {
	MID    int64 `json:"mid"`
	Status int   `json:"status"`
}

type StepStatus struct {
	SID    int64 `json:"sid"`
	Status int   `json:"status"`
}

type EditStatus struct {
	Migrations []MigrationStatus `json:"migrations"`
	Steps      []StepStatus      `json:"steps"`
}

// RegisterStatusEndpoints Register the migration functions for the REST API
func RegisterStatusEndpoints(r *mux.Router) {
	r.HandleFunc("/api/status/edit/", SetStatus)
}

// SetStatus Update Migration and associated step status'
func SetStatus(w http.ResponseWriter, r *http.Request) {
	status := EditStatus{}
	migrationIds := []int64{}
	migrations := []migration.Migration{}
	var err error

	stepIds := []int64{}
	steps := []migration.Step{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&status)

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Unable to Read Migration"), err)
		return
	}

	// Build a list of Migration Ids
	for _, m := range status.Migrations {
		migrationIds = append(migrationIds, m.MID)
	}

	migrations, err = migration.LoadMigrationsList(migrationIds)

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Unable to load Migrations from POST Ids"), err)
		return
	}

	// Build a list of Step Ids
	for _, s := range status.Steps {
		stepIds = append(stepIds, s.SID)
	}
	steps, err = migration.LoadStepsList(stepIds)

	if util.ErrorCheck(err) {
		WriteErrorResponse(w, fmt.Sprintf("Unable to load Steps from POST Ids"), err)
		return
	}

	// Update the migration steps with the sent status for each of the steps and store to the database
	for _, step := range status.Steps {
		for _, dbstep := range steps {
			if dbstep.SID == step.SID {
				dbstep.Status = step.Status
				err = dbstep.Update()

				if util.ErrorCheck(err) {
					WriteErrorResponse(w, fmt.Sprintf("Unable to update Step with ID: %d", dbstep.SID), err)
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

				err = migration.Update()

				if util.ErrorCheck(err) {
					WriteErrorResponse(w, fmt.Sprintf("Unable to update Migration with ID: %d", migration.MID), err)
					return
				}
			}
		}
	}

	WriteResponse(w, "ok", err)
}
