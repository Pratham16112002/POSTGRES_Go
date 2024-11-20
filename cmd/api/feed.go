package main

import (
	"Blog/internal/store"
	"net/http"
)

func (app *application) getUserFeedHandler(res http.ResponseWriter, req *http.Request) {
	// pagination , serach , filters
	ctx := req.Context()
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(1))
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(res, req, err)
			return
		default:
			return
		}
	}
	if err := app.jsonResponse(res, http.StatusOK, feed); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}
