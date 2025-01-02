package main

import (
	"Blog/internal/store"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Content string   `json:"content" validate:"omitempty,max=1000"`
	Title   string   `json:"title" validate:"omitempty,max=50"`
	Tags    []string `json:"tags"`
}

type CreateCommentPaylaod struct {
	Content string `json:"content" validate:"max=200,min=3"`
}
type DeletedPostPayload struct {
	ID int64 `json:"content"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) getPostHanlder(res http.ResponseWriter, req *http.Request) {
	post := getPostFromCtx(req)
	comments, err := app.store.Comments.GetByPostID(req.Context(), post.ID)
	if err != nil {
		app.internalServerError(res, req, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(res, http.StatusOK, post); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}

func (app *application) updatePostHandler(res http.ResponseWriter, req *http.Request) {
	post := getPostFromCtx(req)
	var payload UpdatePostPayload
	if err := readJSON(res, req, &payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	if err := validate.Struct(payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	ctx := req.Context()

	err := app.store.Posts.Update(ctx, post)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(res, req, err)
		default:
			app.badRequestError(res, req, err)
		}
		return
	}
	if err := app.jsonResponse(res, http.StatusOK, payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}
}

func (app *application) deletePostHandler(res http.ResponseWriter, req *http.Request) {

	post := getPostFromCtx(req)

	ctx := req.Context()

	err := app.store.Posts.Delete(ctx, post.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundError(res, req, err)
			return
		default:
			app.internalServerError(res, req, err)
			return
		}
	}
	res.WriteHeader(http.StatusNoContent)
}

func (app *application) createPostHandler(res http.ResponseWriter, req *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(res, req, &payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	err := validate.Struct(payload)
	if err != nil {
		app.badRequestError(res, req, err)
		return
	}
	user := getAuthUser(req)
	post := &store.Post{
		UserId:  user.ID,
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
	}
	ctx := req.Context()
	err = app.store.Posts.Create(ctx, post)
	if err != nil {
		app.internalServerError(res, req, err)
		return
	}
	if err := app.jsonResponse(res, http.StatusCreated, post); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}

func (app *application) postCommentHandler(res http.ResponseWriter, req *http.Request) {
	post := getPostFromCtx(req)
	var payload CreateCommentPaylaod
	if err := readJSON(res, req, &payload); err != nil {
		app.internalServerError(res, req, err)
		return
	}
	if err := validate.Struct(payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	user := getAuthUser(req)
	var comment store.Comment
	comment.Content = payload.Content
	comment.UserID = user.ID
	comment.PostID = post.ID
	ctx := req.Context()
	err := app.store.Comments.Create(ctx, &comment)
	if err != nil {
		app.internalServerError(res, req, err)
		return
	}
	if err := app.jsonResponse(res, http.StatusOK, comment); err != nil {
		app.internalServerError(res, req, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		postId, err := strconv.ParseInt(chi.URLParam(req, "postId"), 10, 64)
		if err != nil {
			app.internalServerError(res, req, err)
			return
		}

		ctx := req.Context()
		post, err := app.store.Posts.GetById(ctx, postId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(res, req, err)
			default:
				app.internalServerError(res, req, err)
			}
			return
		}
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func getPostFromCtx(req *http.Request) *store.Post {
	post, _ := req.Context().Value(postCtx).(*store.Post)
	return post
}
