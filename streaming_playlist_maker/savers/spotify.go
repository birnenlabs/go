package savers

import (
	"birnenlabs.com/spotify"
	"context"
	"fmt"
	"github.com/golang/glog"
	"sync"
)

const validMatch = 75

type spotifySaver struct {
	spotify *spotify.Spotify
	cache   map[string][]spotify.SpotifyTrack
	rwLock  sync.RWMutex
}

// TODO: market should be a parameter
func newSpotify(ctx context.Context) (SongSaver, error) {
	cache := make(map[string][]spotify.SpotifyTrack)
	s, err := spotify.New(ctx, "pl")
	if err != nil {
		return nil, err
	}

	return &spotifySaver{
		spotify: s,
		cache:   cache,
	}, nil
}

func (s *spotifySaver) Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error) {
	glog.V(3).Infof("Saving song: %v", artistTitle)

	bestTrack, bestTrackMatch, err := s.findSpotifyTrack(ctx, artistTitle)
	if err != nil {
		return nil, err
	}

	result := &Status{
		// common fields
		FoundTitle:   fmt.Sprintf("%v", bestTrack),
		MatchQuality: bestTrackMatch,
		// default fields
		SongAdded:      false,
		PlaylistCached: false,
		SimilarTitle:   "",
	}

	if bestTrackMatch >= validMatch {

		// First check if our song was already added there - let's use cache first!
		similarTrack := s.findSimilarSongInCache(conf.Playlist, artistTitle)
		if similarTrack != nil {
			result.PlaylistCached = true
			result.SimilarTitle = similarTrack.String()
			return result, nil

		}

		// If not found update cache and try again
		err = s.updateCache(ctx, conf.Playlist)
		if err != nil {
			return nil, err
		}
		similarTrack = s.findSimilarSongInCache(conf.Playlist, artistTitle)
		if similarTrack != nil {
			result.SimilarTitle = similarTrack.String()
			return result, nil

		}

		// If not found - Add to the playlist
		err = s.spotify.AddToPlaylist(ctx, conf.Playlist, bestTrack.Id)
		if err != nil {
			return nil, err
		}

		s.addToCache(conf.Playlist, *bestTrack)
		result.SongAdded = true
		return result, nil
	} else {
		return result, nil
	}
}

func (s *spotifySaver) findSpotifyTrack(ctx context.Context, artistTitle string) (*spotify.SpotifyTrack, int, error) {
	if len(artistTitle) == 0 {
		return nil, 0, fmt.Errorf("Empty song title")
	}

	tracks, err := s.spotify.FindTracks(ctx, artistTitle)
	if err != nil {
		return nil, 0, err
	}

	bestTrackMatch := -1
	var bestTrack spotify.SpotifyTrack
	for _, track := range tracks {
		currentMatch := spotify.CalculateMatchRatio(artistTitle, track)
		if currentMatch > bestTrackMatch {
			bestTrackMatch = currentMatch
			bestTrack = track
		}
		if bestTrackMatch == 100 {
			break
		}
	}

	return &bestTrack, bestTrackMatch, nil
}

func (s *spotifySaver) findSimilarSongInCache(playlistId string, artistTitle string) *spotify.SpotifyTrack {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	for _, track := range s.cache[playlistId] {
		if spotify.CalculateMatchRatio(artistTitle, track) >= validMatch {
			return &track
		}
	}
	return nil
}

func (s *spotifySaver) updateCache(ctx context.Context, playlistId string) error {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	tracks, err := s.spotify.ListPlaylist(ctx, playlistId)
	if err != nil {
		return err
	}
	s.cache[playlistId] = tracks
	return nil
}

func (s *spotifySaver) addToCache(playlistId string, track spotify.SpotifyTrack) {
        s.rwLock.Lock()
        defer s.rwLock.Unlock()

        s.cache[playlistId] = append(s.cache[playlistId], track)
}

