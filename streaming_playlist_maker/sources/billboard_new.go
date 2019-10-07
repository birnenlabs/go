package sources

import (
	"encoding/json"
	"github.com/golang/glog"
	"html"
	"strings"
	"time"
)

const (
	dataChartsJson = "data-charts=\""
)

type billboardNewSource struct {
	*webSource
}

type billboardJson struct {
	Artist string `json:"artist_name"`
	Title  string `json:"title"`
}

func newBillboardNew() *billboardNewSource {
	result := &billboardNewSource{}
	result.webSource = newWebSource(result.findSongsInHtml, result.generateHistoryUrl)
	return result
}

func (b *billboardNewSource) findSongsInHtml(s string) []string {
	if len(s) > 5000 {
		charts := strings.Index(s, dataChartsJson)
		if charts != -1 {
			jsonString := html.UnescapeString(s[charts+len(dataChartsJson) : len(s)-2])
			glog.V(3).Infof("Found json: %s...", jsonString[:5000])

			songs := make([]billboardJson, 0)
			decoder := json.NewDecoder(strings.NewReader(jsonString))
			err := decoder.Decode(&songs)
			if err != nil {
				glog.Errorf("Could not decode json: %v.", err)
			}
			glog.V(3).Infof("Found %d songs: %v", len(songs), songs)
			result := make([]string, 0)
			for _, s := range songs {
				result = append(result, s.String())
			}
			return result
		}
	}
	return []string{}
}

func (b *billboardNewSource) generateHistoryUrl(urlBase string, t time.Time) (string, time.Time) {
	return urlBase + t.Format("2006-01-02"), t.AddDate(0, 0, -7)
}

func (b *billboardJson) String() string {
	return b.Artist + " - " + b.Title
}
