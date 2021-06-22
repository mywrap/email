package email

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	provider Provider
	username string
	password string
	mailer   *gomail.Dialer
}

// NewSender connects and sends a test email to SMTP server,
// :arg provider: see `popular_providers.go`,
// :arg username: string, example: "daominahpublic@gmail.com"
func NewSender(provider Provider, username string, password string) (
	*Sender, error) {
	server, found := SendingServers[provider]
	if !found {
		return nil, errors.New("provider not found")
	}
	words := strings.Split(server, ":")
	if len(words) < 2 {
		return nil, errors.New("unexpected bad server address")
	}
	host, port := words[0], words[1]
	portInt, _ := strconv.Atoi(port)
	mailer := gomail.NewDialer(host, portInt, username, password)
	mailer.TLSConfig = nil
	//mailer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	ret := &Sender{
		provider: provider, username: username, password: password,
		mailer: mailer,
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err := ret.SendMail(username, "Test send mail at "+now, now)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (m Sender) SendMail(targetEmail string, subject string, content string) error {
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
