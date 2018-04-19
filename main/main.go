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
const (
	TODAY_TASKS = 0
	DELAYED_TASKS = -1
)

func main() {
	registerHandlers()
}

func registerHandlers() {
	router := mux.NewRouter()

	router.Methods("GET").Path("/report/user/{userId}").
		Handler(appHandler(createHandler))
	router.Methods("GET").Path("/report/user/{userId}/{reportId}").
		Handler(appHandler(getByIDHandler))

	log.Fatal(http.ListenAndServe(":8000", router))
}

func getByIDHandler(w http.ResponseWriter, r *http.Request) *appError {
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
	vars := mux.Vars(r)
	userId, err := strconv.ParseInt(vars["userId"], DECIMAL_BASE, INT64_BITS)
	if err != nil {
		return appErrorf(err, "could not parse userId from request: %v", err)
	}

	delayedTasks, err := user_report.ExtractTaskType(DELAYED_TASKS)
	todayTasks, err := user_report.ExtractTaskType(TODAY_TASKS)
	goodHabits, err := user_report.ExtractHabitType("good")
	badHabits, err := user_report.ExtractHabitType("bad")

	report := &user_report.UserReport{UserID:userId, TasksToday:todayTasks,
		TasksDelayed:delayedTasks, HabitsGood:goodHabits, HabitsBad:badHabits}
	reportId, err := user_report.DB.AddUserReport(report)
	http.Redirect(w, r, fmt.Sprintf("/report/user/%d/%d", userId, reportId),
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
