package main

import (
	"context"
)

type Song struct {
	ArtistTitle string
	Error       error
}

type SaverStatus struct {
	// True if playlist cache was used.
	PlaylistCached bool
	// Name of the song found by the saver, empty if not found.
	FoundTitle string
	// Title of the same (or similar) song that is already added to given playlist, when empty assuming that FoundTitle was just added.
	SimilarTitle string
}

type SongSource interface {
	// Starts the song source. This method should start the background thread and return.
	// When new song is found it should be send to the channel. Subsequent calls to "start"
	// should raise an error. Channel should be closed when there is no more songs left.
	Start(ctx context.Context, conf SourceJob, song chan<- Song) error
}

type SongSaver interface {
	// Saves song to the playlist. This method should block and return the result of saving.
	Save(ctx context.Context, conf SaverJob, artistTitle string) (*SaverStatus, error)
}
