package sources

import (
	"html"
	"strings"
	"time"
)

type ukSinglesSource struct {
	*webSource
}

const (
	ukDelimiter = "/"
	ukArtist    = "<a href=\"/artist/"
	ukTitle     = "<a href=\"/search/singles/"
)

func newUkSingles() *ukSinglesSource {
	result := &ukSinglesSource{}
	result.webSource = newWebSource(result.findSongsInHtml, result.generateHistoryUrl)
	result.webSource.Delimiter = "</td>"
	result.webSource.SongLimit = 10
	return result
}

func (b *ukSinglesSource) findSongsInHtml(s string) []string {
	idxT := strings.Index(s, ukTitle)
	if idxT != -1 {
		idxA := strings.Index(s, ukArtist)
		if idxA != -1 {
			artistPrefix := s[idxA+len(ukArtist):]
			titlePrefix := s[idxT+len(ukTitle):]

			artist := strings.SplitN(artistPrefix, ukDelimiter, 3)
			title := strings.SplitN(titlePrefix, ukDelimiter, 2)

			if len(artist) == 3 && len(title) == 2 {
				return []string{html.UnescapeString(strings.ReplaceAll(artist[1], "-", " ")) + " - " + html.UnescapeString(strings.ReplaceAll(title[0], "-", " "))}
			}
		}
	}
	return []string{}
}

func (b *ukSinglesSource) generateHistoryUrl(urlBase string, t time.Time) (string, time.Time) {
	return strings.ReplaceAll(urlBase, "{DATE}", t.Format("20060102")), t.AddDate(0, 0, -7)
}
