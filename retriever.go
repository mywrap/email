package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"
	"regexp"
	"strings"
	"sync"
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

	// real mail box names on provider server,
	// must be read only after inited because of simple lock
	boxNames map[MailBox]string
	// this retriever uses 1 connection per box,
	// must be read only after inited because of simple lock
	boxClients map[MailBox]*client.Client
	mutex      *sync.Mutex // protect above 2 maps
}

// MailBox is a mail box regex to match provider mail box name,
type MailBox string

// MailBox enum
const (
	Inbox MailBox = "INBOX"
	Spam  MailBox = "SPAM|BULK"
)

func (p MailBox) CheckMatch(mailBoxName string) bool {
	regexObj, err := regexp.Compile("(?i)" + string(p))
	if err != nil {
		return false // should be unreachable
	}
	return regexObj.MatchString(mailBoxName)
}

// NewSender connects to IMAP server then selects mail boxes,
// :arg providerAddrIMAP: example: "imap.gmail.com:993", see `popular_providers.go` for more examples,
// :arg username: string, example: "daominahpublic@gmail.com"
func NewRetriever(providerAddrIMAP string, username string, password string) (
	*Retriever, error) {
	ret := &Retriever{
		providerAddrIMAP: providerAddrIMAP,
		username:         username,
		password:         password,
		boxNames:         make(map[MailBox]string),
		boxClients:       make(map[MailBox]*client.Client),
		mutex:            &sync.Mutex{},
	}
	boxesToFetch := []MailBox{Inbox, Spam}
	errsChan := make(chan error, len(boxesToFetch))
	for _, mailBoxPtn := range boxesToFetch {
		mailBoxPtn := mailBoxPtn
		go func() {
			var tlsConfig *tls.Config = nil
			//var tlsConfig = &tls.Config{InsecureSkipVerify: true}
			client0, err := client.DialTLS(providerAddrIMAP, tlsConfig)
			if err != nil {
				errsChan <- fmt.Errorf("client DialTLS: %v", err)
				return
			}
			if err := client0.Login(username, password); err != nil {
				errsChan <- fmt.Errorf("client Login: %v", err)
				return
			}
			mailBoxes := make(chan *imap.MailboxInfo, 100)
			err = client0.List("", "*", mailBoxes)
			if err != nil {
				errsChan <- fmt.Errorf("client List boxes: %v", err)
				return
			}
			mailBoxName := string(mailBoxPtn)
			for mailBox := range mailBoxes {
				//fmt.Printf("debug pattern: %v, box name: %v\n", mailBoxPtn, mailBox.Name)
				if mailBoxPtn.CheckMatch(mailBox.Name) {
					mailBoxName = mailBox.Name
					break
				}
			}
			mailBoxStatus, err := client0.Select(mailBoxName, true)
			if err != nil {
				errsChan <- fmt.Errorf("select mail box %v: %v", mailBoxName, err)
				return
			}
			_ = mailBoxStatus
			ret.mutex.Lock()
			ret.boxNames[mailBoxPtn] = mailBoxName
			ret.boxClients[mailBoxPtn] = client0
			ret.mutex.Unlock()
			errsChan <- nil
		}()
	}
	for i := 0; i < len(boxesToFetch); i++ {
		oneBoxErr := <-errsChan
		if oneBoxErr != nil {
			return nil, oneBoxErr
		}
	}
	//fmt.Printf("debug boxNames: %#v\n", ret.boxNames)
	return ret, nil
}

// CloseConnections tries to gracefully closes the connections
func (r Retriever) CloseConnections() {
	for _, cli := range r.boxClients {
		cli.Logout()
	}
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
	Date    time.Time // Envelope.Date
	From    string    // Envelope.From[0].Address
	Subject string    // Envelope.Subject
	Body    string    // only support TextPlain or TextHTML

	// following fields are not important, can be ignore

	MIMEType         MIMEType // BodyStructure.MIMEType/BodyStructure.MIMESubType
	MainPartMIMEType MIMEType // only support TextPlain or TextHTML
	MailBox          MailBox  // only support INBOX and SPAM
}

