package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("internal server error : %s : path %s : err : %s", req.Method, req.URL.Path, err.Error())
	writeJSONError(res, http.StatusInternalServerError, "server encountered a problem")
}
func (app *application) badRequestError(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("bad request error : %s :  path : %s err : %s", req.Method, req.URL.Path, err.Error())
	writeJSONError(res, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("already exist : %s : path %s : err : %s", req.Method, req.URL.Path, err.Error())
	writeJSONError(res, http.StatusConflict, "already exist")
}

func (app *application) notFoundError(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("not found error : %s : path : %s err : %s", req.Method, req.URL.Path, err.Error())
	writeJSONError(res, http.StatusNotFound, "not found")
}
