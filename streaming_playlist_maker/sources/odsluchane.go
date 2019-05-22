package sources

import (
	"github.com/golang/glog"
	"html"
	"strings"
)

const odsluchaneSpotifyUrl = "https://open.spotify.com/search/results/"

type odsluchaneSource struct {
	*webSource
}

func newOdsluchane() *odsluchaneSource {
	result := &odsluchaneSource{}
	result.webSource = newWebSource(result.findSongsInHtml)
	return result
}

func (o *odsluchaneSource) findSongsInHtml(s string) []string {
	idx := strings.Index(s, odsluchaneSpotifyUrl)
	if idx != -1 {
		s = s[idx+len(odsluchaneSpotifyUrl) : len(s)]
		idx = strings.Index(s, "\"")
		if idx != -1 {
			s = s[0:idx]
			s = html.UnescapeString(s)
			glog.V(3).Infof("Odsluchane: %v", s)
			return []string{s}
		}
	}
	return []string{}
}
