package email

// Provider is a const string determines email provider,
// this file has map provider name to their SMTP and IMAP server address
type Provider string

// Provider enum
const (
	// AOL account need to create password for third-party app at URL
	// https://login.aol.com/account/security/app-passwords/list
	AOLMail Provider = "AOLMail"

	// Google account have to change account setting at URL
	// https://myaccount.google.com/u/2/lesssecureapps
	GMail Provider = "GMail"

	// Zoho account need to enable IMAP at URL
	// https://mail.zoho.com/zm/#settings/all/mailaccounts
	ZohoMail Provider = "ZohoMail"
)

// SendingServers maps provider to SMTP host:port
var SendingServers = map[Provider]string{
	AOLMail:  "smtp.aol.com:465",
	GMail:    "smtp.gmail.com:587",
	ZohoMail: "smtp.zoho.com:465",
}

// RetrievingServers maps provider to IMAP host:port
var RetrievingServers = map[Provider]string{
	AOLMail:  "imap.aol.com:993",
	GMail:    "imap.gmail.com:993",
	ZohoMail: "imap.zoho.com:993",
}
