package email

import (
	"io/ioutil"
	"log"
	"net/textproto"
	"testing"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
)

func TestReceiver(t *testing.T) {
	retriever, err := NewRetriever(RetrievingServers[GMail],
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}

	messages, err := retriever.RetrieveMails(SearchCriteria{
		From: "noreply@zohoaccounts.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("unexpected result len: real %v, expected %v", len(messages), 1)
	}
	msg0 := messages[0]
	if msg0.Subject != "Welcome to Zoho!" {
		t.Errorf("unexpected Subject: %v", msg0.Subject)
	}
	if msg0.MainPartMIMEType != TextHTML {
		t.Errorf("unexpected MainPartMIMEType: %v", msg0.MainPartMIMEType)
	}
	t.Logf("msg0: %v, %v", msg0.MIMEType, msg0.Body)
}

func _TestReceiverDebug(t *testing.T) {
	retriever, err := NewRetriever(RetrievingServers[GMail],
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}

	fromDate, _ := time.Parse(time.RFC3339, "2020-01-01T12:00:00Z")
	toDate, _ := time.Parse(time.RFC3339, "2021-06-24T12:00:00Z")
	_ = textproto.MIMEHeader{}
	seqNums, _ := retriever.mailer.Search(&imap.SearchCriteria{
		SentSince:  fromDate,
		SentBefore: toDate,
		Header: textproto.MIMEHeader{
			//"From": []string{"daominahpublic@gmail.com"},
			"From": []string{"no-reply@youtube.com"},
			//"Subject": []string{"subject0"},
		},
		//Text: []string{"aww"},
	})
	t.Log("seqNums: ", seqNums)
	if len(seqNums) == 0 {
		t.Fatal("empty search result")
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	// https://github.com/emersion/go-imap/wiki/Fetching-messages#fetching-the-whole-message-body

	bodySection := &imap.BodySectionName{}
	fetchItems := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody, bodySection.FetchItem()}
	retChan := make(chan *imap.Message, len(seqNums))
	err = retriever.mailer.Fetch(seqSet, fetchItems, retChan)
	if err != nil {
		t.Fatal(err)
	}
	messages := make([]*imap.Message, 0)
	for msg := range retChan {
		messages = append(messages, msg)
	}
	for _, msg := range messages {
		t.Logf("_______________________________________________________")
		t.Logf("%v", msg.Envelope.Date)
		t.Logf("%v", msg.Envelope.From[0].Address())
		t.Logf("%v", msg.Envelope.Subject)
		t.Logf("%v/%v", msg.BodyStructure.MIMEType, msg.BodyStructure.MIMESubType)

		bodyReader := msg.GetBody(bodySection)
		if bodyReader == nil {
			t.Fatal("bodySection not found")
		}
		mailReader, err := mail.CreateReader(bodyReader)
		if err != nil {
			log.Fatal(err)
		}
		for {
			part, err := mailReader.NextPart()
			if err != nil {
				break
			}
			t.Logf("part ContentType: %v", part.Header.Get("Content-Type"))
			switch header := part.Header.(type) {
			case *mail.InlineHeader: // text/plain or text/html
				content, _ := ioutil.ReadAll(part.Body)
				t.Logf("content: %s", content)
			case *mail.AttachmentHeader:
				filename, _ := header.Filename()
				t.Logf("attachment: %v", filename)
			}
		}
	}
}
