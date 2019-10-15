package ratelimit

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

type limitTest struct {
	t          time.Duration
	requests   int
	expTimeSec int
}

type myHttpClient struct {
	*http.Client
	wg *sync.WaitGroup
}

func (c *myHttpClient) Get(url string) (*http.Response, error) {
	c.wg.Done()
	return nil, nil
}

var tests = []limitTest{
	limitTest{
		t:          time.Millisecond * 250, // 4 requests per second
		requests:   4,
		expTimeSec: 1,
	},
	limitTest{
		t:          time.Millisecond,
		requests:   1001,
		expTimeSec: 1,
	},
	limitTest{
		t:          time.Millisecond,
		requests:   5,
		expTimeSec: 0,
	},
}

func TestHttpClientIsAccepted(t *testing.T) {
	New(&http.Client{}, time.Second)
}

func TestTimeElapsed(t *testing.T) {
	for _, limitTest := range tests {
		var wg sync.WaitGroup
		c := New(&myHttpClient{wg: &wg}, limitTest.t)
		start := time.Now()
		wg.Add(limitTest.requests)
		for i := 0; i < limitTest.requests; i++ {
			go c.Get("")
		}
		wg.Wait()
		elapsed := time.Since(start)
		if int(elapsed.Seconds()) != limitTest.expTimeSec {
			t.Errorf("got: %v, want: %v", elapsed.Seconds(), limitTest.expTimeSec)
		}
	}
}
