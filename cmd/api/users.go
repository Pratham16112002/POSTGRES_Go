package main

import (
	"Blog/internal/store"
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

func (app *application) getUserHandler(res http.ResponseWriter, req *http.Request) {
	user := getUserFromCtx(req)
	if err := app.jsonResponse(res, http.StatusFound, user); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}

type FollowUser struct {
	UserId int64 `json:"user_id"`
}

func (app *application) followUserHandler(res http.ResponseWriter, req *http.Request) {
	// : TODO get the followee from authentication context
	follower := getUserFromCtx(req)
	var paylaod FollowUser
	if err := readJSON(res, req, &paylaod); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	ctx := req.Context()
	err := app.store.Followers.Follow(ctx, paylaod.UserId, follower.ID)
	if err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictError(res, req, err)
		default:
			app.internalServerError(res, req, err)
		}
		return
	}
	if err := app.jsonResponse(res, http.StatusNoContent, struct{}{}); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}
func (app *application) unfollowUserHandler(res http.ResponseWriter, req *http.Request) {
	// : TODO get the followee from authentication context
	follower := getUserFromCtx(req)
	var paylaod FollowUser
	if err := readJSON(res, req, &paylaod); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	ctx := req.Context()
	err := app.store.Followers.Unfollow(ctx, paylaod.UserId, follower.ID)
	if err != nil {
		app.internalServerError(res, req, err)
		return
	}
	if err := app.jsonResponse(res, http.StatusNoContent, struct{}{}); err != nil {
		app.internalServerError(res, req, err)
		return
	}

}

func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		userId, err := strconv.ParseInt(chi.URLParam(req, "userId"), 10, 64)
		if err != nil {
			app.badRequestError(res, req, err)
			return
		}
		ctx := req.Context()
		user, err := app.store.Users.GetUserById(ctx, userId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(res, req, err)
			default:
				app.internalServerError(res, req, err)
			}
			return
		}
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func getUserFromCtx(req *http.Request) *store.User {
	user, _ := req.Context().Value(userCtx).(*store.User)
	return user
}
