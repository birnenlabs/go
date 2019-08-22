package spotify

import (
	"github.com/golang/glog"
	"regexp"
	"strings"
)

var wordMatcher = regexp.MustCompile("[\\p{L}\\d]+")

var TerribleSongNames = []string{
	"acapella",
	"acappella",
	"as made famous",
	"in the style of",
	"in style of",
	"karaoke",
	"made famous by",
	"originally performed by",
	"reprise",
	"tribute",
}

var penaltyWords = []string{
	"acoustic",
	"instrumental",
	"live",
	"unplugged",
	"remix",
}

var awardWords = []string{
	"radio",
	"remastered",
	"single",
}

// This should contain the same things as "awardWords". These expressions will be removed from radio title to avoid matches:
// "artist - song1 [radio edit]" == "artist - song2 [radio edit]"
var awardExpressions = []string{
	"radio edit",
	"remastered",
	"single edit",
}

var artistJoiners = []string{
	"feat",
	"vs",
}

// Calculates match ratio between song name stored in a string from radio (e.g. "Artist - Some song")
// and a given SpotifyTrack. Returns match ratio from 0 to 100, anything below 75 is bad quality,
// while 50 and less is probably worthless.
func CalculateMatchRatio(radio string, spotify *ImmutableSpotifyTrack) int {
	radioArtistTitle := strings.SplitN(radio, " - ", 2)
	if len(radioArtistTitle) != 2 {
		glog.Warningf("Could not split artist+title: %q.", radio)
		return 0
	}

	spotifyString := strings.ToLower(spotify.String())
	for _, terribleName := range TerribleSongNames {
		if strings.Contains(spotifyString, terribleName) {
			glog.V(3).Infof("Terrible name: %v", spotify)
			return 0
		}
	}

	spotifyTitle := strings.ToLower(spotify.Title())
	spotifyArtist := strings.ToLower(spotify.Artist())

	radioTitle := strings.ToLower(radioArtistTitle[1])
	radioArtist := strings.ToLower(radioArtistTitle[0])
	for _, awardExpression := range awardExpressions {
		radioTitle = strings.Replace(radioTitle, awardExpression, "", -1)
	}

	radioArtistArray := wordMatcher.FindAllString(radioArtist, -1)
	radioTitleArray := wordMatcher.FindAllString(radioTitle, -1)
	spotifyArtistArray := wordMatcher.FindAllString(spotifyArtist, -1)
	spotifyTitleArray := wordMatcher.FindAllString(spotifyTitle, -1)

	artistMatch := calculateMatchRatioArray(radioArtistArray, spotifyArtistArray)
	titleMatch := calculateMatchRatioArray(radioTitleArray, spotifyTitleArray)
	result := (artistMatch + titleMatch) / 2

	if result < 100 {
		// Trying to match all the words but in reverse - if everything from spotify is in
		// radio array, it means that we are still good.
		combinedMatch := calculateMatchRatioArray(
			append(spotifyArtistArray, spotifyTitleArray...),
			append(radioArtistArray, radioTitleArray...))
		glog.V(3).Infof("Result less than 100, trying combined match %v.", combinedMatch)
		if combinedMatch == 100 {
			// 99 so we would not override the proper artist/title match.
			result = 99
		}
	}
	glog.V(3).Infof("CalculateMatchRatio result: %v.", result)
	return result
}

// Returns match ratio of string arrays.
func calculateMatchRatioArray(radio []string, spotify []string) int {
	if len(radio) == 0 || len(spotify) == 0 {
		glog.V(3).Infof("radio: %v, spotify: %v, result: 0", radio, spotify)
		return 0
	}

	result := 0
	for _, r := range radio {
		if contains(r, spotify) || contains(r, artistJoiners) {
			result = result + 100
		}
	}
	result = result / len(radio)
	// penalty for size difference
	result = max(0, result-max(0, 5*(len(spotify)-len(radio))))
	glog.V(3).Infof("radio: %v, spotify: %v, initial result: %v", radio, spotify, result)

	for _, penalty := range penaltyWords {
		if !contains(penalty, radio) && contains(penalty, spotify) {
			result = max(0, result-10)
			glog.V(3).Infof("radio: %v, spotify: %v, penalty for: %q", radio, spotify, penalty)
		}
	}

	// Award is cancelling size difference penalty
	for _, award := range awardWords {
		if !contains(award, radio) && contains(award, spotify) {
			result = min(100, result+5)
			glog.V(3).Infof("radio: %v, spotify: %v, award for: %q", radio, spotify, award)
		}
	}
	glog.V(3).Infof("radio: %v, spotify: %v, result: %v", radio, spotify, result)
	return result
}

func contains(s string, list []string) bool {
	for _, w := range list {
		if w == s {
			return true
		}
	}
	return false
}

func min(i1 int, i2 int) int {
	if i1 > i2 {
		return i2
	}
	return i1
}

func max(i1 int, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}
