package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/gorilla/mux"
)

// registerTableEndpoints Register the table functions for the REST API
func registerTableEndpoints(r *mux.Router) {
	r.HandleFunc("/api/table/create", createTable).Methods("POST")
	r.HandleFunc("/api/table/{id}", getTable)
	r.HandleFunc("/api/table/{id}/edit", editTable).Methods("POST")
	r.HandleFunc("/api/table/{id}/delete", deleteTable).Methods("DELETE")
	r.HandleFunc("/api/table/{id}/diff", diffTable).Methods("POST")
	r.HandleFunc("/api/table/diff", diffAllTables)
	r.HandleFunc("/api/table/list/{start}/{count}", listTables)
}

type DeleteRequest struct {
	Table string
}

type DeleteResponse struct {
	Details string
}

var yamlPath string

func tableSetup(conf config.Config) (err error) {

	// Put Metadata into Cache Mode, otherwise the server is going to be quite slow.
	// This is not super necessary as the server doesn't need to handle high RPS.
	metadata.UseCache(true)

	// Read the YAML schema
	yamlPath = util.WorkingSubDir(strings.ToLower(conf.Project.Name))

	util.LogInfo(yamlPath)

	_, err = util.DirExists(yamlPath)
	if util.ErrorCheck(err) {
		return fmt.Errorf("Table Setup failed. Unable to read Local Schema Path")
	}

	// Read tables relative to the current working directory (which is the project name)
	err = yaml.ReadTables(strings.ToLower(conf.Project.Name))

	if util.ErrorCheck(err) {
		return fmt.Errorf("Table Setup failed. Unable to read YAML Tables")
	}

	util.LogInfof("Found %d YAML Tables", len(yaml.Schema))

	return err
}

func createTable(w http.ResponseWriter, r *http.Request) {
	var newTable table.Table
	var err error
	var errors id.ValidationErrors

	// Parse table from post body
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&newTable)
	if err != nil {
		writeErrorResponse(w, r, "Create FAILED! Malformed Table.", err, nil)
	}

	// Generate any missing PropertyIDs
	newTable.GeneratePropertyIDs()

	// check for table errors, or name conflicts
	// insert table into temporary yaml table slice and validate
	tmpTables := yaml.Schema[:]
	tmpTables = append(tmpTables, newTable)

	errors, err = id.ValidateSchema(tmpTables, "YAML Schema", false)
	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, "Create FAILED! YAML Validation Errors Detected", err, errors)
		return
	}

	// Serialise table to disk
	err = yaml.WriteTable(yamlPath, newTable)

	if err != nil {
		writeErrorResponse(w, r, "Create FAILED! Unable to create YAML Table definition", err, errors)
		return
	}

	// insert into the yaml.Schema array
	yaml.Schema = append(yaml.Schema, newTable)

	// Return Success!
	var response struct {
		Result string
		Table  table.Table
	}

	response.Result = "success"
	response.Table = newTable

	writeResponse(w, response, err)
}

// getTable Get Table by Id
func getTable(w http.ResponseWriter, r *http.Request) {
	var err error
	var t *table.Table
	vars := mux.Vars(r)

	tableID := vars["id"]

	for i, tbl := range yaml.Schema {
		if tbl.Metadata.PropertyID == tableID {
			t = &yaml.Schema[i]
			break
		}
	}

	if t != nil {
		writeResponse(w, t, err)
		return
	}
	writeErrorResponse(w, r, fmt.Sprintf("Table with PropertyID: %s not found.", tableID), nil, nil)
}

func editTable(w http.ResponseWriter, r *http.Request) {

}

func deleteTable(w http.ResponseWriter, r *http.Request) {
	var deleteRequest DeleteRequest
	// Parse table from post body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&deleteRequest)
	if err != nil {
		writeErrorResponse(w, r, "Unable to delete Table", err, nil)
		return
	}

	// Check that the table is in the YAML Schema
	tableExists := false
	for i := 0; i < len(yaml.Schema); i++ {
		if yaml.Schema[i].Name == deleteRequest.Table {
			tableExists = true
			break
		}
	}

	if !tableExists {
		writeErrorResponse(w, r, "Unable to delete non-existant Table.  Unknown Table.", err, nil)
		return
	}

	filename := fmt.Sprintf("%s.yml", deleteRequest.Table)
	fp := filepath.Join(yamlPath, filename)

	// Check that the file exists

	fileExists, err := util.FileExists(fp)

	if !fileExists || err != nil {
		writeErrorResponse(w, r, "Unable to delete non-existant Table.  File doesn't exist.", err, nil)
		return
	}

	if tableExists && fileExists {
		// Delete the table YAML file
		err = util.DeleteFile(fp)

		if err != nil {
			writeErrorResponse(w, r, "Unable to delete Table", err, nil)
			return
		}

		// Remove from the YAML Schema array
		removeSuccess := false
		for i := 0; i < len(yaml.Schema); i++ {
			if yaml.Schema[i].Name == deleteRequest.Table {
				yaml.Schema = append(yaml.Schema[:i], yaml.Schema[i+1:]...)
				removeSuccess = true
				break
			}
		}

		if !removeSuccess {
			writeErrorResponse(w, r, "Unable to delete Table from Schema", err, nil)
			return
		}
	}

	writeResponse(w, DeleteResponse{
		Details: "Successfully Deleted",
	}, err)

}

func diffTable(w http.ResponseWriter, r *http.Request) {

}

// TableDiffList Helper type used to return a List of Table Diffs
type TableDiffList struct {
	// Migrations []migration.Migration `json:"migrations"`
	Start int64 `json:"start"`
	End   int64 `json:"end"`
	Total int64 `json:"total"`
}

func diffAllTables(w http.ResponseWriter, r *http.Request) {

}

// TableList Helper type used to return a List of Tables
type TableList struct {
	Tables []table.Table `json:"tables"`
	Start  int64         `json:"start"`
	End    int64         `json:"end"`
	Total  int64         `json:"total"`
}

// listTables List all tables
func listTables(w http.ResponseWriter, r *http.Request) {
	var start int64
	var count int64
	var end int64
	var total int64
	var err error

	// Setting default parameters
	start = 0
	count = 10

	vars := mux.Vars(r)

	numTables := int64(len(yaml.Schema))
	var iValue int64

	for key, value := range vars {

		switch key {

		case "start":
			iValue, err = strconv.ParseInt(value, 10, 64)
			start = iValue

		case "count":
			iValue, err = strconv.ParseInt(value, 10, 64)
			count = iValue
		}

		if util.ErrorCheck(err) {
			writeErrorResponse(w, r, fmt.Sprintf("Unable to parse. Param: %s value: %s", key, value), err, nil)
			return
		}

		if iValue > numTables {
			writeErrorResponse(w, r, fmt.Sprintf("Invalid offset. Param: %s value: %s ", key, value), err, nil)
			return
		}
	}

	if numTables < start+count {
		writeErrorResponse(w, r, fmt.Sprintf("Invalid parameters. Out of bounds Start: %d Count: %d ", start, count), err, nil)
		return
	}

	listTables := yaml.Schema[start : start+count]

	payload := TableList{
		Tables: listTables,
		Start:  start,
		End:    end,
		Total:  total,
	}

	writeResponse(w, payload, err)
}
