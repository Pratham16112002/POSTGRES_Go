package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(res http.ResponseWriter, status int, data any) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	return json.NewEncoder(res).Encode(data)
}

func readJSON(res http.ResponseWriter, req *http.Request, data any) error {
	maxBytes := 1_048_578
	req.Body = http.MaxBytesReader(res, req.Body, int64(maxBytes))
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

func writeJSONError(res http.ResponseWriter, status int, message string) error {
	type errorJSON struct {
		Error string `json:"error"`
	}
	return writeJSON(res, status, &errorJSON{Error: message})
}
