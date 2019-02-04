package email

import (
	"bytes"
	"html/template"
	"net/smtp"
	"strconv"
)

var (
	defaultTemplate, _ = template.New("email").Parse(emailTemplate)
)

// Email template
const emailTemplate = `
    From: {{.From}}
	To: {{.To}}
	Subject {{.Subject}}

	{{.Body}}`

// MailBean - all settings for email package in Bean-like struct
type MailBean struct {
	Message
	Credentials
}

// SendEmailByDefaultTemplate -  send email with default template @see email.emailTemplate const
func (m *MailBean) SendEmailByDefaultTemplate(body string) error {
	return SendEmailByDefaultTemplate(
		Message{m.From, m.To, m.Subject, body},
		Credentials{m.Username, m.Password, m.Server, m.Port})
}

// SendEmailByCustomTemplate -  send email with custom template
func (m *MailBean) SendEmailByCustomTemplate(emailTemplate, body string) error {
	return SendEmailByCustomTemplate(
		emailTemplate,
		Message{m.From, m.To, m.Subject, body},
		Credentials{m.Username, m.Password, m.Server, m.Port})
}

// Message - email message struct
type Message struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// Credentials - credentials
type Credentials struct {
	Username string
	Password string
	Server   string
	Port     int
}

// SendEmailByDefaultTemplate - send email with default template @see email.emailTemplate const
func SendEmailByDefaultTemplate(message Message, authCredential Credentials) error {
	var body bytes.Buffer
	_ = defaultTemplate.Execute(&body, message)
	auth := smtp.PlainAuth("", authCredential.Username, authCredential.Password, authCredential.Server)
	return smtp.SendMail(authCredential.Server+":"+strconv.Itoa(authCredential.Port),
		auth, message.From, message.To, body.Bytes())
}

// SendEmailByCustomTemplate - send email with custom template
func SendEmailByCustomTemplate(emailTemplate string, message Message, authCredential Credentials) error {
	customTemplate := template.New("email_custom")
	customTemplate, _ = customTemplate.Parse(emailTemplate)
	var body bytes.Buffer
	_ = customTemplate.Execute(&body, message)
	auth := smtp.PlainAuth("", authCredential.Username, authCredential.Password, authCredential.Server)
	return smtp.SendMail(authCredential.Server+":"+strconv.Itoa(authCredential.Port),
		auth, message.From, message.To, body.Bytes())
}
