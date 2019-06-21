package spotify

// Methods that are using spotify API. They are using rate limited client.

import (
	"context"
)

func (s *SpotifyCached) AddToPlaylist(ctx context.Context, playlistId string, track *ImmutableSpotifyTrack) error {
	err := s.cl.AddToPlaylist(ctx, playlistId, track.Id())
	if err != nil {
		return err
	}

	// add to cache
	return nil
}

func (s *SpotifyCached) ListPlaylist(ctx context.Context, playlistId string) ([]*ImmutableSpotifyTrack, error) {
	// check cache first
	tracks, err := s.cl.ListPlaylist(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmutableSpotifyTrack, 0)
	for _, track := range tracks {
		result = append(result, track.immutable())
	}
	return result, nil
}

func (s *SpotifyCached) FindTracks(ctx context.Context, query string) ([]*ImmutableSpotifyTrack, error) {
	tracks, err := s.cl.FindTracks(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmutableSpotifyTrack, 0)
	for _, track := range tracks {
		result = append(result, track.immutable())
	}
	return result, nil
}
