package sources

import (
	"context"
	"fmt"
	"github.com/golang/glog"
)

type SourceJob struct {
	SourceUrl  string
	SourceType string
	SubstrMap  map[string]string
}

type Song struct {
	ArtistTitle string
	Error       error
}

type SongSource interface {
	// Starts the song source. This method should start the background thread and return.
	// When new song is found it should be send to the channel. Channel should be closed
	// when there is no more songs left.
	Start(ctx context.Context, conf SourceJob, song chan<- Song) error
}

func Create(ctx context.Context, sourceType string) (SongSource, error) {
	glog.V(3).Infof("Creating %v source", sourceType)
	switch sourceType {
	case "icy":
		return newIcy(), nil
	case "odsluchane":
		return newOdsluchane(), nil
	case "billboard":
		return newBillboard(), nil
	default:
		return nil, fmt.Errorf("Invalid source type definition (%v).", sourceType)
	}
}
