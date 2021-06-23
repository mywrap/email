package email

import (
	"net/textproto"
	"testing"
	"time"

	"github.com/emersion/go-imap"
)

func TestReceiver(t *testing.T) {
	retriever, err := NewRetriever(RetrievingServers[GMail],
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}

	fromDate, _ := time.Parse(time.RFC3339, "2020-01-01T12:00:00Z")
	toDate, _ := time.Parse(time.RFC3339, "2021-06-23T12:00:00Z")
	_ = textproto.MIMEHeader{}
	seqNums, err := retriever.mailer.Search(&imap.SearchCriteria{
		SentSince:  fromDate,
		SentBefore: toDate,
		//Header: textproto.MIMEHeader{
		//	"From": []string{"daominahpublic@gmail.com"},
		//},
		Text: []string{"rồi"},
	})
	t.Log("seqNums: ", seqNums)
	if len(seqNums) == 0 {
		t.Fatal("empty search result")
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)
	retChan := make(chan *imap.Message, len(seqNums))
	err = retriever.mailer.Fetch(seqSet, imap.FetchFull.Expand(), retChan)
	if err != nil {
		t.Fatal(err)
	}
	messages := make([]*imap.Message, 0)
	for msg := range retChan {
		messages = append(messages, msg)
	}
	for _, msg := range messages {
		t.Logf("_______________________________________________________")
		t.Logf("%v", msg.Envelope.From[0].Address())
		t.Logf("%v", msg.Envelope.Date)
		t.Logf("%v", msg.Envelope.Subject)
		t.Logf("%v/%v", msg.BodyStructure.MIMEType, msg.BodyStructure.MIMESubType)
	}
}
