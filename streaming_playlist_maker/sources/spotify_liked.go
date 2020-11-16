package sources

import (
	"birnenlabs.com/lib/spotify"
	"context"
)

// TODO: this should be in config.
const spotifyMarket = "PL"

type spotifyLikedSource struct {
	spotify *spotify.Spotify
}

func newSpotifyLiked() SongSource {
	return &spotifyLikedSource{}
}

func (s *spotifyLikedSource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {

	sp, err := spotify.New(ctx, spotifyMarket)
	if err != nil {
		close(song)
		return err
	}

	s.spotify = sp
	go s.listSongs(ctx, conf, song)

	return nil
}

func (s *spotifyLikedSource) listSongs(ctx context.Context, conf SourceJob, song chan<- Song) {
	defer close(song)

	tracks, err := s.spotify.ListLiked(ctx)

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
