package email

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"
)

// Sender wrapped a SMTP client
type Sender struct {
	providerAddrSMTP string
	username         string
	password         string
	mailer           *gomail.Dialer
}

// NewSender connects and sends a test email to SMTP server,
// :arg providerAddrSMTP: example: "smtp.gmail.com:587", see `popular_providers.go` for more examples,
// :arg username: example: "daominahpublic@gmail.com"
func NewSender(providerAddrSMTP string, username string, password string) (
	*Sender, error) {
	words := strings.Split(providerAddrSMTP, ":")
	if len(words) < 2 {
		return nil, errors.New("unexpected bad server address")
	}
	host, port := words[0], words[1]
	portInt, _ := strconv.Atoi(port)
	mailer := gomail.NewDialer(host, portInt, username, password)
	mailer.TLSConfig = nil
	//mailer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	ret := &Sender{
		providerAddrSMTP: providerAddrSMTP, username: username, password: password,
		mailer: mailer,
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err := ret.SendMail(username, "initing Sender test "+now, TextPlain, now)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// SendMail sends an email,
// this func opens a connection, sends the given emails and closes the connection,
// TODO: consider to reuse a persistent connection,
// :arg contentType: can be TextPlain or TextHTML
func (m Sender) SendMail(targetEmail string,
	subject string, contentType MIMEType, content string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.username)
	msg.SetHeader("To", targetEmail)
	msg.SetHeader("Subject", subject)
	msg.SetBody(string(contentType), content)
	err := m.mailer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send %v to %v: %v", m.username, targetEmail, err)
	}
	return nil
}

// MIMEType stands for Multipurpose Internet Mail Extensions,
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types,
// this package only supports "text/plain" and "text/html"
type MIMEType string

// MIMEType enum
const (
	TextPlain MIMEType = "text/plain"
	TextHTML  MIMEType = "text/html"
)
