package savers

import (
	"birnenlabs.com/spotify"
	"context"
	"fmt"
	"github.com/golang/glog"
)

const validMatch = 75

// TODO: this should be in config.
const spotifyMarket = "PL"

type spotifySaver struct {
	spotify  *spotify.Spotify
	notFound *nfCache
}

func newSpotify(ctx context.Context) (SongSaver, error) {
	s, err := spotify.New(ctx, spotifyMarket)
	if err != nil {
		return nil, err
	}

	return &spotifySaver{
		spotify:  s,
		notFound: newCache(),
	}, nil
}

func (s *spotifySaver) Clean(ctx context.Context, conf SaverJob) (*CleanStatus, error) {
	// Replace unplayable should be first as it uses ListPlaylistWithFilter method that always connects to spotify.
	unplayable, err := s.replaceUnplayable(ctx, conf.Playlist)
	if err != nil {
		return nil, err
	}

	duplicates, err := s.findDuplicatesById(ctx, conf.Playlist)
	if err != nil {
		return nil, err
	}

	similarTracks, err := s.findDuplicatesByName(ctx, conf.Playlist)
	if err != nil {
		return nil, err
	}

	return &CleanStatus{
		Unavailable:         unplayable.Unavailable,
		UnavailableReplaced: unplayable.UnavailableReplaced,
		Duplicates:          duplicates,
		Similar:             similarTracks,
	}, nil
}

func (s *spotifySaver) Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error) {
	glog.V(2).Infof("Saving song: %v", artistTitle)

	if len(artistTitle) == 0 {
		return nil, fmt.Errorf("Empty song title")
	}

	// First check if the track is in not found cache
	cachedStatus := s.notFound.IsNotFound(artistTitle)
	if cachedStatus != nil {
		return cachedStatus, nil
	}

	// Then check if the track is already in playlist
	existingTracks, err := s.spotify.ListPlaylist(ctx, conf.Playlist)
	if err != nil {
		return nil, err
	}

	existingTrack, existingTrackMatch := s.findBestMatch(existingTracks, artistTitle)
	glog.V(2).Infof("Best match from existing songs %q for %q (%d).", existingTrack, artistTitle, existingTrackMatch)
	if existingTrackMatch >= validMatch {
		return &Status{
			FoundTitle:   existingTrack.String(),
			MatchQuality: existingTrackMatch,
			SongAdded:    false,
			SongExists:   true,
		}, nil
	}

	// If not in the playlist search for it in spotify
	newTracks, err := s.spotify.FindTracks(ctx, artistTitle)
	if err != nil {
		return nil, err
	}
	newTrack, newTrackMatch := s.findBestMatch(newTracks, artistTitle)

	// if new track is a good match add it to the playlist
	if newTrackMatch >= validMatch {
		err = s.spotify.AddToPlaylist(ctx, conf.Playlist, newTrack)
		if err != nil {
			return nil, err
		}

		return &Status{
			FoundTitle:   newTrack.String(),
			MatchQuality: newTrackMatch,
			SongAdded:    true,
			SongExists:   false,
		}, nil
	}

	// If there was no match add it to cache
	status := &Status{
		FoundTitle:   newTrack.String(),
		MatchQuality: newTrackMatch,
		SongAdded:    false,
		SongExists:   false,
	}
	s.notFound.AddNotFound(artistTitle, status)
	return status, nil
}

func (s *spotifySaver) findBestMatch(tracks []*spotify.ImmutableSpotifyTrack, artistTitle string) (*spotify.ImmutableSpotifyTrack, int) {
	bestTrackMatch := -1
	var bestTrack *spotify.ImmutableSpotifyTrack
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

	return bestTrack, bestTrackMatch
}

