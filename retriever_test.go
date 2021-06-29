package email

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/textproto"
	"strings"
	"testing"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
)

func TestReceiver(t *testing.T) {
	beginT := time.Now()
	provider0, username0, password0 := GMail, "daominahpublic@gmail.com", "HayQuen0*"
	//provider0, username0, password0 := ZohoMail, "a84869433334@zohomail.com", "HayQuen0*"
	retriever, err := NewRetriever(RetrievingServers[provider0],
		username0, password0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("initing retriever duration: %v", time.Since(beginT))

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
	//t.Logf("msg0: %v, %v, %v", msg0.Date, msg0.MIMEType, msg0.Body)

	t1, _ := time.Parse(time.RFC3339, "2021-05-11T13:16:00+07:00")
	messages1, err1 := retriever.RetrieveMails(SearchCriteria{
		From:      "noreply@zohoaccounts.com",
		SentSince: t1,
	})
	if err1 != nil {
		t.Fatal(err1)
	}
	if len(messages1) != 0 {
		t.Fatalf("unexpected result1 len: real %v, expected %v", len(messages), 0)
	}

	t2, _ := time.Parse(time.RFC3339, "2021-05-11T13:15:00+07:00")
	messages2, err2 := retriever.RetrieveMails(SearchCriteria{
		From:      "noreply@zohoaccounts.com",
		SentSince: t2,
	})
	if err2 != nil {
		t.Fatal(err2)
	}
	if len(messages2) != 1 {
		t.Fatalf("unexpected result2 len: real %v, expected %v", len(messages), 1)
	}

	t3, _ := time.Parse(time.RFC3339, "2021-06-07T00:00:00+07:00")
	messages3, err3 := retriever.RetrieveMails(SearchCriteria{
		SentSince:  t3,
		SentBefore: t3.Add(24 * time.Hour),
	})
	//t.Logf("messages3: %#v", messages3)
	if err3 != nil {
		t.Fatal(err3)
	}
	if len(messages3) != 1 {
		t.Fatalf("unexpected result3 len: real %v, expected %v", len(messages), 1)
	}
	if messages3[0].MailBox != Spam {
		t.Errorf("unexpected result3 mail box: real %v, expected %v", messages3[0].MailBox, Spam)
	}
}

func TestSendRetriever(t *testing.T) {
	provider0, username0, password0 := GMail, "daominahpublic@gmail.com", "HayQuen0*"
	sender, err0 := NewSender(SendingServers[provider0], username0, password0)
	retriever, err1 := NewRetriever(RetrievingServers[provider0], username0, password0)
	if err0 != nil || err1 != nil {
		t.Fatal(err0, err1)
	}
	_, _ = sender, retriever
	beginT := time.Now()
	rand.Seed(time.Now().UnixNano())
	content0 := fmt.Sprintf("%06d", rand.Intn(1000000))
	sentChan := make(chan bool)
	go func() {
		sender.SendMail(username0, "TestSendRetriever", TextPlain, content0)
		t.Logf("sent duration: %v, content0: %v", time.Since(beginT), content0)
		sentChan <- true
	}()
	ctx, ccl := context.WithTimeout(context.Background(), 125*time.Second)
	newMsg, err := retriever.RetrieveNewMail(ctx, SearchCriteria{
		SentSince: beginT.Add(-1 * time.Minute),
		From:      username0, Subject: "TestSendRetriever",
	})
	ccl()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("retrieved duration: %v", time.Since(beginT))
	<-sentChan
	if strings.TrimSpace(newMsg.Body) != content0 {
		t.Errorf("error RetrieveNewMail: real: %v, expected: %v", newMsg.Body, content0)
	}
}

func _TestReceiverDebug(t *testing.T) {
	retriever, err := NewRetriever(RetrievingServers[GMail],
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}
	inbox := retriever.boxClients[Inbox]
	if inbox == nil {
		t.Fatal("inbox client is nil")
	}

	fromDate, _ := time.Parse(time.RFC3339, "2020-01-01T12:00:00Z")
	toDate, _ := time.Parse(time.RFC3339, "2021-06-24T12:00:00Z")
	_ = textproto.MIMEHeader{}
	seqNums, _ := inbox.Search(&imap.SearchCriteria{
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
	err = inbox.Fetch(seqSet, fetchItems, retChan)
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
