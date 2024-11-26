package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"text/template"
)

type NewMailerType struct {
	fromEmail string
	apiKey    string
}

type MailPayload struct {
	From          string `json:"From" validate:"email"`
	To            string `json:"To" validate:"email"`
	Subject       string `json:"Subject" validate:"max=50"`
	TextBody      string `json:"Textbody"`
	HtmlBody      string `json:"HtmlBody"`
	MessageStream string `json:"MessageStream"`
}

func NewMailer(apiKey, fromEmail string) *NewMailerType {
	return &NewMailerType{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}
}

func (m *NewMailerType) Send(templateFile, username, email string, data any, isSandbox bool) (int64, error) {
	var payload MailPayload
	payload.From = m.fromEmail
	payload.To = email
	payload.TextBody = fmt.Sprintf("Hi, %s", username)
	payload.MessageStream = "outbound"
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}
	// template parsing and building
	subject := new(bytes.Buffer)

	tmpl.ExecuteTemplate(subject, "subject", data)

	body := new(bytes.Buffer)

	tmpl.ExecuteTemplate(body, "body", data)

	payload.Subject = subject.String()
	payload.HtmlBody = body.String()

	var data_payload_writer bytes.Buffer
	if err := json.NewEncoder(&data_payload_writer).Encode(&payload); err != nil {
		return -1, err
	}

	client := &http.Client{}
	ext_req, err := http.NewRequest(http.MethodPost, "https://api.postmarkapp.com/email", &data_payload_writer)
	if err != nil {
		return -1, err
	}
	ext_req.Header.Add("Accept", "application/json")
	ext_req.Header.Add("Content-Type", "application/json")
	ext_req.Header.Add("X-Postmark-Server-Token", m.apiKey)

	res, err := client.Do(ext_req)
	if err != nil || res.StatusCode != http.StatusOK {
		return http.StatusInternalServerError, errors.New("failed to send invite, server error")
	}
	defer res.Body.Close()
	return http.StatusOK, nil
}
