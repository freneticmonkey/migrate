package serve

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/freneticmonkey/migrate/go/util"
	"github.com/gorilla/mux"
)

// ResponseError Standardised response error helper struct
type ResponseError struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// Response Standardised response helper struct
type Response struct {
	Result interface{}   `json:"result"`
	Error  ResponseError `json:"error"`
}

// Run Start the REST API Server
func Run(frontend bool, port int) (err error) {
	r := mux.NewRouter()

	// Register API endpoints
	RegisterMigrationEndpoints(r)
	RegisterStatusEndpoints(r)
	RegisterDatabaseEndpoints(r)

	// Serve the Javascript Frontend UI as well
	if frontend {
		log.Printf("Serving static")
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	}

	http.Handle("/", r)
	log.Printf("Server started on port: %d\n", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	return err
}

// WriteResponse Helper function for building a standardised JSON response
func WriteResponse(w http.ResponseWriter, body interface{}, e error) (err error) {
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

// WriteErrorResponse Helper function for building a standardised JSON error response
func WriteErrorResponse(w http.ResponseWriter, detail string, e error) (err error) {
	var response []byte
	var mt []byte
	response, err = json.Marshal(Response{
		Result: mt,
		Error: ResponseError{
			Error:  fmt.Sprintf("%v", e),
			Detail: detail,
		},
	})

	if !util.ErrorCheck(err) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", response)
	}
	return err
}
