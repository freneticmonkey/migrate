package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/freneticmonkey/migrate/go/id"
	"github.com/freneticmonkey/migrate/go/sandbox"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/gorilla/mux"
)

// registerTableEndpoints Register the table functions for the REST API
func registerTableEndpoints(r *mux.Router) {
	r.HandleFunc("/api/table/create", createTable).Methods("PUT")
	r.HandleFunc("/api/table/{id}", getTable)
	r.HandleFunc("/api/table/{id}/edit", editTable).Methods("POST")
	r.HandleFunc("/api/table/{id}/delete", deleteTable).Methods("DELETE")
	r.HandleFunc("/api/table/list/", listTables)
	r.HandleFunc("/api/table/list/{start}", listTables)
	r.HandleFunc("/api/table/list/{start}/{count}", listTables)
}

type DeleteRequest struct {
	Table string
}

type DeleteResponse struct {
	Details string
}

type DiffRequest struct {
	Table string
}

func replaceTable(context string, w http.ResponseWriter, r *http.Request, tbl table.Table) {
	var errors id.ValidationErrors
	var err error

	// Generate any missing PropertyIDs
	tbl.GeneratePropertyIDs()

	// check for table errors, or name conflicts
	// insert table into temporary yaml table slice and validate
	tmpTables := yaml.Schema[:]
	tmpTables = append(tmpTables, tbl)

	errors, err = id.ValidateSchema(tmpTables, "YAML Schema", false)
	if util.ErrorCheck(err) {
		writeErrorResponse(w, r, fmt.Sprintf("%s FAILED! YAML Validation Errors Detected", context), err, errors)
		return
	}

	// Serialise table to disk
	err = yaml.WriteTable(yamlPath, tbl)

	// Serialise to template file
	err = sandbox.GenerateTable(conf, tbl)

	if err != nil {
		writeErrorResponse(w, r, fmt.Sprintf("%s FAILED! Unable to create YAML Table definition", context), err, errors)
		return
	}

	// insert into the yaml.Schema array
	yaml.Schema = append(yaml.Schema, tbl)

	// Return Success!
	var response struct {
		Result string
		Table  table.Table
	}

	response.Result = "success"
	response.Table = tbl

	writeResponse(w, response, err)
}

func tableExists(tableName string) bool {
	// Check that the table is in the YAML Schema
	for i := 0; i < len(yaml.Schema); i++ {
		if yaml.Schema[i].Name == tableName {
			return true
		}
	}
	return false
}

func createTable(w http.ResponseWriter, r *http.Request) {
	var newTable table.Table
	var err error

	// Parse table from post body
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&newTable)
	if err != nil {
		writeErrorResponse(w, r, "Create Table FAILED! Malformed Table.", err, nil)
	} else {
		replaceTable("Create", w, r, newTable)
	}
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
	var editTable table.Table
	var err error
	vars := mux.Vars(r)

	// Parse table from post body
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&editTable)
	if err != nil {
		writeErrorResponse(w, r, "Edit Table FAILED! Malformed Table.", err, nil)
		return
	}

	// Verify that the table is valid
	tableID := vars["id"]

	if editTable.ID != tableID {
		writeErrorResponse(w, r, "Edit Table FAILED! Table ID mismatch.", err, nil)
		return
	}

	if !tableExists(tableID) {
		writeErrorResponse(w, r, "Edit Table FAILED! Unknown table.", err, nil)
		return
	}

	replaceTable("Edit", w, r, editTable)
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
	if !tableExists(deleteRequest.Table) {
		writeErrorResponse(w, r, "Delete Table FAILED! Unknown table.", err, nil)
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

	writeResponse(w, DeleteResponse{
		Details: "Successfully Deleted",
	}, err)
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

	until := start + count

	if until > int64(len(yaml.Schema)-1) {
		until = int64(len(yaml.Schema) - 1)
	}

	listTables := yaml.Schema[start:until]

	payload := TableList{
		Tables: listTables,
		Start:  start,
		End:    end,
		Total:  total,
	}

	writeResponse(w, payload, err)
}
