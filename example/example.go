package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mywrap/email"
)

func main() {
	sender0 := "daominahpublic@gmail.com"
	retriever0 := "a84869433334@zohomail.com"
	beginT := time.Now()
	sender, err := email.NewSender("smtp.gmail.com:587",
		sender0, "HayQuen0*")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		content0 := fmt.Sprintf(`<p>OTP is <h1>%v</h1></p>`, otp)
		err = sender.SendMail("a84869433334@zohomail.com",
			"Test send OTP "+otp, email.TextHTML, content0)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("sent")
	}()

	retriever, err := email.NewRetriever("imap.zoho.com:993",
		retriever0, "HayQuen0*")
	if err != nil {
		log.Fatal(err)
	}
	ctx, ccl := context.WithTimeout(context.Background(), 30*time.Second)
	// periodically check new email from the sender0 (inbox and spam)
	msg, err := retriever.RetrieveNewMail(ctx, sender0, beginT)
	ccl()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("got a new email from %v: %v", msg.From, msg.Body)

	// see *_test.go for more usages
}
