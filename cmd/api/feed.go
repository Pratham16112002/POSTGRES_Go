package main

import (
	"Blog/internal/store"
	"Blog/internal/store/paginate"
	"net/http"
)

type UserFeedPayload struct {
	UserId int64 `json:"user_id"`
}

func (app *application) getUserFeedHandler(res http.ResponseWriter, req *http.Request) {
	// pagination , serach , filters
	fq := &paginate.PostPaginateQuery{}
	err := fq.Parse(req)
	if err != nil {
		app.badRequestError(res, req, err)
		return
	}
	if err := validate.Struct(fq); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	ctx := req.Context()
	user := getAuthUser(req)
	feed, err := app.store.Posts.GetUserFeed(ctx, user.ID, fq)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(res, req, err)
			return
		default:
			app.internalServerError(res, req, err)
			return
		}
	}
	if err := app.jsonResponse(res, http.StatusOK, feed); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}
