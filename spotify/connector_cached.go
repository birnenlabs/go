package spotify

import (
	"context"
	"github.com/golang/glog"
)

func (s *SpotifyCached) AddToPlaylist(ctx context.Context, playlistId string, track *ImmutableSpotifyTrack) error {
	err := s.cl.AddToPlaylist(ctx, playlistId, track.Id())
	if err != nil {
		return err
	}

	return s.cache.Add(playlistId, track)
}

func (s *SpotifyCached) ListPlaylist(ctx context.Context, playlistId string) ([]*ImmutableSpotifyTrack, error) {
	cachedTracks := s.cache.Get(playlistId)
	if len(cachedTracks) > 0 {
		glog.V(1).Infof("Found %d cached tracks for %v, not connecting to spotify.", len(cachedTracks), playlistId)
		return cachedTracks, nil
	}
	glog.V(1).Infof("Found 0 cached tracks for %v, connecting to spotify.", playlistId)

	tracks, err := s.cl.ListPlaylist(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmutableSpotifyTrack, 0)
	for i := range tracks {
		result = append(result, tracks[i].immutable())
	}

	err = s.cache.ReplaceAll(playlistId, result)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("ListPlaylist cached: %v", result)
	return result, nil
}

func (s *SpotifyCached) FindTracks(ctx context.Context, query string) ([]*ImmutableSpotifyTrack, error) {
	// TODO add not found cache
	tracks, err := s.cl.FindTracks(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmutableSpotifyTrack, 0)
	for i := range tracks {
		result = append(result, tracks[i].immutable())
	}
	return result, nil
}
