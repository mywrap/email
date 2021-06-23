package email

// Provider is a const string determines email provider,
// this file has map provider name to their SMTP and IMAP server address
type Provider string

// Provider enum
const (
	// Google account have to change account setting at
	// https://myaccount.google.com/u/2/lesssecureapps
	GMail Provider = "GMail"

	ZohoMail Provider = "ZohoMail"
)

// SendingServers maps provider to SMTP host:port
var SendingServers = map[Provider]string{
	GMail:    "smtp.gmail.com:587",
	ZohoMail: "smtp.zoho.com:465",
}

// RetrievingServers maps provider to IMAP host:port
var RetrievingServers = map[Provider]string{
	GMail:    "imap.gmail.com:993",
	ZohoMail: "imap.zoho.com:993",
}
