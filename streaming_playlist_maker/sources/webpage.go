package sources

import (
	"birnenlabs.com/lib/ratelimit"
	"context"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type webSource struct {
	httpClient      ratelimit.AnyClient
	findSongsInHtml func(line string) []string
	// Generates url for a given date and the previous valid timepoint (e.g. if page generates new content every week, returned time should be t minus week).
	generateHistoryUrl func(urlBase string, t time.Time) (string, time.Time)
	delimiter          string
}

func newWebSource(findSongsInHtml func(html string) []string, generateHistoryUrl func(urlBase string, t time.Time) (string, time.Time)) *webSource {
	return newWebSourcePage("\n", findSongsInHtml, generateHistoryUrl)
}

func newWebSourcePage(delimiter string, findSongsInHtml func(html string) []string, generateHistoryUrl func(urlBase string, t time.Time) (string, time.Time)) *webSource {
	return &webSource{
		findSongsInHtml:    findSongsInHtml,
		delimiter:          delimiter,
		generateHistoryUrl: generateHistoryUrl,
		httpClient:         ratelimit.New(&http.Client{}, time.Second*3),
	}
}

func (w *webSource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {
	if w.findSongsInHtml == nil {
		return fmt.Errorf("findSongsInHtml not set")
	}
	if w.generateHistoryUrl == nil {
		return fmt.Errorf("generateHistoryUrl not set")
	}

	urlParts := strings.Split(conf.SourceUrl, "|")

	if len(urlParts) == 1 {
		go w.doStart(song, conf.SourceUrl)
	} else if len(urlParts) == 3 {
		start, err := time.Parse("2006-01-02", urlParts[1])
		if err != nil {
			return err
		}
		end, err := time.Parse("2006-01-02", urlParts[2])
		if err != nil {
			return err
		}
		go w.doStartHistory(song, urlParts[0], start, end)
	} else {
		return fmt.Errorf("Too many url parts.")
	}
	return nil
}

func (w *webSource) doStart(song chan<- Song, url string) {
	defer close(song)

	glog.V(3).Infof("Starting web source with url: %v", url)

	songs, err := w.findSongsInPage(url)
	if err != nil {
		song <- Song{
			Error: err,
		}
	}
	for _, s := range songs {
		song <- Song{
			ArtistTitle: s,
			Error:       nil,
		}
	}
}

func (w *webSource) doStartHistory(song chan<- Song, urlBase string, start time.Time, end time.Time) {
	defer close(song)

	glog.V(1).Infof("Starting historical web source %v-%v with url: %v", start, end, urlBase)

	t := end
	for !t.Before(start) {
		url, nextTs := w.generateHistoryUrl(urlBase, t)
		songs, err := w.findSongsInPage(url)
		if err != nil {
			song <- Song{
				Error: err,
			}
		}
		glog.V(2).Infof("%v returned %v songs", url, len(songs))
		for _, s := range songs {
			song <- Song{
				ArtistTitle: s,
				Error:       nil,
			}
		}
		if !nextTs.Before(t) {
			song <- Song{
				Error: fmt.Errorf("Timestamp returned by generateHistoryUrl (%v) not before current (%v)", nextTs, t),
			}
			break
		}
		t = nextTs
	}
}

func (w *webSource) findSongsInPage(url string) ([]string, error) {
	resp, err := w.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, s := range strings.Split(string(body), w.delimiter) {
		found := w.findSongsInHtml(s)
		result = append(result, found...)
	}

	glog.V(3).Infof("%q returned %v results", url, len(result))
	if len(result) == 0 {
		return nil, fmt.Errorf("No results for %q", url)
	}
	return result, nil
}
