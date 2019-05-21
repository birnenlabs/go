package sources

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/glog"
	"html"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type odsluchaneSource struct {
	r *rand.Rand
}

const odsluchaneSpotifyUrl = "https://open.spotify.com/search/results/"
const initialSleep = 30

func newOdsluchane() *odsluchaneSource {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return &odsluchaneSource{
		r: r,
	}
}

func (o *odsluchaneSource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {
	urlParts := strings.Split(conf.SourceUrl, "|")

	if len(urlParts) == 1 {
		go o.doStart(song, conf.SourceUrl)
	} else if len(urlParts) == 3 {
		start, err := time.Parse("2006-01-02", urlParts[1])
		if err != nil {
			return err
		}
		end, err := time.Parse("2006-01-02", urlParts[2])
		if err != nil {
			return err
		}
		go o.doStartHistory(song, urlParts[0], start, end)
	} else {
		return fmt.Errorf("Too many url parts.")
	}
	return nil
}

func (o *odsluchaneSource) doStart(song chan<- Song, url string) {
	defer close(song)

	glog.V(3).Infof("Starting odsluchane with url: %v", url)

	// Sleep between 30 to 60 seconds to avoid many requests to the server.
	time.Sleep(time.Second * time.Duration(initialSleep+o.r.Intn(initialSleep)))
	songs, err := o.findSongsInPage(url)
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

func (o *odsluchaneSource) doStartHistory(song chan<- Song, urlBase string, start time.Time, end time.Time) {
	defer close(song)

	glog.V(3).Infof("Starting historical odsluchane %v-%v with url: %v", start, end, urlBase)

	t := end
	for !t.Before(start) {
		url := fmt.Sprintf("%v&m=%v&y=%v", urlBase, int(t.Month()), t.Year())

		songs, err := o.findSongsInPage(url)
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

		t = t.AddDate(0, -1, 0)
	}
}

func (o *odsluchaneSource) findSongsInPage(url string) ([]string, error) {
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
		idx := strings.Index(s, odsluchaneSpotifyUrl)
		if idx != -1 {
			s = s[idx+len(odsluchaneSpotifyUrl) : len(s)]
			idx = strings.Index(s, "\"")
			if idx != -1 {
				s = s[0:idx]
				s = html.UnescapeString(s)
				result = append(result, s)
			}
		}
	}
	glog.V(3).Infof("%v returned %v results", url, len(result))
	return result, nil
}
