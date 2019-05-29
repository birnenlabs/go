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
	spotify           *spotify.Spotify
	playlistCache     map[string][]spotify.SpotifyTrack
	notFoundCache     map[string]spotify.SpotifyTrack
	playlistCacheLock sync.RWMutex
	notFoundCacheLock sync.RWMutex
}

// TODO: market should be a parameter
func newSpotify(ctx context.Context) (SongSaver, error) {
	playlistCache := make(map[string][]spotify.SpotifyTrack)
	notFoundCache := make(map[string]spotify.SpotifyTrack)
	s, err := spotify.New(ctx, "pl")
	if err != nil {
		return nil, err
	}

	return &spotifySaver{
		spotify:       s,
		playlistCache: playlistCache,
		notFoundCache: notFoundCache,
	}, nil
}

func (s *spotifySaver) Init(ctx context.Context, conf SaverJob) error {
	glog.V(3).Infof("Initializing cache for playlist: %v", conf.Playlist)
	return s.updateCache(ctx, conf.Playlist)
}

func (s *spotifySaver) Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error) {
	glog.V(3).Infof("Saving song: %v", artistTitle)

	// First check if our song is already in cache
	similarTrack, similarTrackMatch := s.findSongInPlaylistCache(conf.Playlist, artistTitle)
	if similarTrackMatch >= validMatch {
		return &Status{
			FoundTitle:   similarTrack.String(),
			MatchQuality: similarTrackMatch,
			SongAdded:    false,
			SongExists:   true,
			Cached:       true,
		}, nil
	} else {
		// If not in playlist cache, let's check in notFound cache:
		s.notFoundCacheLock.RLock()
		track, ok := s.notFoundCache[artistTitle]
		s.notFoundCacheLock.RUnlock()

		if ok {
			return &Status{
				FoundTitle:   track.String(),
				MatchQuality: spotify.CalculateMatchRatio(artistTitle, track),
				SongAdded:    false,
				SongExists:   false,
				Cached:       true,
			}, nil
		}
	}

	// If not found in caches, let's try to search in spotify
	newTrack, newTrackMatch, err := s.findSpotifyTrack(ctx, artistTitle)
	if err != nil {
		return nil, err
	}

	// if new track is good match add it to the playlist
	if newTrackMatch >= validMatch {

		// Temporarily removing cache check to speed up adding process.
		//		// but update cache first and check if it is not there already
		//		err := s.updateCache(ctx, conf.Playlist)
		//		if err != nil {
		//			return nil, err
		//		}
		//		similarTrack, similarTrackMatch = s.findSongInPlaylistCache(conf.Playlist, artistTitle)
		//		if similarTrackMatch >= validMatch {
		//			return &Status{
		//				FoundTitle:   newTrack.String(),
		//				MatchQuality: similarTrackMatch,
		//				SongAdded:    false,
		//				Cached:       false,
		//				SimilarTitle: similarTrack.String(),
		//			}, nil
		//		}

		err = s.spotify.AddToPlaylist(ctx, conf.Playlist, newTrack.Id)
		if err != nil {
			return nil, err
		}

		// Also add to cache
		s.addToCache(conf.Playlist, *newTrack)

		return &Status{
			FoundTitle:   newTrack.String(),
			MatchQuality: newTrackMatch,
			SongAdded:    true,
			SongExists:   false,
			Cached:       false,
		}, nil
	} else {
		// Add not found track to cache
		s.notFoundCacheLock.Lock()
		s.notFoundCache[artistTitle] = *newTrack
		s.notFoundCacheLock.Unlock()

		return &Status{
			FoundTitle:   newTrack.String(),
			MatchQuality: newTrackMatch,
			SongAdded:    false,
			SongExists:   false,
			Cached:       false,
		}, nil
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

func (s *spotifySaver) findSongInPlaylistCache(playlistId string, artistTitle string) (*spotify.SpotifyTrack, int) {
	s.playlistCacheLock.RLock()
	defer s.playlistCacheLock.RUnlock()

	bestTrackMatch := -1
	var bestTrack spotify.SpotifyTrack
	for _, track := range s.playlistCache[playlistId] {
		currentMatch := spotify.CalculateMatchRatio(artistTitle, track)
		if currentMatch > bestTrackMatch {
			bestTrackMatch = currentMatch
			bestTrack = track
		}
		if bestTrackMatch == 100 {
			break
		}

	}
	return &bestTrack, bestTrackMatch
}

func (s *spotifySaver) updateCache(ctx context.Context, playlistId string) error {
	tracks, err := s.spotify.ListPlaylist(ctx, playlistId)
	if err != nil {
		return err
	}

	s.playlistCacheLock.Lock()
	defer s.playlistCacheLock.Unlock()

	s.playlistCache[playlistId] = tracks
	return nil
}

func (s *spotifySaver) addToCache(playlistId string, track spotify.SpotifyTrack) {
	s.playlistCacheLock.Lock()
	defer s.playlistCacheLock.Unlock()

	s.playlistCache[playlistId] = append(s.playlistCache[playlistId], track)
}
