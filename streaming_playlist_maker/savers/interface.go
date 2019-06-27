package savers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/glog"
	"strconv"
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

type CleanStatus struct {
	// Number of songs that are unavailable
	Unavailable int
	// Number of songs that were unavailable but replacement was found
	UnavailableReplaced int
	// Number of duplicates that were removed
	Duplicates int
	// List of possible duplicates that were not removed
	Similar []*SimilarTrack
}

type SimilarTrack struct {
	Title1        string
	Title2        string
	AvgMatchRatio int
}

type SongSaver interface {
	// Cleans the playlist
	Clean(ctx context.Context, conf SaverJob) (*CleanStatus, error)

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

func (c *CleanStatus) String() string {
	var buf bytes.Buffer
	buf.WriteString("Unavailable:          ")
	buf.WriteString(strconv.Itoa(c.Unavailable))
	buf.WriteString("\nUnavailable replaced: ")
	buf.WriteString(strconv.Itoa(c.UnavailableReplaced))
	buf.WriteString("\nRemoved duplicates:   ")
	buf.WriteString(strconv.Itoa(c.Duplicates))
	for _, s := range c.Similar {
		buf.WriteString("\n")
		buf.WriteString(strconv.Itoa(s.AvgMatchRatio))
		buf.WriteString(" ")
		buf.WriteString(s.Title1)
		buf.WriteString(" = ")
		buf.WriteString(s.Title2)
	}
	return buf.String()
}
