package main

import (
	"Blog/internal/store"
	"fmt"
	"net/http"
	"time"
)

type UserFeedPayload struct {
	UserId int64 `json:"user_id"`
}

func (app *application) getUserFeedHandler(res http.ResponseWriter, req *http.Request) {
	// pagination , serach , filters
	fq := store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 20,
		Sort:   "desc",
		Tags:   make([]string, 0),
		Since:  store.ParseTime("1990-11-21T11:26:57+05:30"),
		Until:  time.Now().Format(time.RFC3339),
	}
	fq, err := fq.Parse(req)
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
	fmt.Println(feed)
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
