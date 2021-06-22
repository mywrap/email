# Send and receive email with Go

Sending email via SMTP and receiving via IMAP.  
Wrapped [gopkg.in/gomail.v2](https://github.com/go-gomail/gomail/tree/v2) and 
[emersion/go-imap](https://github.com/emersion/go-imap).

## Glossary

* SMTP: Simple Mail Transfer Protocol allows you to send emails from an 
  email application through a specific server. Default ports: 465, 587, 2525.

* IMAP: Internet Message Access Protocol is an email retrieval and 
  storage protocol, which syncs with the servers and maintains the 
  status of messages across multiple email clients.

* POP: Post Office Protocol enables you to receive the emails but POP 
  performs one-way email retrieval and there is no sync between the email
  clients and server. POP can be used only from a single device. With 
  default option, emails are downloaded and deleted from the server.
  This package does not support POP.
