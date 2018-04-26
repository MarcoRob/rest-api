package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
	"github/godspeedkil/rest-api"
	"fmt"
	"encoding/json"
)

const DECIMAL_BASE = 10
const INT64_BITS = 64

func main() {
	registerHandlers()
}

func registerHandlers() {
	router := mux.NewRouter()

	router.Methods("GET").Path("/users/{userId}/reports").
		Handler(appHandler(createHandler))
	router.Methods("GET").Path("/users/reports/{reportId}").
		Handler(appHandler(getByIDHandler))

	log.Fatal(http.ListenAndServe(":8000", router))
}

func getByIDHandler(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	vars := mux.Vars(r)
	reportId, err := strconv.ParseInt(vars["reportId"], DECIMAL_BASE, INT64_BITS)
	if err != nil {
		return appErrorf(err, "could not parse reportId from request: %v", err)
	}
	report, err := user_report.DB.GetUserReport(reportId)
	if err != nil {
		return appErrorf(err, "could not get report: %v", err)
	}
	json.NewEncoder(w).Encode(report)
	return nil
}

func createHandler(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	vars := mux.Vars(r)
	userId := vars["userId"]

	report, err := user_report.GenerateUserReport(userId)
	if err != nil {
		return appErrorf(err, "could not generate user report: %v", err)
	}

	reportId, err := user_report.DB.AddUserReport(&report)
	http.Redirect(w, r, fmt.Sprintf("/users/reports/%d", reportId),
		http.StatusFound)
	return nil
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	if e := fn(w, r); e != nil {
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
