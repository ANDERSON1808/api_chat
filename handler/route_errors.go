package handler

import (
	"api_chat/config"
	"net/http"
	"strings"
)

type ErrHandler func(http.ResponseWriter, *http.Request) error

func (fn ErrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		if apierr, ok := err.(*config.APIError); ok {
			w.Header().Set("Content-Type", "application/json")
			apierr.SetMsg()
			config.Warning("API error:", apierr.Error())
			if apierr.Code == 101 || apierr.Code == 201 {
				notFound(w, r)
			} else if apierr.Code == 102 || apierr.Code == 202 || apierr.Code == 303 || apierr.Code == 105 {
				badRequest(w, r)
			} else if apierr.Code == 104 || apierr.Code == 204 || apierr.Code == 304 || apierr.Code == 401 || apierr.Code == 402 {
				unauthorized(w, r)
			} else if apierr.Code == 403 {
				forbidden(w, r)
			} else {
				badRequest(w, r)
			}
			config.ReportStatus(w, false, apierr)
		} else {
			config.Danger("Server error", err.Error())
			http.Error(w, err.Error(), 500)
		}
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	config.Info("Not found request:", r.RequestURI)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	config.Info("forbidden:", r.RequestURI, r.Body)
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	config.Warning("forbidden:", r.RequestURI, r.Body)
}

func badRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	config.Info("Bad request:", r.RequestURI, r.Body)
}

// Convenience function to redirect to the error message page
func errorMessage(writer http.ResponseWriter, request *http.Request, msg string) {
	url := []string{"/err?msg=", msg}
	http.Redirect(writer, request, strings.Join(url, ""), 302)
}
