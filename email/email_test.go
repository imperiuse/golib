package email

import (
	"bytes"
	"net/smtp"
	"strconv"
)

func ExampleSendEmailByDefaultTemplate() {

	// Заполнение набора данных для шаблона и клиента электронной почты
	message := &Message{
		"me@example.com",
		[]string{"you@example.com"},
		"A test",
		"Just saying hi",
	}

	// Создание электронного письма по шаблону и запись его в буфер
	var body bytes.Buffer
	_ = defaultTemplate.Execute(&body, message)

	// Настройка SMTP-клиента
	authCredential := &Credentials{
		"MyUserName",
		"myPass",
		"smtp.example.com",
		25,
	}
	auth := smtp.PlainAuth("", authCredential.Username, authCredential.Password, authCredential.Server)

	// Отправка элетронного письма
	_ = smtp.SendMail(authCredential.Server+":"+strconv.Itoa(authCredential.Port),
		auth, message.From, message.To, body.Bytes())
}
