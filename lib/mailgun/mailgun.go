package mailgun

import (
	"birnenlabs.com/lib/conf"
	"birnenlabs.com/lib/ratelimit"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiUser = "api"

const EventFailed = "failed"
const EventRejected = "rejected"
const EventStored = "stored"

type Mailgun struct {
	apiKey string
	domain string
	eu     string
	client ratelimit.AnyClient
}

func New() (*Mailgun, error) {
	var config Config
	err := conf.LoadConfigFromJson("mailgun", &config)
	if err != nil {
		return nil, err
	}
	eu := ""
	if config.Eu {
		eu = ".eu"
	}
	return &Mailgun{
		apiKey: config.ApiKey,
		domain: config.Domain,
		eu:     eu,
		client: ratelimit.New(&http.Client{}, time.Second),
	}, nil
}

func (m *Mailgun) SendEmail(email Email) error {
	uri := fmt.Sprintf("https://api%s.mailgun.net/v3/%s/messages", m.eu, m.domain)
	glog.Infof("SendEmail url: %v", uri)

	payload := createPayload(email)

	_, err := m.makePostRequest(uri, payload)
	return err
}

// Sends bounce email. Email.From is ignored.
func (m *Mailgun) SendBounceEmail(email Email, failedRecipient string) error {
	uri := fmt.Sprintf("https://api%s.mailgun.net/v3/%s/messages", m.eu, m.domain)
	glog.Infof("SendBounceEmail url: %v", uri)

	if !m.IsInMyDomain(email.To) {
		// Send bounces to our domain only.
		return fmt.Errorf("Bounces should be sent to own domain only")
	}

	// Ignore from
	email.From = fmt.Sprintf("Mail Delivery Subsystem <mailer-daemon@%s>", m.domain)

	payload := createPayload(email)
	payload.Add("h:Auto-Submitted", "auto-replied")
	payload.Add("h:Return-path", "<>")
	payload.Add("h:X-Failed-Recipients", failedRecipient)

	_, err := m.makePostRequest(uri, payload)
	return err

}

func (m *Mailgun) ListFailedEvents(begin, end int64) ([]Item, error) {
	return m.listEvents(begin, end, EventFailed)
}

func (m *Mailgun) ListRejectedEvents(begin, end int64) ([]Item, error) {
	return m.listEvents(begin, end, EventRejected)
}

func (m *Mailgun) ListStoredEvents(begin, end int64) ([]Item, error) {
	return m.listEvents(begin, end, EventStored)
}

func (m *Mailgun) ListAllEvents(begin, end int64) ([]Item, error) {
	return m.listEvents(begin, end, fmt.Sprintf("%s%%20OR%%20%s%%20OR%%20%s", EventFailed, EventRejected, EventStored))
}

func (m *Mailgun) listEvents(begin, end int64, event string) ([]Item, error) {
	result := make([]Item, 0)

	nextUrl := fmt.Sprintf("https://api%s.mailgun.net/v3/%s/events?event=%s&begin=%v&end=%v", m.eu, m.domain, event, begin, end)

	for nextUrl != "" {
		glog.Infof("ListEvents url: %v", nextUrl)
		body, err := m.makeGetRequest(nextUrl)
		if err != nil {
			return nil, err
		}
		var events EventsResponse
		err = json.Unmarshal(body, &events)
		if err != nil {
			return nil, err
		}
		result = append(result, events.Items...)

		glog.Infof("%v items were returned by the server", len(events.Items))
		if len(events.Items) == 0 {
			nextUrl = ""
		} else {
			nextUrl = events.Paging.Next
		}
	}
	return result, nil
}

// TO DELETE
func (m *Mailgun) groupItems(items []Item) map[Headers][]Item {
	result := make(map[Headers][]Item)
	for _, item := range items {
		result[item.Message.Headers] = append(result[item.Message.Headers], item)

	}
	return result
}

func (m *Mailgun) IsInMyDomain(address string) bool {
	return strings.HasSuffix(address, "@"+m.domain)
}

func (m *Mailgun) CreateAddress(user string) string {
	return user + "@" + m.domain
}

func (m *Mailgun) MailerDaemon() string {
	return fmt.Sprintf("Mail Delivery Subsystem <mailer-daemon@%s>", m.domain)
}

func createPayload(email Email) url.Values {
	payload := url.Values{}
	payload.Add("from", email.From)
	payload.Add("subject", email.Subject)
	payload.Add("text", email.Text)
	payload.Add("to", email.To)

	if len(email.Reference) > 0 {
		payload.Add("h:In-Reply-To", email.Reference)
		payload.Add("h:References", email.Reference)
	}

	return payload
}

func (m *Mailgun) makeGetRequest(uri string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(apiUser, m.apiKey)
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response code: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (m *Mailgun) makePostRequest(uri string, payload url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(apiUser, m.apiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response code: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
