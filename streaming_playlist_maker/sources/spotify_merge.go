package sources

import (
	"birnenlabs.com/lib/spotify"
	"context"
	"strings"
)

type spotifyMergeSource struct {
	spotify *spotify.Spotify
}

func newSpotifyMerge() SongSource {
	return &spotifyMergeSource{}
}

func (s *spotifyMergeSource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {

	sp, err := spotify.New(ctx, "PL")
	if err != nil {
		close(song)
		return err
	}

	s.spotify = sp
	playlists := strings.Split(conf.SourceUrl, "|")
	for _, playlist := range playlists {
		go s.listSongs(ctx, playlist, song)
	}

	return nil
}

func (s *spotifyMergeSource) listSongs(ctx context.Context, playlist string, song chan<- Song) {
	defer close(song)

	tracks, err := s.spotify.ListPlaylist(ctx, playlist)

	if err != nil {
		song <- Song{
			Error: err,
		}
	}

	for _, t := range tracks {
		song <- Song{
			ArtistTitle: t.String(),
			Error:       nil,
		}
	}
}
