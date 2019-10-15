package mailgun

import (
	"regexp"
)

type Config struct {
	ApiKey string
	Domain string
	Eu     bool
}

type EventsResponse struct {
	Items  []Item `json:"items"`
	Paging Paging `json:"paging"`
}

type Item struct {
	Storage         Storage        `json:"storage"`
	Severity        string         `json:"severity"`
	DeliveryStatus  DeliveryStatus `json:"delivery-status"`
	RecipientDomain string         `json:"recipient-domain"`
	Reason          string         `json:"reason"`
	Flags           Flags          `json:"flags"`
	Timestamp       float64        `json:"timestamp"`
	Envelope        Envelope       `json:"envelope"`
	Message         Message        `json:"message"`
	Recipient       string         `json:"recipient"`
	Event           string         `json:"event"`
}

type Storage struct {
	Key string `json:"key"`
}

type DeliveryStatus struct {
	Tls                 bool   `json:"tls"`
	MxHost              string `json:"mx-host"`
	Code                int    `json:"code"`
	Description         string `json:"description"`
	AttemptNo           int    `json:"attempt-no"`
	Message             string `json:"message"`
	CertificateVerified bool   `json:"certificate-verified"`
}

type Flags struct {
	IsRouted        bool `json:"is-routed"`
	IsAuthenticated bool `json:"is-authenticated"`
	IsSystemTest    bool `json:"is-system-test"`
	IsTestMode      bool `json:"is-test-mode"`
}

type Envelope struct {
	Transport string `json:"transport"`
	Sender    string `json:"sender"`
	SendingIp string `json:"sending-ip"`
	Targets   string `json:"targets"`
}

type Message struct {
	Headers Headers `json:"headers"`
	Size    int     `json:"size"`
}

type Headers struct {
	To        string `json:"to"`
	MessageId string `json:"message-id"`
	From      string `json:"from"`
	Subject   string `json:"subject"`
}

type Paging struct {
	Previous string `json:"previous"`
	First    string `json:"first"`
	Last     string `json:"last"`
	Next     string `json:"next"`
}

type Email struct {
	From       string `json:"From"`
	To         string `json:"To"`
	Subject    string `json:"Subject"`
	Text       string `json:"body-plain"`
	References string `json:"References"`
	MessageId  string `json:"Message-Id"`
	InReplyTo  string `json:"In-Reply-To"`

	Cc  string
	Bcc string
}

var emailRegexp = regexp.MustCompile(`<[^<>]*>`)

func (i *Item) From() string {
	if len(i.Envelope.Sender) > 0 {
		return i.Envelope.Sender
	}
	return findEmail(i.Message.Headers.From)
}

func (i *Item) To() string {
	if len(i.Envelope.Targets) > 0 {
		return i.Envelope.Targets
	}
	return findEmail(i.Message.Headers.To)
}

func findEmail(header string) string {
	a := emailRegexp.FindString(header)
	if len(a) < 2 {
		return ""
	}
	return a[1 : len(a)-1]
}
