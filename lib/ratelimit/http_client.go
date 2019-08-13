package ratelimit

import (
	"github.com/golang/glog"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Interface of client that is accepted by this class. http.Client is satisfying these dependencies.
type AnyClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}

type httpClient struct {
	client AnyClient
	t      <-chan time.Time
}

func New(client AnyClient, minInterval time.Duration) AnyClient {
	return &httpClient{
		client: client,
		t:      time.Tick(minInterval),
	}
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	c.throttle()
	return c.client.Do(req)
}

func (c *httpClient) Get(url string) (*http.Response, error) {
	c.throttle()
	return c.client.Get(url)
}

func (c *httpClient) Head(url string) (*http.Response, error) {
	c.throttle()
	return c.client.Head(url)
}

func (c *httpClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	c.throttle()
	return c.client.Post(url, contentType, body)
}

func (c *httpClient) PostForm(url string, data url.Values) (*http.Response, error) {
	c.throttle()
	return c.client.PostForm(url, data)
}

func (c *httpClient) throttle() {
	glog.V(3).Info("Waiting for throttle.")
	<-c.t
	glog.V(3).Info("Obtained throttle.")
}
