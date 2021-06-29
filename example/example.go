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
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	rand.Seed(time.Now().UnixNano())

	sender0 := "daominahpublic@gmail.com"
	retriever0 := "a84869433334@zohomail.com"
	sender, err := email.NewSender("smtp.gmail.com:587",
		sender0, "HayQuen0*")
	if err != nil {
		log.Fatal(err)
	}
	beginT := time.Now()
	log.Println("beginT")
	go func() {
		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		content0 := fmt.Sprintf(`<p>OTP is <h1>%v</h1></p>`, otp)
		err = sender.SendMail("a84869433334@zohomail.com",
			"Test send OTP "+otp, email.TextHTML, content0)
		if err != nil {
			log.Println(err)
		}
		log.Printf("sent duration: %v\n", time.Since(beginT))
	}()

	retriever, err := email.NewRetriever("imap.zoho.com:993",
		retriever0, "HayQuen0*")
	if err != nil {
		log.Fatal(err)
	}
	ctx, ccl := context.WithTimeout(context.Background(), 125*time.Second)
	// periodically check new email from the sender0 (inbox and spam),
	msg, err := retriever.RetrieveNewMail(ctx, email.SearchCriteria{
		SentSince: beginT.Add(-1 * time.Minute), From: sender0,
	})
	ccl()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("new email: from %v date %v: delay1 %v, delay2 %v):\n%v",
		msg.From, msg.Date.Format(time.RFC3339),
		msg.Date.Sub(beginT), time.Since(beginT),
		msg.Body)

	// see *_test.go for more usages
}
