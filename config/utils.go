package config

import (
	"encoding/json"
	"log"
	"net/http"
)

var Logger *log.Logger

// Info will log information with "INFO" prefix to logger
func Info(args ...interface{}) {
	Logger.SetPrefix("INFO ")
	Logger.Println(args...)
}

// Danger will log information with "ERROR" prefix to logger
func Danger(args ...interface{}) {
	Logger.SetPrefix("ERROR ")
	Logger.Println(args...)
}

// Warning will log information with "WARNING" prefix to logger
func Warning(args ...interface{}) {
	Logger.SetPrefix("WARNING ")
	Logger.Println(args...)
}

// ReportStatus is a helper function to return a JSON response indicating outcome success/failure
func ReportStatus(w http.ResponseWriter, success bool, err *APIError) {
	var res *Outcome
	w.Header().Set("Content-Type", "application/json")
	if success {
		res = &Outcome{
			Status: success,
		}
	} else {
		res = &Outcome{
			Status: success,
			Error:  err,
		}
	}
	response, _ := json.Marshal(res)
	if _, err := w.Write(response); err != nil {
		Danger("Error writing", response)
	}
}
