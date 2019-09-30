package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"

	"github.com/sirupsen/logrus"

	"github.com/imperiuse/golib/concat"
)

// MailBean - all settings for email package in Bean-like struct
type (
	MailBean struct {
		from mail.Address   // mail.Address{"FromMe", "from@yandex.ru"}
		to   []mail.Address // mail.Address{"ToYou", "to@yandex.ru"}
		Subj string
		Credentials
		EnableNotify bool
		log          logrus.FieldLogger
	}

	Credentials struct {
		Username   string // "from@yandex.ru"
		Password   string // "token app"
		ServerName string // "smtp.yandex.ru:465"
		Identity   string // ""
	}
)

func NewMailBean(log logrus.FieldLogger, from mail.Address, to []mail.Address, subj string, creds Credentials, enableNotify bool) *MailBean {
	if log == nil {
		log = logrus.New()
	}
	return &MailBean{
		from:         from,
		to:           to,
		Subj:         subj,
		Credentials:  creds,
		EnableNotify: enableNotify,
		log:          log,
	}
}

// SetFromAndToEmailAddresses - set from and to email addresses
func (m *MailBean) SetFromAndToEmailAddresses(from mail.Address, to []mail.Address) {
	m.from = from
	m.to = to
}

// SendEmails -  send email with default template @see email.emailTemplate const
func (m *MailBean) SendEmails(body string) error {
	if m.EnableNotify {
		for _, to := range m.to {
			if err := m.sendEmail(to, m.Subj, body); err != nil {
				return err
			}
		}
	}
	return nil
}

// sendEmail - send one email
func (m *MailBean) sendEmail(to mail.Address, subj, body string) (err error) {
	// Setup headers
	headers := make(map[string]string)
	headers["from"] = m.from.String()
	headers["to"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message = concat.Strings(message, fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message = concat.Strings(message, "\r\n", body)

	// Connect to the SMTP Server
	host, _, _ := net.SplitHostPort(m.Credentials.ServerName)

	auth := smtp.PlainAuth(m.Credentials.Identity, m.Credentials.Username, m.Credentials.Password, host)

	// TLS config
	// nolint i undestang so it bad for sec set True, but i need do this!
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", m.Credentials.ServerName, tlsconfig)
	if err != nil {
		m.log.WithError(err).Error("tls.Dial('tcp', m.Credentials.ServerName, tlsconfig)")
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		m.log.WithError(err).Error("smtp.NewClient(conn, host)")
	}

	// Auth
	if err = client.Auth(auth); err != nil {
		m.log.WithError(err).Error("client.Auth(auth)")
	}

	// to && from
	if err = client.Mail(m.from.Address); err != nil {
		m.log.WithError(err).Error("client.Mail(m.from.Address)")
	}

	if err = client.Rcpt(to.Address); err != nil {
		m.log.WithError(err).Error("client.Rcpt(to.Address)")
	}

	// Data
	w, err := client.Data()
	if err != nil {
		m.log.WithError(err).Error("client.Data")
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		m.log.WithError(err).Error("w.Write([]byte(message))")
	}

	err = w.Close()
	if err != nil {
		m.log.WithError(err).Error("w.Close()")
	}

	err = client.Quit()

	return err
}
