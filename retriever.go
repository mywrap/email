package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"
	"strings"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// Sender wrapped an IMAP client
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
		return nil, fmt.Errorf("client test Fetch: %v", err)
	}
	for msg := range messages {
		_ = msg.Envelope.Subject
		//println("last inbox: ", msg.Envelope.Subject)
	}
	return ret, nil
}

// SearchCriteria simplifies IMAP's search criteria format
type SearchCriteria struct {
	SentSince  time.Time // header Date is later than the filter, regarding time
	SentBefore time.Time // header Date is earlier than the filter, disregarding time
	From       string    // sender address matches exactly the specified string
	Subject    string    // header Subject contains the specified string
	Text       string    // header or body (space split) contains the specified string
}

// Message simplifies IMAP's email format
type Message struct {
	Date             time.Time // Envelope.Date
	From             string    // Envelope.From[0].Address
	Subject          string    // Envelope.Subject
	MIMEType         MIMEType  // BodyStructure.MIMEType/BodyStructure.MIMESubType
	Body             string    // only support TextPlain or TextHTML
	MainPartMIMEType MIMEType  // only support TextPlain or TextHTML
}

func (r Retriever) RetrieveMails(filter SearchCriteria) ([]Message, error) {
	search := &imap.SearchCriteria{}
	if !filter.SentSince.IsZero() {
		search.SentSince = filter.SentSince
	}
	if !filter.SentBefore.IsZero() {
		search.SentBefore = filter.SentBefore
	}
	searchHeader := textproto.MIMEHeader{}
	if filter.From != "" {
		searchHeader.Set("From", filter.From)
	}
	if filter.Subject != "" {
		searchHeader.Set("Subject", filter.Subject)
	}
	if len(searchHeader) != 0 {
		search.Header = searchHeader
	}
	if filter.Text != "" {
		search.Text = []string{filter.Text}
	}

	seqNums, err := r.mailer.Search(search)
	if err != nil {
		return nil, fmt.Errorf("imap search request failed: %v", err)
	}
	if len(seqNums) == 0 {
		return nil, errors.New("empty search result")
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	bodySection := &imap.BodySectionName{} // const
	fetchItems := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody, bodySection.FetchItem()}
	retChan := make(chan *imap.Message, len(seqNums))
	err = r.mailer.Fetch(seqSet, fetchItems, retChan)
	if err != nil {
		return nil, fmt.Errorf("imap fetch request failed: %v", err)
	}
	imapMessages := make([]*imap.Message, 0)
	for msg := range retChan {
		imapMessages = append(imapMessages, msg)
	}
	ret := make([]Message, 0)
	for _, imapMsg := range imapMessages {
		var msg Message

		if imapMsg.Envelope != nil {
			msg.Date = imapMsg.Envelope.Date
			if len(imapMsg.Envelope.From) > 0 {
				msg.From = imapMsg.Envelope.From[0].Address()
			}
			msg.Subject = imapMsg.Envelope.Subject
			msg.MIMEType = MIMEType(fmt.Sprintf("%v/%v",
				imapMsg.BodyStructure.MIMEType, imapMsg.BodyStructure.MIMESubType))
		}

		bodyReader := imapMsg.GetBody(bodySection)
		if bodyReader == nil {
			return nil, fmt.Errorf("imap body section not found")
		}
		mailReader, err := mail.CreateReader(bodyReader)
		if err != nil {
			return nil, fmt.Errorf("mail CreateReader: %v", err)
		}
		_ = mailReader
		for { // loop through all parts but only care about text part
			part, err := mailReader.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, fmt.Errorf("mailReader NextPart: %v", err)
			}
			switch header := part.Header.(type) {
			case *mail.InlineHeader: // text/plain or text/html
				if strings.Contains(part.Header.Get("Content-Type"), "text/plain") {
					if msg.MainPartMIMEType == TextHTML && msg.Body != "" {
						// skip fetch plain text that is similar to fetched html
						continue
					} else {
						msg.MainPartMIMEType = TextPlain
					}
				} else {
					msg.MainPartMIMEType = TextHTML
				}
				content, err := ioutil.ReadAll(part.Body)
				if err != nil {
					return nil, fmt.Errorf("ioutil ReadAll part: %v", err)
				}
				msg.Body = string(content)
			case *mail.AttachmentHeader:
				_, _ = header.Filename()
				continue // ignore images, files, ..
			}
		}

		ret = append(ret, msg)
	}
	return ret, nil
}
