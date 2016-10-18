package serve

import (
	"fmt"
	"net/http"

	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/sandbox"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/gorilla/mux"
)

// registerSandboxEndpoints Register the table functions for the REST API
func registerSandboxEndpoints(r *mux.Router) {
	r.HandleFunc("/api/sandbox/diff/", diffTables).Methods("GET")
	r.HandleFunc("/api/sandbox/diff/{id}", diffTables).Methods("GET")
	r.HandleFunc("/api/sandbox/migrate", migrate)
	r.HandleFunc("/api/sandbox/recreate", recreate)
	r.HandleFunc("/api/sandbox/pull-diff", pullDiff)
}

// migrate Get Table by Id
func migrate(w http.ResponseWriter, r *http.Request) {
	var err error
	var result string

	verboseLogging(r)
	yaml.Schema = []table.Table{}

	result, err = sandbox.Action(conf, false, false, "REST API Migrate")
	if err != nil {
		writeErrorResponse(w, r, "Migrate FAILED.", err, nil)
		return
	}

	writeResponse(w, result, err)
}

// recreate Recreate sandbox database from YAML
func recreate(w http.ResponseWriter, r *http.Request) {
	var err error
	var result string

	verboseLogging(r)
	yaml.Schema = []table.Table{}

	result, err = sandbox.Action(conf, false, true, "REST API Recreate Database")
	if err != nil {
		writeErrorResponse(w, r, "Migrate FAILED.", err, nil)
		return
	}

	writeResponse(w, result, err)
}

// pullDiff Serialise any manual MySQL changes to YAML
func pullDiff(w http.ResponseWriter, r *http.Request) {
	var err error
	var result string

	verboseLogging(r)
	vars := mux.Vars(r)

	tableName := vars["id"]

	result, err = sandbox.PullDiff(conf, tableName)

	if err != nil {
		writeErrorResponse(w, r, "Pull-diff FAILED.", err, nil)
		return
	}

	writeResponse(w, result, err)
}

func diffTables(w http.ResponseWriter, r *http.Request) {
	var forwardDiff table.Differences
	var forwardOps mysql.SQLOperations
	var problems id.ValidationErrors
	var err error

	verboseLogging(r)

	// Extract the table name
	vars := mux.Vars(r)
	tableName := vars["id"]

	// Empty the yaml.Schema prediff
	yaml.Schema = []table.Table{}

	// Read tables relative to the current working directory (which is the project name)
	err = yaml.ReadTables(conf)

	// Configure table filter
	diffSchema := yaml.Schema[:]

	if tableName != "" {

		targetTableFound := false

		// Filter by tableName in the YAML Schema
		tgtTbl := []table.Table{}

		for _, tbl := range diffSchema {
			if tbl.Name == tableName {
				tgtTbl = append(tgtTbl, tbl)
				targetTableFound = true
				break
			}
		}
		// Reduce the YAML schema to the single target table
		diffSchema = tgtTbl

		if !targetTableFound {
			writeErrorResponse(w, r, fmt.Sprintf("Diff failed.  Table `%s` doesn't exist.", tableName), err, nil)
			return
		}
	}

	// Read the MySQL tables from the target database

	mysql.Schema = []table.Table{}

	err = mysql.ReadTables(conf)
	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, "Diff failed.  Unable to read MySQL Schema`", err, nil)
		return
	}

	problems, err = id.ValidateSchema(mysql.Schema, "Target Database Schema", true)
	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, "Diff failed. MySQL schema failed validation.", err, problems)
		return
	}

	// Filter by tableName in the MySQL Schema
	dbSchema := mysql.Schema[:]
	if tableName != "" {
		tgtTbl := []table.Table{}

		for _, tbl := range mysql.Schema {
			if tbl.Name == tableName {
				tgtTbl = append(tgtTbl, tbl)
				break
			}
		}
		// Reduce the YAML schema to the single target table
		dbSchema = tgtTbl
	}

	problems, err = id.ValidatePropertyIDs(yaml.Schema, mysql.Schema, true)
	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, "Diff failed due to invalid schema. Detected YAML PropertyID problems", err, problems)
		return
	}

	forwardDiff, err = table.DiffTables(diffSchema, dbSchema, true)
	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, "Diff failed. Problems while calculating differences.", err, problems)
		return
	}

	forwardOps = mysql.GenerateAlters(forwardDiff)

	writeResponse(w, forwardOps, err)
}
