package email

//import (
//	"crypto/tls"
//	"errors"
//	"fmt"
//	"strconv"
//	"strings"
//	"time"
//
//	imap "github.com/emersion/go-imap"
//	"github.com/emersion/go-imap/client"
//)
//
//type Retriever struct {
//	provider Provider
//	username string
//	password string
//	mailer   *client.Client
//}
//
//// NewSender connects to IMAP server and tries to retrieve the last inbox,
//// :arg provider: see `popular_providers.go`,
//// :arg username: string, example: "daominahpublic@gmail.com"
//func NewRetriever(provider Provider, username string, password string) (
//	*Retriever, error) {
//	server, found := RetrievingServers[provider]
//	if !found {
//		return nil, errors.New("provider not found")
//	}
//
//	var tlsConfig *tls.Config = nil
//	//var tlsConfig = &tls.Config{InsecureSkipVerify: true}
//	imapClient, err := client.DialTLS(server, tlsConfig)
//	if err != nil {
//		return nil, fmt.Errorf("fail to connect ")
//	}
//	//defer imapClient.Logout()
//	if err := imapClient.Login(username, password); err != nil {
//		return nil, err
//	}
//}
//
//func (m Sender) SendMail(targetEmail string, subject string, content string) error {
//	msg := gomail.NewMessage()
//	msg.SetHeader("From", m.username)
//	msg.SetHeader("To", targetEmail)
//	msg.SetHeader("Subject", subject)
//	msg.SetBody("text/plain", content)
//	err := m.mailer.DialAndSend(msg)
//	if err != nil {
//		return fmt.Errorf("send %v to %v: %v", m.username, targetEmail, err)
//	}
//	return nil
//}