// retrieveMails simplifies IMAP's fetch
func (r Retriever) retrieveMails(filter SearchCriteria, boxName MailBox) (
	[]Message, error) {
	boxClient := r.boxClients[boxName]
	if boxClient == nil {
		return nil, fmt.Errorf("invalid mail box name %v", boxName)
	}
	// feels like we need to reselect the mail box to get new message
	_, err := boxClient.Select(r.boxNames[boxName], true)
	if err != nil {
		return nil, fmt.Errorf("select mail box: %v", err)
	}

	search := &imap.SearchCriteria{}
	if !filter.SentSince.IsZero() {
		// IMAP's search specs disregards time so this filter should be excess,
		// we will filter by time for each message later
		search.SentSince = filter.SentSince.Add(-24 * time.Hour)
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

	seqNums, err := boxClient.Search(search)
	if err != nil {
		return nil, fmt.Errorf("imap search request failed: %v", err)
	}
	if len(seqNums) == 0 {
		return nil, nil
	}
	if len(seqNums) > 1000 { // just for safe, input query should limit date range
		seqNums = seqNums[len(seqNums)-1000:]
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	bodySection := &imap.BodySectionName{} // const
	fetchItems := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody, bodySection.FetchItem()}
	imapMessages := make(chan *imap.Message, len(seqNums))
	err = boxClient.Fetch(seqSet, fetchItems, imapMessages)
	if err != nil {
		return nil, fmt.Errorf("imap fetch request failed: %v", err)
	}
	ret := make([]Message, 0)
	for imapMsg := range imapMessages {
		msg := Message{MailBox: boxName}

		if imapMsg.Envelope != nil {
			msg.Date = imapMsg.Envelope.Date
			if msg.Date.Before(filter.SentSince) {
				continue
			}
			if len(imapMsg.Envelope.From) > 0 {
				msg.From = imapMsg.Envelope.From[0].Address()
			}
			msg.Subject = imapMsg.Envelope.Subject
		}
		if imapMsg.BodyStructure != nil {
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

// retrieveMails simplifies IMAP's fetch (from inbox and spam)
func (r Retriever) RetrieveMails(filter SearchCriteria) ([]Message, error) {
	retChan := make(chan []Message, len(r.boxClients))
	errChan := make(chan error, len(r.boxClients))
	for boxName, _ := range r.boxClients {
		boxName := boxName
		go func() {
			msgs, err := r.retrieveMails(filter, boxName)
			retChan <- msgs
			errChan <- err
		}()
	}
	for i := 0; i < len(r.boxClients); i++ {
		oneBoxErr := <-errChan
		if oneBoxErr != nil {
			return nil, oneBoxErr
		}
	}
	ret := make([]Message, 0)
	for i := 0; i < len(r.boxClients); i++ {
		oneBoxMsgs := <-retChan
		ret = append(ret, oneBoxMsgs...)
	}
	return ret, nil
}

// RetrieveNewMail periodically check inbox and spam until getting a new message
// or the input context is cancelled
func (r Retriever) RetrieveNewMail(
	ctx context.Context, filter SearchCriteria) (Message, error) {
	var lastErr error
	for i := 0; true; i++ {
		if i > 0 {
			time.Sleep(10 * time.Second)
		}
		select {
		case <-ctx.Done():
			if lastErr == nil {
				lastErr = ctx.Err()
			}
			return Message{}, lastErr
		default:
			// continue to check inbox
		}
		msgs, err := r.RetrieveMails(filter)
		if err != nil {
			lastErr = err
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		//for _, m := range msgs {
		//	fmt.Printf("debug msg %v: %v\n", m.Date, m.Body)
		//}

		return msgs[len(msgs)-1], nil
	}
	return Message{}, errors.New("unreachable")
}
