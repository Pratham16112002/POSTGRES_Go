package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/resend/resend-go/v2"
)

type ResendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *resend.Client
}

func NewResend(apikey, fromEmail string) *ResendGridMailer {
	client := resend.NewClient(apikey)
	return &ResendGridMailer{
		apiKey:    apikey,
		fromEmail: fromEmail,
		client:    client,
	}
}

func (m *ResendGridMailer) Send(templateFile, username, email string, data any) error {

	templ, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)

	err = templ.ExecuteTemplate(body, "body", data)

	params := &resend.SendEmailRequest{
		To:      []string{email},
		From:    m.fromEmail,
		Subject: "Activation Code",
		Html:    body.String(),
	}
	if err != nil {
		return err
	}

	for i := 0; i < MaxRetries; i++ {
		sent, err := m.client.Emails.Send(params)
		if err != nil {
			log.Printf("Failed to send email to %v, attempt %d of %d\n", email, i+1, MaxRetries)
			log.Printf("%s", err.Error())
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("Email sent to %v  succcessfully with %v\n", email, sent.Id)
		return nil
	}

	return fmt.Errorf("failed to sent email")
}
