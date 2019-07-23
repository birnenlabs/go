package mailgun

import (
	"birnenlabs.com/conf"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const apiUser = "api"

type Mailgun struct {
	apiKey string
	domain string
	eu     string
	client *http.Client
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
		client: &http.Client{},
	}, nil
}

func (m *Mailgun) SendEmail(email Email) error {
	uri := fmt.Sprintf("https://api%s.mailgun.net/v3/%s/messages", m.eu, m.domain)
	glog.Infof("SendEmail url: %v", uri)

	payload := url.Values{}
	payload.Add("from", email.From)
	payload.Add("subject", email.Subject)
	payload.Add("text", email.Text)
	for _, to := range email.To {
		payload.Add("to", to)
	}

	if len(email.Reference) > 0 {
		payload.Add("h:In-Reply-To", email.Reference)
		payload.Add("h:References", email.Reference)
	}

	_, err := m.makePostRequest(uri, payload)
	return err
}

// Sends bounce email. Email.From should be set to the failed recipient.
func (m *Mailgun) SendBounceEmail(email Email) error {
	uri := fmt.Sprintf("https://api%s.mailgun.net/v3/%s/messages", m.eu, m.domain)
	glog.Infof("SendBounceEmail url: %v", uri)

	payload := url.Values{}
	payload.Add("from", fmt.Sprintf("Mail Delivery Subsystem <mailer-daemon@%s>", m.domain))
	payload.Add("subject", email.Subject)
	payload.Add("text", email.Text)
	for _, to := range email.To {
		payload.Add("to", to)
	}

	payload.Add("h:Auto-Submitted", "auto-replied")
	payload.Add("h:Return-path", "<>")
	payload.Add("h:X-Failed-Recipients", email.From)
	if len(email.Reference) > 0 {
		payload.Add("h:In-Reply-To", email.Reference)
		payload.Add("h:References", email.Reference)
	}

	_, err := m.makePostRequest(uri, payload)
	return err

}

func (m *Mailgun) ListFailedEvents() ([]Item, error) {
	return m.ListFailedEventsTimeRange(4000000000, 0) // 4000000000 = 2096.10.02
}

func (m *Mailgun) ListFailedEventsTimeRange(begin, end int64) ([]Item, error) {
	result := make([]Item, 0)

	nextUrl := fmt.Sprintf("https://api%s.mailgun.net/v3/%s/events?event=rejected%%20OR%%20failed&begin=%v&end=%v", m.eu, m.domain, begin, end)

	for nextUrl != "" {
		glog.Infof("ListFailedEvent url: %v", nextUrl)
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
