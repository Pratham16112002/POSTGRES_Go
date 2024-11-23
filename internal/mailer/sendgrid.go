package mailer

import (
	"bytes"
	"encoding/json"
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

type MailPayloadData struct {
	Data MailPayload
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

	var PayloadData MailPayloadData

	PayloadData.Data = payload

	var data_payload_writer []byte
	if err := json.NewEncoder(bytes.NewBuffer(data_payload_writer)).Encode(&PayloadData); err != nil {
		return -1, err
	}

	ext_req, err := http.NewRequest(http.MethodPost, "https://api.postmarkapp.com/email", bytes.NewReader(data_payload_writer))
	if err != nil {
		return -1, err
	}
	ext_req.Header.Add("Accept", "application/json")
	ext_req.Header.Add("Content-Type", "application/json")
	ext_req.Header.Add("X-Postmark-Server-Token", m.apiKey)

	for i := 0; i < MaxRetries; i++ {
		client := &http.Client{}
		res, err := client.Do(ext_req)
		if err != nil {
			return -1, fmt.Errorf("smtp error")
		}

		fmt.Printf("The status %d \n, body is %v", res.StatusCode, res.Body)
	}
	return http.StatusInternalServerError, fmt.Errorf("failed to send email after %d attemps, error : ", MaxRetries)
}
