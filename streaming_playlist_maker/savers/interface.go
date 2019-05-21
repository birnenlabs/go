package savers

import (
	"context"
	"fmt"
	"github.com/golang/glog"
)

type SaverJob struct {
	Playlist  string
	SaverType string
}

type Status struct {
	// True if playlist cache was used.
	PlaylistCached bool
	// True if song was added.
	SongAdded bool
	// Name of the song found by the saver, empty if not found.
	FoundTitle string
	// Title of the same (or similar) song that is already added to given playlist, when empty assuming that FoundTitle was just added.
	SimilarTitle string
	// Match quality 0-100
	MatchQuality int
}

type SongSaver interface {
	// Saves song to the playlist. This method should block and return the result of saving.
	Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error)
}

func Create(ctx context.Context, saverType string) (SongSaver, error) {
	glog.V(3).Infof("Creating %v saver", saverType)
	switch saverType {
	case "spotify":
		return newSpotify(ctx)
	case "stdout":
		return newStdout()
	default:
		return nil, fmt.Errorf("Invalid saver type definition (%v).", saverType)
	}
}
