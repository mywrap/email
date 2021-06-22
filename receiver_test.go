package email

import (
	"testing"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func TestReceiver(t *testing.T) {
	t.Log("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("daominahpublic@gmail.com", "HayQuen0*"); err != nil {
		t.Fatal(err)
	}
	t.Log("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	t.Log("Mailboxes:")
	for m := range mailboxes {
		t.Log("* " + m.Name)
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	t.Log("Last 4 messages:")
	for msg := range messages {
		t.Log(time.Now().Format(time.RFC3339Nano) + " * " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}

	t.Log("Done!")
}
