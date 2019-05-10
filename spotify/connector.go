package spotify

// Methods that are using spotify API. They are using rate limited client.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/golang/glog"
)

func (s *Spotify) AddToPlaylist(ctx context.Context, playlistId string, trackId string) error {
	url := fmt.Sprintf(
		"https://api.spotify.com/v1/playlists/%s/tracks?uris=spotify:track:%s",
		playlistId, trackId)

	glog.V(1).Infof("Add to playlist url: %q.", url)
	resp, err := s.httpClient.Post(url, "text/plain", bytes.NewReader(nil))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 201 == created
	if resp.StatusCode != 201 {
		return fmt.Errorf("response code: %v", resp.StatusCode)
	}
	return nil
}

func (s *Spotify) ListPlaylist(ctx context.Context, playlistId string) ([]SpotifyTrack, error) {
	result := make([]SpotifyTrack, 0)

	nextUrl := fmt.Sprintf(
		"https://api.spotify.com/v1/playlists/%s/tracks",
		playlistId)
	for nextUrl != "" {
		glog.V(2).Infof("List playlist url: %q.", nextUrl)
		resp, err := s.httpClient.Get(nextUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("response code: %v", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var r = new(PlaylistResponse)
		err = json.Unmarshal(body, &r)
		if err != nil {
			return nil, err
		}
		nextUrl = r.Next
		for _, playlistItem := range r.Items {
			result = append(result, playlistItem.Track)
		}
	}

	return result, nil
}

func (s *Spotify) FindTracks(ctx context.Context, query string) ([]SpotifyTrack, error) {
	url := fmt.Sprintf(
		"https://api.spotify.com/v1/search?type=track&market=%s&limit=50&q=%s",
		"pl", url.QueryEscape(updateQueryString(query)))

	glog.V(1).Infof("Find tracks url: %q.", url)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response code: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r = new(SearchResponse)
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Found %v tracks for query %q.", len(r.Tracks.Items), query)
	return r.Tracks.Items, nil
}

func updateQueryString(query string) string {
	for _, token := range []string{"&", "feat.", "feat", "vs.", "vs"} {
		query = strings.Replace(query, token, "", -1)
	}
	return query
}