func (s *spotifySaver) replaceUnplayable(ctx context.Context, playlistId string) (*CleanStatus, error) {
	trueForUnavailable := func(track spotify.SpotifyTrack) bool {
		for _, market := range track.AvailableMarkets {
			if market == spotifyMarket {
				return false
			}
		}
		return true
	}

	tracks, err := s.spotify.ListPlaylistWithFilter(ctx, playlistId, trueForUnavailable)
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("[%v] Found %d unavailable tracks.", playlistId, len(tracks))

	unavailable := 0
	unavailableReplaced := 0
	for _, t := range tracks {
		artistTitle := t.String()
		if len(artistTitle) == 0 {
			// If artist - titile is empty song is removed.
			glog.V(1).Infof("[%v] Removing invalid song: %#v", playlistId, t)
			err = s.spotify.RemoveFromPlaylist(ctx, playlistId, t)
			if err != nil {
				return nil, fmt.Errorf("error while removing: %q when removing wrong song %#v", err, t)
			}
		} else {
			// If not available try to find replacement
			newTracks, err := s.spotify.FindTracks(ctx, artistTitle)
			if err != nil {
				return nil, err
			}

			newTrack, newTrackMatch := s.findBestMatch(newTracks, artistTitle)
			if newTrackMatch >= validMatch {
				unavailableReplaced++
				glog.V(1).Infof("[%v] Replacing track:  %3d %q -> %q", playlistId, newTrackMatch, artistTitle, newTrack)
				// We have a good match
				// first remove old song
				err = s.spotify.RemoveFromPlaylist(ctx, playlistId, t)
				if err != nil {
					return nil, fmt.Errorf("error while removing: %q during the process of replacing %q with %q", err, artistTitle, newTrack)
				}

				// then add new song
				err = s.spotify.AddToPlaylist(ctx, playlistId, newTrack)
				if err != nil {
					return nil, fmt.Errorf("error while ADDING: %q during the process of replacing %q with %q - song was removed but new song was not added", err, artistTitle, newTrack)
				}
			} else {
				unavailable++
				glog.V(1).Infof("[%v] Unavailable track: %3d %q -> %q", playlistId, newTrackMatch, artistTitle, newTrack)
			}
		}
	}
	return &CleanStatus{
		Unavailable:         unavailable,
		UnavailableReplaced: unavailableReplaced,
	}, nil
}

func (s *spotifySaver) findDuplicatesById(ctx context.Context, playlistId string) (int, error) {
	tracks, err := s.spotify.ListPlaylist(ctx, playlistId)
	if err != nil {
		return 0, err
	}

	toRemove := make(map[string]*spotify.ImmutableSpotifyTrack)
	for i, t1 := range tracks[0 : len(tracks)-1] {
		for _, t2 := range tracks[i+1:] {
			if t1.Id() == t2.Id() {
				toRemove[t1.Id()] = t1
			}
		}
	}

	for _, t := range toRemove {
		glog.V(1).Infof("[%v] Removing duplicates: %q", playlistId, t)

		// remove all occurences of the song
		err = s.spotify.RemoveFromPlaylist(ctx, playlistId, t)
		if err != nil {
			return 0, fmt.Errorf("error while removing: %q during the process of removing duplicates of %q", err, t)
		}

		// then add it back once
		err = s.spotify.AddToPlaylist(ctx, playlistId, t)
		if err != nil {
			return 0, fmt.Errorf("error while ADDING: %q during the process of removing duplicates of %q - song was removed but new song was not added", err, t)
		}
	}

	return len(toRemove), nil
}

func (s *spotifySaver) findDuplicatesByName(ctx context.Context, playlistId string) ([]*SimilarTrack, error) {
	tracks, err := s.spotify.ListPlaylist(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	result := make([]*SimilarTrack, 0)
	for i, t1 := range tracks[0 : len(tracks)-1] {
		for _, t2 := range tracks[i+1:] {
			match12 := spotify.CalculateMatchRatio(t1.String(), t2)
			match21 := spotify.CalculateMatchRatio(t2.String(), t1)
			if match12+match21 >= 2*validMatch {
				glog.V(1).Infof("[%v] %3d %3d %q==%q", playlistId, match12, match21, t1, t2)
				result = append(result, &SimilarTrack{
					Title1:        t1.String(),
					Title2:        t2.String(),
					AvgMatchRatio: (match12 + match21) / 2,
				})
			}
		}
	}
	return result, nil
}
