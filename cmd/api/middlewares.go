package main

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			// read the auth header
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				app.authorizationError(res, req, errors.New("user not authorized"))
				return
			}
			// decode it
			authHeader_parts := strings.Split(authHeader, " ")
			if len(authHeader_parts) != 2 || authHeader_parts[0] != "Basic" {
				app.authorizationError(res, req, errors.New("auth token not found"))
				return
			}
			decoded, err := base64.StdEncoding.DecodeString(authHeader_parts[1])
			if err != nil {
				app.authorizationError(res, req, err)
				return
			}
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 {

			}
			// check the creadential
			next.ServeHTTP(res, req)
		})
	}

}
