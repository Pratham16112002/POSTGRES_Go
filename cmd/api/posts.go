package main

import (
	"Blog/internal/store"
	"net/http"
)

type CreatePostPayload struct {
	Content string `json:"content"`
	Title string `json:"title"`
	UserId int64 `json:"user_id"`
	Tags []string `json:"tags"`
}

func (app *application) createPostHandler(res http.ResponseWriter, req *http.Request) {
	var post 
	if err := readJSON(res, req, post); err != nil {
		writeJSONError(res, http.StatusBadRequest, err.Error())
		return
	}
	userID := 1
	ctx := req.Context()
	app.store.Posts.Create()

}
