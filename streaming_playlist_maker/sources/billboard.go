package sources

import (
	"github.com/golang/glog"
	"html"
	"strings"
	"time"
)

type billboardSource struct {
	*webSource
}

const (
	delimiter = "\""
	bbArtist  = "data-artist=" + delimiter
	bbTitle   = "data-title=" + delimiter
)

func newBillboard() *billboardSource {
	result := &billboardSource{}
	result.webSource = newWebSource(result.findSongsInHtml, result.generateHistoryUrl)
	return result
}

func (b *billboardSource) findSongsInHtml(s string) []string {
	// Line syntax: <div class="chart-list-item  " data-rank="2" data-artist="Artist" data-title="Title" data-has-content="true">
	if strings.Contains(s, "chart-list-item") {
		glog.V(3).Infof("Found match in line %s", s)
		idxA := strings.Index(s, bbArtist)
		idxT := strings.Index(s, bbTitle)
		if idxA == -1 || idxT == -1 {
			return []string{}
		}

		artistPrefix := s[idxA+len(bbArtist):]
		titlePrefix := s[idxT+len(bbTitle):]

		idxA = strings.Index(artistPrefix, delimiter)
		idxT = strings.Index(titlePrefix, delimiter)

		if idxA == -1 || idxT == -1 {
			return []string{}
		}

		return []string{html.UnescapeString(artistPrefix[:idxA]) + " - " + html.UnescapeString(titlePrefix[:idxT])}
	}
	return []string{}
}

func (b *billboardSource) generateHistoryUrl(urlBase string, t time.Time) (string, time.Time) {
	return urlBase + t.Format("2006-01-02"), t.AddDate(0, 0, -7)
}
