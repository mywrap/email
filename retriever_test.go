package email

import (
	"testing"
)

func TestReceiver(t *testing.T) {
	retriever, err := NewRetriever(GMail,
		"daominahpublic@gmail.com", "HayQuen0*")
	if err != nil {
		t.Fatal(err)
	}
	_ = retriever
}
