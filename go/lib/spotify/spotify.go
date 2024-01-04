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
	// Add to cache only if playlist is already cached (this is to avoid creating cache before list)
	if s.cache.IsCached(playlistId) {
		err := s.cache.Add(playlistId, track)
		if err != nil {
			return err
		}
	}

	err := s.connector.addToPlaylist(ctx, playlistId, track.Id())
	if err != nil {
		// Try to remove what was added to cache in case of error
		err2 := s.cache.Remove(playlistId, track)
		if err2 != nil {
			glog.Errorf("Error returned from cache when removing incorrectly added track: %v", err2)
		}
		return err
	}

	return nil
}

func (s *Spotify) RemoveFromPlaylist(ctx context.Context, playlistId string, track *ImmutableSpotifyTrack) error {
	err := s.cache.Remove(playlistId, track)
	if err != nil {
		return err
	}

	err = s.connector.removeFromPlaylist(ctx, playlistId, track.Id())
	if err != nil {
		// Try to add what was removed to cache in case of error
		err2 := s.cache.Add(playlistId, track)
		if err2 != nil {
			glog.Errorf("Error returned from cache when adding incorrectly removed track: %v", err2)
		}
		return err
	}

	return nil
}

func (s *Spotify) ListLiked(ctx context.Context) ([]*ImmutableSpotifyTrack, error) {
	tracks, err := s.connector.listLiked(ctx)
	if err != nil {
		return nil, err
	}

	// Not using cache for liked songs for now.
	result := make([]*ImmutableSpotifyTrack, len(tracks))
	for i := range tracks {
		result[i] = tracks[i].immutable()
	}

	glog.V(2).Infof("ListLiked: %v", tracks)
	return result, nil

}

func (s *Spotify) ListPlaylist(ctx context.Context, playlistId string) ([]*ImmutableSpotifyTrack, error) {
	if s.cache.IsCached(playlistId) {
		glog.V(1).Infof("Found cached tracks for %v, not connecting to spotify.", playlistId)
		return s.cache.Get(playlistId), nil
	}
	glog.V(1).Infof("Found 0 cached tracks for %v, connecting to spotify.", playlistId)

	//Ignoring the error as the method updates cache
	return s.ListPlaylistWithFilter(ctx, playlistId, func(s SpotifyTrack) bool { return true })
}

func (s *Spotify) ListPlaylistWithFilter(ctx context.Context, playlistId string, filter func(SpotifyTrack) bool) ([]*ImmutableSpotifyTrack, error) {
	tracks, err := s.connector.listPlaylist(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	// Add songs to cache
	cached := make([]*ImmutableSpotifyTrack, len(tracks))
	result := make([]*ImmutableSpotifyTrack, 0)
	for i := range tracks {
		imm := tracks[i].immutable()
		cached[i] = imm
		if filter(tracks[i]) {
			result = append(result, imm)
		}
	}

	s.cache.ReplaceAll(playlistId, cached)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("ListPlaylistWithFilter: %v", tracks)
	return result, nil
}

func (s *Spotify) FindTracks(ctx context.Context, query string) ([]*ImmutableSpotifyTrack, error) {
	tracks, err := s.connector.findTracks(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmutableSpotifyTrack, len(tracks))
	for i := range tracks {
		result[i] = tracks[i].immutable()
	}
	return result, nil
}
