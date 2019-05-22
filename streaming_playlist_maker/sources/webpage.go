package sources

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type webSource struct {
	r               *rand.Rand
	findSongsInHtml func(line string) []string
}

const initialSleep = 5

func newWebSource(findSongsInHtml func(line string) []string) *webSource {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return &webSource{
		r:               r,
		findSongsInHtml: findSongsInHtml,
	}
}

func (w *webSource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {
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

	// Sleep between 30 to 60 seconds to avoid many requests to the server.
	time.Sleep(time.Second * time.Duration(initialSleep+w.r.Intn(initialSleep)))
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

	glog.V(3).Infof("Starting historical web source %v-%v with url: %v", start, end, urlBase)

	t := end
	for !t.Before(start) {
		// Skip christmas songs
		if int(t.Month()) != 12 {
			url := fmt.Sprintf("%v&m=%v&y=%v", urlBase, int(t.Month()), t.Year())

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

		t = t.AddDate(0, -1, 0)
	}
}

func (w *webSource) findSongsInPage(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []string

	reader := bufio.NewReader(resp.Body)
	for err == nil {
		var b []byte
		b, err = reader.ReadBytes('\n')
		s := string(b)

		found := w.findSongsInHtml(s)
		result = append(result, found...)
	}
	glog.V(3).Infof("%q returned %v results", url, len(result))
	if len(result) == 0 {
		return nil, fmt.Errorf("No results for %q", url)
	}
	return result, nil
}
