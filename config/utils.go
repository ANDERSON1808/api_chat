package handler

import (
	"api_chat/config"
	"encoding/json"
	"log"
	"net/http"
)

var logger *log.Logger

/* Convenience function for printing to stdout
func p(a ...interface{}) {
	fmt.Println(a...)
}*/

// Info will log information with "INFO" prefix to logger
func Info(args ...interface{}) {
	logger.SetPrefix("INFO ")
	logger.Println(args...)
}

// Danger will log information with "ERROR" prefix to logger
func Danger(args ...interface{}) {
	logger.SetPrefix("ERROR ")
	logger.Println(args...)
}

// Warning will log information with "WARNING" prefix to logger
func Warning(args ...interface{}) {
	logger.SetPrefix("WARNING ")
	logger.Println(args...)
}

// ReportStatus is a helper function to return a JSON response indicating outcome success/failure
func ReportStatus(w http.ResponseWriter, success bool, err *config.APIError) {
	var res *config.Outcome
	w.Header().Set("Content-Type", "application/json")
	if success {
		res = &config.Outcome{
			Status: success,
		}
	} else {
		res = &config.Outcome{
			Status: success,
			Error:  err,
		}
	}
	response, _ := json.Marshal(res)
	if _, err := w.Write(response); err != nil {
		Danger("Error writing", response)
	}
}
