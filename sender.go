package email

import (
	"crypto/tls"
	"fmt"
	"time"

	"gopkg.in/gomail.v2"
)

// Mailer uses a google account to send email. Have to change account setting
// at https://myaccount.google.com/u/2/lesssecureapps to make this Mailer work
type Mailer struct {
	username string
	password string
	mailer   *gomail.Dialer
}

// NewMailer init a Mailer,
// :arg username: string, example: "daominahpublic@gmail.com"
func NewMailer(username string, password string) (*Mailer, error) {
	mailer := gomail.NewDialer("smtp.gmail.com", 587, username, password)
	mailer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	result := &Mailer{
		username: username,
		password: password,
		mailer:   mailer,
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err := result.SendMail(username, "Test SendMail "+now, now)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m Mailer) SendMail(targetEmail string, subject string, content string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.username)
	msg.SetHeader("To", targetEmail)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", content)
	err := m.mailer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send %v to %v: %v", m.username, targetEmail, err)
	}
	return nil
}
