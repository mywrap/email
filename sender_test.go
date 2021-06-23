package email

import (
	"testing"
)

func TestSender_GMail(t *testing.T) {
	sender, err := NewSender(SendingServers[GMail],
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}
	content0 := `
		<h1>hello</h1>
		<a href="http://127.0.0.1">localhost</a>
		<pre>{"aww": "bii"}</pre>
	`
	err = sender.SendMail("daominahpublic@gmail.com",
		"Test send HTML", TextHTML, content0)
	if err != nil {
		t.Error(err)
	}
}

func TestSender_ZohoMail(t *testing.T) {
	_, err := NewSender(SendingServers[ZohoMail],
		"84869433334a@zohomail.com", "HayQuen0*")
	if err != nil {
		t.Error(err)
	}
}
