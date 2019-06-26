package spotify

import (
	"context"
	"github.com/golang/glog"
)

type Spotify struct {
	connector *connector
	cache     Cache
}

func New(ctx context.Context, market string) (*Spotify, error) {
	c, err := newConnector(ctx, market)
	if err != nil {
		return nil, err
	}
	return &Spotify{
		connector: c,
		cache:     newCache(),
	}, nil
}

func (s *Spotify) AddToPlaylist(ctx context.Context, playlistId string, track *ImmutableSpotifyTrack) error {
	// If playlist is cached add to its cache
	if s.cache.IsCached(playlistId) {
		err := s.cache.Add(playlistId, track)
		if err != nil {
			return err
		}
	}

	err := s.connector.addToPlaylist(ctx, playlistId, track.Id())
	if err != nil {
		// Try to remove what was added to cache in case of error
		err2 := s.cache.Replace(playlistId, track, nil)
		if err2 != nil {
			glog.Errorf("Error returned from cache when removing incorrectly added track: %v", err2)
		}
		return err
	}

	return nil
}

func (s *Spotify) ListPlaylist(ctx context.Context, playlistId string) ([]*ImmutableSpotifyTrack, error) {
	if s.cache.IsCached(playlistId) {
		glog.V(1).Infof("Found cached tracks for %v, not connecting to spotify.", playlistId)
		return s.cache.Get(playlistId), nil
	}
	glog.V(1).Infof("Found 0 cached tracks for %v, connecting to spotify.", playlistId)

	//Ingoring the error as the method updates cache
	_, err := s.ListPlaylistNoCache(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	return s.cache.Get(playlistId), nil
}

func (s *Spotify) ListPlaylistNoCache(ctx context.Context, playlistId string) ([]SpotifyTrack, error) {
	result, err := s.connector.listPlaylist(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	// Add songs to cache
	cached := make([]*ImmutableSpotifyTrack, len(result))
	for i := range result {
		cached[i] = result[i].Immutable()
	}

	s.cache.ReplaceAll(playlistId, cached)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("ListPlaylistNoCache cached: %v", result)
	return result, nil
}

func (s *Spotify) FindTracks(ctx context.Context, query string) ([]*ImmutableSpotifyTrack, error) {
	tracks, err := s.connector.findTracks(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmutableSpotifyTrack, len(tracks))
	for i := range tracks {
		result[i] = tracks[i].Immutable()
	}
	return result, nil
}
