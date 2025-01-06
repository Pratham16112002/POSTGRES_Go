package main

import (
	"Blog/internal/store"
	"Blog/internal/store/paginate"
	"net/http"
)

func (app *application) getUserSearchFriend(res http.ResponseWriter, req *http.Request) {
	fq := &paginate.FriendPaginateQuery{}
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
	list, err := app.store.Users.SearchFriends(ctx, user.ID, fq)
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
	if err := app.jsonResponse(res, http.StatusOK, list); err != nil {
		app.internalServerError(res, req, err)
		return
	}

}
