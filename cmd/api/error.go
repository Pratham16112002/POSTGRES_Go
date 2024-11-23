package main

import (
	"net/http"
)

func (app *application) internalServerError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnf("internal server error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusInternalServerError, "server encountered a problem")
}
func (app *application) badRequestError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnf("bad request error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnf("already exist error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusConflict, "already exist")
}

func (app *application) notFoundError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnf("not found error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusNotFound, "not found")
}

func (app *application) authorizationError(res http.ResponseWriter, req *http.Request, err error) {

	app.logger.Warnf("user not authorized", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusNotFound, "not authorized")
}
