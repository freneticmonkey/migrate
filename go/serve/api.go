package serve

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
	"github.com/gorilla/mux"
)

// ResponseError Standardised response error helper struct
type ResponseError struct {
	Error  string      `json:"error"`
	Detail string      `json:"detail"`
	Data   interface{} `json:"data"`
}

// Response Standardised response helper struct
type Response struct {
	Result interface{}   `json:"result"`
	Error  ResponseError `json:"error"`
}

var yamlPath string
var conf config.Config

func setupServer(apiConfig config.Config) (err error) {

	conf = apiConfig

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

// Run Start the REST API Server
func Run(apiConfig config.Config, frontend bool, port int) (err error) {
	util.LogInfo("Starting Migrate Server")

	// Configuring server cache
	err = setupServer(apiConfig)
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	// Register API endpoints
	registerMigrationEndpoints(r)
	registerStatusEndpoints(r)
	registerDatabaseEndpoints(r)
	registerTableEndpoints(r)
	registerSandboxEndpoints(r)
	registerHealthEndpoints(r)

	// Serve the Javascript Frontend UI as well
	if frontend {
		log.Printf("Serving static")
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	}

	http.Handle("/", r)
	log.Printf("Migrate Server started on port: %d\n", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	return err
}

// writeResponse Helper function for building a standardised JSON response
func writeResponse(w http.ResponseWriter, body interface{}, e error) (err error) {
	var payload []byte
	response := Response{
		Result: body,
		Error:  ResponseError{},
	}
	if e != nil {
		response.Error.Error = fmt.Sprintf("%v", e)
	}
	payload, err = json.MarshalIndent(response, "", "\t")

	if !util.ErrorCheck(err) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	}
	return err
}

// writeErrorResponse Helper function for building a standardised JSON error response
func writeErrorResponse(w http.ResponseWriter, r *http.Request, detail string, e error, errorData interface{}) (err error) {
	var response []byte
	var mt []byte

	util.LogErrorf("URL: '%s' Error during request: %v", r.URL, e)

	response, err = json.Marshal(Response{
		Result: mt,
		Error: ResponseError{
			Error:  fmt.Sprintf("%v", e),
			Detail: detail,
			Data:   errorData,
		},
	})

	if !util.ErrorCheck(err) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", response)
	}
	return err
}
