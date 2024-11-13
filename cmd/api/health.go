package main

import "net/http"

func (app *application) healthCheckHandler(res http.ResponseWriter, req *http.Request) {
	// res.Write([]byte("Welcome to the heal check handler changed\n"))
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.addr,
		"version": "yahoo",
	}
	if err := writeJSON(res, http.StatusOK, data); err != nil {
		writeJSONError(res, http.StatusInternalServerError, "err.Error()")
	}
}
