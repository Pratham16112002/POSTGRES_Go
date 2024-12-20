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

const ctxUser userKey = "user"

func (app *application) getUserHandler(res http.ResponseWriter, req *http.Request) {
	user := getUserFromContext(req)
	if err := app.jsonResponse(res, http.StatusFound, user); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}

func (app *application) followUserHandler(res http.ResponseWriter, req *http.Request) {
	follower := getAuthUser(req)
	followee := getUserFromContext(req)
	ctx := req.Context()
	err := app.store.Followers.Follow(ctx, followee.ID, follower.ID)
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
	follower := getUserFromContext(req)
	followee := getUserFromContext(req)
	ctx := req.Context()
	err := app.store.Followers.Unfollow(ctx, followee.ID, follower.ID)
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
		ctx = context.WithValue(ctx, ctxUser, user)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func getAuthUser(req *http.Request) *store.User {
	user, _ := req.Context().Value(authUser).(*store.User)
	return user
}

func (app *application) userActivationHandler(res http.ResponseWriter, req *http.Request) {
	token := chi.URLParam(req, "token")
	ctx := req.Context()
	err := app.store.Users.Activate(ctx, token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(res, req, err)
		default:
			app.internalServerError(res, req, err)
		}
		return
	}
	if err := app.jsonResponse(res, http.StatusNoContent, ""); err != nil {
		app.internalServerError(res, req, err)
		return
	}

}
func getUserFromContext(req *http.Request) *store.User {
	user, _ := req.Context().Value(ctxUser).(*store.User)
	return user
}
