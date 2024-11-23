package main

import (
	"Blog/internal/mailer"
	"Blog/internal/store"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,email,max=50"`
	Password string `json:"password" validate:"required,min=3,max=88"`
}

// type UserWithToken struct {
// 	user  *store.User
// 	token string
// }

func (app *application) userRegisterHandler(res http.ResponseWriter, req *http.Request) {
	var payload RegisterUserPayload
	// validation of request body
	if err := readJSON(res, req, &payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}

	// validation of payload
	if err := validate.Struct(payload); err != nil {
		app.badRequestError(res, req, err)
		return
	}
	user := &store.User{}
	user.Username = payload.Username
	user.Email = payload.Email

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(res, req, err)
		return
	}
	ctx := req.Context()
	// creation of unique token for user activation
	token := uuid.New().String() // Creation of user and user invitation
	fmt.Println(token)
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])
	err := app.store.Users.CreateAndInvite(ctx, user, hashedToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestError(res, req, err)
		case store.ErrDuplicateUsername:
			app.badRequestError(res, req, err)
		default:
			app.internalServerError(res, req, err)
		}
		return
	}
	// Sending email
	isProdEnv := app.config.env == "production"
	// userWithToken := UserWithToken{
	// 	user:  user,
	// 	token: token,
	// }
	app.logger.Infow("token is : ", token)
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, token)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	email_status_code, err := app.mailer.Send(mailer.UserActivationTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		if email_status_code != -1 {
			if err := app.store.Users.Delete(ctx, user.ID); err != nil {
				app.internalServerError(res, req, err)
				return
			}
		}
	}

	if err := app.jsonResponse(res, http.StatusNoContent, ""); err != nil {
		app.internalServerError(res, req, err)
		return
	}

}
