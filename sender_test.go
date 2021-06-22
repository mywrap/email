package email

import (
	"testing"
)

func TestSender_GMail(t *testing.T) {
	sender, err := NewSender(GMail,
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}
	err = sender.SendMail("daominahpublic@gmail.com",
		"subject0", "content0")
	if err != nil {
		t.Error(err)
	}
}

func TestSender_ZohoMail(t *testing.T) {
	_, err := NewSender(ZohoMail,
		"84869433334a@zohomail.com", "HayQuen0*")
	if err != nil {
		t.Error(err)
	}
}
