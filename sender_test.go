package email

import (
	"testing"
)

func TestGmailer_Send(t *testing.T) {
	_, err := NewMailer("daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Error(err)
	}
}
