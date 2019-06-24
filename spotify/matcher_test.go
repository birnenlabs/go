package spotify

import (
	"strings"
	"testing"
)

type matchTest struct {
	track SpotifyTrack
	title string
	want  int
}

var tests = []matchTest{
	matchTest{
		want:  0,
		title: "Taylor Swift - Blank Space",
		track: makeTrack("As Made Famous by Taylor Swift", "Blank Space"),
	},
	matchTest{
		want:  0,
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Singers", "Blank Space tribute to Taylor"),
	},
	matchTest{
		want:  0,
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Taylor Swift", "Blank Space KARAOKE"),
	},
	matchTest{
		want:  100,
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Taylor Swift", "Blank Space"),
	},
	matchTest{
		want:  100,
		title: "Taylor Swift feat. Selena - Blank Space",
		track: makeTrack("Taylor Swift", "Selena", "Blank Space"),
	},
	matchTest{
		want:  100, // penalty not applied for title in radio song
		title: "Taylor Swift vs Dj Aligator - Blank Space Remix",
		track: makeTrack("Taylor Swift", "Dj Aligator", "Blank Space Remix"),
	},
	matchTest{
		want:  97, // -5 because of length difference
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Taylor Swift", "Blank Space New"),
	},
	matchTest{
		want:  92, // -5 because of lenght difference, -10 because of penalty
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Taylor Swift", "Blank Space Remix"),
	},
	matchTest{
		want:  85, // -10 because of lenght difference, -20 because of penalty
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Taylor Swift", "Blank Space Remix Live"),
	},
	matchTest{
		want:  50,
		title: "Taylor Swift - Blank Space",
		track: makeTrack("Taylor Swift", "Something else"),
	},
	matchTest{
		want:  52, // award for radio edit
		title: "Taylor Swift - Blank Space (Radio Edit)",
		track: makeTrack("Taylor Swift", "Something else (Radio Edit)"),
	},
	matchTest{
		want:  50,
		title: "Queen - Blank Space",
		track: makeTrack("Taylor Swift", "Blank Space"),
	},
	matchTest{
		want:  95, // 2 * -5 for size difference
		title: "abc - long title",
		track: makeTrack("abc", "very very long title"),
	},
	matchTest{
		want:  75, // only 50 points for title as not all words are in spotify
		title: "abc - very very long title",
		track: makeTrack("abc", "long title"),
	},
	matchTest{
		want:  100, // -5 for size difference, +5 for award word
		title: "abc - long title",
		track: makeTrack("abc", "long title single"),
	},
	matchTest{
		want:  100, // award words are equal to difference penalty
		title: "abc - long title",
		track: makeTrack("abc", "long title single radio remastered"),
	},
	matchTest{
		want:  100, // award not applied for word in radio title
		title: "abc - long title single",
		track: makeTrack("abc", "long title single"),
	},
	matchTest{
		want:  0,
		title: "abc- long title",
		track: makeTrack("abc", "long title"),
	},
	matchTest{
		want:  99, // 99 as all the words are ok, it is just their order.
		title: "Avicii feat. Sandro Cavazza - Without You",
		track: makeTrack("Avicii, Sandro Cavazza", "Without You (feat. Sandro Cavazza)"),
	},
	matchTest{
		want:  100,
		title: "こんにちは feat Здравствуйте - नमस्ते",
		track: makeTrack("こんにちは", "Здравствуйте", "नमस्ते"),
	},
}

func TestMatchRatio(t *testing.T) {
	for _, matchTest := range tests {
		got := CalculateMatchRatio(matchTest.title, matchTest.track.immutable())
		if got != matchTest.want {
			t.Errorf("radio: %q, spotify: %q, got: %v, want: %v", matchTest.title, matchTest.track, got, matchTest.want)
		}
	}
}

func makeTrack(artistsAndTitle ...string) SpotifyTrack {
	var artists []SpotifyArtist
	for _, a := range artistsAndTitle[:len(artistsAndTitle)-1] {
		artists = append(artists, SpotifyArtist{Name: a})
	}
	return SpotifyTrack{
		Name:    artistsAndTitle[len(artistsAndTitle)-1],
		Artists: artists,
		Id:      strings.Join(artistsAndTitle, ":"),
	}
}
