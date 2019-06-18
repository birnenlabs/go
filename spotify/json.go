package spotify

// Structures used to parse JSON returned by Spotify API

import (
	"strings"
)

type SpotifyArtist struct {
	Id   string
	Name string
}

type SpotifyAlbum struct {
	Id      string
	Name    string
	Artists []SpotifyArtist
}

type SpotifyTrack struct {
	Id               string
	Name             string
	DurationMs       int64 `json:"duration_ms"`
	Popularity       int
	Artists          []SpotifyArtist
	Album            SpotifyAlbum
	AvailableMarkets []string `json:"available_markets"`
}

type PlaylistItem struct {
	Track SpotifyTrack
}

type PlaylistResponse struct {
	Items []PlaylistItem
	Total int
	Next  string
}

type SearchResponseBody struct {
	Items []SpotifyTrack
}

type SearchResponse struct {
	Tracks SearchResponseBody
}

type ImmutableSpotifyTrack struct {
	id     string
	title  string
	artist string
}

func (t SpotifyTrack) String() string {
	artist := t.ArtistAsString()
	if len(artist)+len(t.Name) > 0 {
		return t.ArtistAsString() + " - " + t.Name
	} else {
		return ""
	}
}

func (t SpotifyTrack) ArtistAsString() string {
	var artists = []string{}
	for _, artist := range t.Artists {
		artists = append(artists, artist.Name)
	}
	return strings.Join(artists, ", ")
}

func (t *SpotifyTrack) Immutable() *ImmutableSpotifyTrack {
	return &ImmutableSpotifyTrack{
		id:     t.Id,
		title:  t.Name,
		artist: t.ArtistAsString(),
	}
}

func (t *ImmutableSpotifyTrack) Id() string {
	return t.id
}

func (t *ImmutableSpotifyTrack) Title() string {
	return t.title
}

func (t *ImmutableSpotifyTrack) Artist() string {
	return t.artist
}

func (t *ImmutableSpotifyTrack) String() string {
        if len(t.artist)+len(t.title) > 0 {
                return t.artist + " - " + t.title
        } else {
                return ""
        }
}

