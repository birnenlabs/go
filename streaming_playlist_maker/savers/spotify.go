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
	spotify *spotify.Spotify
}

func newSpotify(ctx context.Context) (SongSaver, error) {
	s, err := spotify.New(ctx, spotifyMarket)
	if err != nil {
		return nil, err
	}

	return &spotifySaver{
		spotify: s,
	}, nil
}

func (s *spotifySaver) Clean(ctx context.Context, conf SaverJob) error {
	//	s.cleanPlaylist(ctx, conf.Playlist)
	return nil
}

func (s *spotifySaver) Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error) {
	glog.V(3).Infof("Saving song: %v", artistTitle)

	if len(artistTitle) == 0 {
		return nil, fmt.Errorf("Empty song title")
	}

	// First check if the track is already in playlist
	existingTracks, err := s.spotify.ListPlaylist(ctx, conf.Playlist)
	if err != nil {
		return nil, err
	}

	existingTrack, existingTrackMatch := s.findBestMatch(existingTracks, artistTitle)
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

	return &Status{
		FoundTitle:   newTrack.String(),
		MatchQuality: newTrackMatch,
		SongAdded:    false,
		SongExists:   false,
	}, nil
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

func (s *spotifySaver) cleanPlaylist(ctx context.Context, playlistId string, tracks []*spotify.ImmutableSpotifyTrack) {

	for _, t := range tracks {
		if !isAvailable(t) {
			artistTitle := t.String()
			if len(artistTitle) == 0 {
				glog.Infof("Something wrong with the song: %#v", t)
				//			} else {
				//				track, match, err := s.findSpotifyTrack(ctx, artistTitle)
				//				if err != nil {
				//					glog.Errorf("Could not find track %q: %v", artistTitle, err)
				//				} else {
				//					glog.Infof("Unavailable track: %3d %q -> %q", match, artistTitle, track)
				//					if match >= validMatch {
				//						//s.replaceInCache(playlistId, t, *track)
				//					}
				//				}
			}
		}
	}

	for i, t1 := range tracks[0 : len(tracks)-1] {
		for _, t2 := range tracks[i+1:] {
			match12 := spotify.CalculateMatchRatio(t1.String(), t2)
			match21 := spotify.CalculateMatchRatio(t2.String(), t1)
			if match12 >= validMatch || match21 >= validMatch {
				glog.Infof("%v %3d %3d %q==%q", playlistId, match12, match21, t1, t2)
			}
		}
	}
}

func isAvailable(track *spotify.ImmutableSpotifyTrack) bool {
	//	for _, market := range track.AvailableMarkets {
	//		if market == spotifyMarket {
	//			return true
	//		}
	//	}
	return false
}
