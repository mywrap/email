package email

import (
	"crypto/tls"
	"errors"
	"fmt"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Retriever struct {
	provider Provider
	username string
	password string
	mailer   *client.Client
}

// NewSender connects to IMAP server and tries to retrieve the last inbox,
// :arg provider: see `popular_providers.go`,
// :arg username: string, example: "daominahpublic@gmail.com"
func NewRetriever(provider Provider, username string, password string) (
	*Retriever, error) {
	server, found := RetrievingServers[provider]
	if !found {
		return nil, errors.New("provider not found")
	}

	var tlsConfig *tls.Config = nil
	//var tlsConfig = &tls.Config{InsecureSkipVerify: true}
	client0, err := client.DialTLS(server, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("client DialTLS: %v", err)
	}
	//defer client0.Logout()
	if err := client0.Login(username, password); err != nil {
		return nil, err
	}
	ret := &Retriever{
		provider: provider, username: username, password: password,
		mailer: client0,
	}

	box0, err := client0.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("client select INBOX: %v", err)
	}
	if box0.Messages <= 0 { // number of messages in the mail box
		return ret, nil
	}
	from := box0.Messages - 1
	msgIds := new(imap.SeqSet)
	msgIds.AddRange(from, box0.Messages)
	messages := make(chan *imap.Message, 10)
	err = client0.Fetch(msgIds, []imap.FetchItem{imap.FetchEnvelope}, messages)
	if err != nil {
		return nil, fmt.Errorf("client Fetch messages: %v", err)
	}
	for msg := range messages {
		println("last inbox: ", msg.Envelope.Subject)
	}
	return ret, nil
}

func (m Sender) RetrieveMail() error {
	return errors.New("not implemented")
}
