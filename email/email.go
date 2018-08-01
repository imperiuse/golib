package email

import (
	"bytes"
	"html/template"
	"net/smtp"
	"strconv"
)

// Структура данных Электронного письма
type EmailMessage struct {
	From    string
	To      []string
	Subject string
	Body    string
}

type EmailCredentials struct {
	Username string
	Password string
	Server   string
	Port     int
}

// Шаблон письма ввиде строки
const emailTemplate = `
    From: {{.From}}
	To: {{.To}}
	Subject {{.Subject}}

	{{.Body}}`

var defaultTemplate *template.Template

func init() {
	defaultTemplate = template.New("email")
	defaultTemplate.Parse(emailTemplate)
}

func SendEmailMsgDefaultTemplate(message EmailMessage, authCredential EmailCredentials) error {
	var body bytes.Buffer
	defaultTemplate.Execute(&body, message)
	auth := smtp.PlainAuth("", authCredential.Username, authCredential.Password, authCredential.Server)
	return smtp.SendMail(authCredential.Server+":"+strconv.Itoa(authCredential.Port), auth, message.From, message.To, body.Bytes())
}

func SendEmailMsgCustomTemplate(emailTemplate string, message EmailMessage, authCredential EmailCredentials) error {
	customTemplate := template.New("email_custom")
	customTemplate.Parse(emailTemplate)
	var body bytes.Buffer
	customTemplate.Execute(&body, message)
	auth := smtp.PlainAuth("", authCredential.Username, authCredential.Password, authCredential.Server)
	return smtp.SendMail(authCredential.Server+":"+strconv.Itoa(authCredential.Port), auth, message.From, message.To, body.Bytes())
}

func ExampleSendEmail() error {

	// Заполнение набора данных для шаблона и клиента электронной почты
	message := &EmailMessage{
		"me@example.com",
		[]string{"you@example.com"},
		"A test",
		"Just saying hi",
	}

	// Создание электронного письма по шаблону и запись его в буфер
	var body bytes.Buffer
	defaultTemplate.Execute(&body, message)

	// Настройка SMTP-клиента
	authCredential := &EmailCredentials{
		"MyUserName",
		"myPass",
		"smtp.example.com",
		25,
	}
	auth := smtp.PlainAuth("", authCredential.Username, authCredential.Password, authCredential.Server)

	// Отправка элетронного письма
	return smtp.SendMail(authCredential.Server+":"+strconv.Itoa(authCredential.Port), auth, message.From, message.To, body.Bytes())
}
