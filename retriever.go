package email

import (
	"crypto/tls"
	"errors"
	"fmt"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Retriever struct {
	providerAddrIMAP string
	username         string
	password         string
	mailer           *client.Client
}

// NewSender connects to IMAP server and tries to retrieve the last inbox,
// :arg providerAddrIMAP: example: "imap.gmail.com:993", see `popular_providers.go` for more examples,
// :arg username: string, example: "daominahpublic@gmail.com"
func NewRetriever(providerAddrIMAP string, username string, password string) (
	*Retriever, error) {
	var tlsConfig *tls.Config = nil
	//var tlsConfig = &tls.Config{InsecureSkipVerify: true}
	client0, err := client.DialTLS(providerAddrIMAP, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("client DialTLS: %v", err)
	}
	//defer client0.Logout()
	if err := client0.Login(username, password); err != nil {
		return nil, err
	}
	ret := &Retriever{
		providerAddrIMAP: providerAddrIMAP, username: username, password: password,
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

func (m Retriever) RetrieveMail() error {
	return errors.New("not implemented")
}
