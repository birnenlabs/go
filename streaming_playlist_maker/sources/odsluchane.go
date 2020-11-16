package sources

import (
	"fmt"
	"github.com/golang/glog"
	"net/url"
	"strings"
	"time"
)

const odsluchaneSpotifyUrl = "https://open.spotify.com/search/"

type odsluchaneSource struct {
	*webSource
}

func newOdsluchane() *odsluchaneSource {
	result := &odsluchaneSource{}
	result.webSource = newWebSource(result.findSongsInHtml, result.generateHistoryUrl)
	return result
}

func (o *odsluchaneSource) findSongsInHtml(s string) []string {
	idx := strings.Index(s, odsluchaneSpotifyUrl)
	if idx != -1 {
		s = s[idx+len(odsluchaneSpotifyUrl) : len(s)]
		idx = strings.Index(s, "\"")
		if idx != -1 {
			s = s[0:idx]
			s, _ = url.QueryUnescape(s)
			glog.V(3).Infof("Odsluchane: %v", s)
			return []string{s}
		}
	}
	return []string{}
}

func (o *odsluchaneSource) generateHistoryUrl(urlBase string, t time.Time) (string, time.Time) {
	return fmt.Sprintf("%v&m=%v&y=%v", urlBase, int(t.Month()), t.Year()), t.AddDate(0, -1, 0)
}
