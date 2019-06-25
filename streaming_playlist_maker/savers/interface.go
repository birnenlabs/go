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
	// True if song was added.
	SongAdded bool
	// True if song already existed in the playlsit
	SongExists bool
	// Title of the song found by the saver, empty if not found.
	FoundTitle string
	// Match quality 0-100
	MatchQuality int
}

type SongSaver interface {
	// Cleans the playlist
	Clean(ctx context.Context, conf SaverJob) error

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
