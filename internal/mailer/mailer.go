package mailer

import (
	"embed"
)

const (
	FromName               = "BloggerSpot"
	MaxRetries             = 3
	UserActivationTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) (int64, error)
}
