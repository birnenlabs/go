package mailgun

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
	From      string
	To        string
	Subject   string
	Text      string
	Reference string
}
