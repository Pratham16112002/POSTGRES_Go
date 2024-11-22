package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendgrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)
	return &SendGridMailer{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	// template parsing and building
	subject := new(bytes.Buffer)

	tmpl.ExecuteTemplate(subject, "subject", data)

	body := new(bytes.Buffer)

	tmpl.ExecuteTemplate(body, "body", data)

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})
	for i := 0; i < MaxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send emial to %v, attempt %d of %d", email, i+1, MaxRetries)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("Email sent with status code %v", response.StatusCode)
		return nil
	}
	return fmt.Errorf("failed to send email after %d attemps", MaxRetries)
}
