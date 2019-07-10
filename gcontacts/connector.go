package gcontacts

// Methods that are using Google Contacts GData API.

import (
	"birnenlabs.com/oauth"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
)

const urlBase = "https://www.google.com/m8/feeds/contacts/default/full?v=3.0"

type Gcontacts struct {
	httpClient *http.Client
}

func New(ctx context.Context) (*Gcontacts, error) {
	return NewWithCustomConfig(ctx, "gcontacts")
}

func NewWithCustomConfig(ctx context.Context, config string) (*Gcontacts, error) {
	// First create OAuth.
	oauthClient, err := oauth.Create(config)
	if err != nil {
		return nil, err
	}

	// Verify the token
	err = oauthClient.VerifyToken(ctx)
	if err != nil {
		return nil, err
	}

	// Get http client with Bearer
	httpClient, err := oauthClient.CreateAuthenticatedHttpClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Gcontacts{
		httpClient: httpClient,
	}, nil
}

func (s *Gcontacts) ListContacts(ctx context.Context) (*Feed, error) {
	url := urlBase + "&max-results=10000"

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Feed
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	glog.Infof("Found %d contacts.", len(result.Entries))

	return &result, nil
}

// TODO: maybe use Entry struct instead of string.
func (s *Gcontacts) AddContact(ctx context.Context, body string) error {
	url := urlBase

	resp, err := s.httpClient.Post(url, "text/plain", strings.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 201 == created
	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("response code: %v\n%v", resp.StatusCode, string(body))
	}
	return nil
}

// TODO: maybe use Entry struct instead of delete url
func (s *Gcontacts) RemoveContact(ctx context.Context, deleteUrl string) error {
	r, err := http.NewRequest(http.MethodDelete, deleteUrl, nil)
	r.Header.Set("If-Match", "*")

	resp, err := s.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("response code: %v\n%v", resp.StatusCode, string(body))
	}

	return nil
}
