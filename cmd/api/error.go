package main

import (
	"net/http"
)

func (app *application) internalServerError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("internal server error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusInternalServerError, "server encountered a problem")
}
func (app *application) badRequestError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("bad request error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusBadRequest, "bad request")
}

func (app *application) conflictError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("already exist error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusConflict, "already exist")
}

func (app *application) notFoundError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("not found error", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusNotFound, "not found")
}
func (app *application) basicAuthorizationError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("user not authorized", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	res.Header().Add("WWW-Authenticate", `Basic realm="restricted" charset="UTF-8"`)
	writeJSONError(res, http.StatusUnauthorized, "not authorized")
}

func (app *application) authorizationError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("user not authorized", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusUnauthorized, "not authorized")
}

func (app *application) forbiddenError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("forbidden", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusForbidden, "forbidden")
}

func (app *application) authenticationError(res http.ResponseWriter, req *http.Request, err error) {
	app.logger.Warnw("authentication failed", err, "path", req.URL.Path, "method", req.Method, "message", err.Error())
	writeJSONError(res, http.StatusInternalServerError, "oops, redo authentication")
}

func (app *application) rateLimitExceededResponse(res http.ResponseWriter, req *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "path", req.URL.Path, "method", req.Method)
	res.Header().Set("Retry-After", retryAfter)
	writeJSONError(res, http.StatusTooManyRequests, "rate limit exceeded please try after : "+retryAfter)
}
