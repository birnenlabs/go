package savers

import (
	"context"
	"time"
)

type stdoutSaver struct{}

func newStdout() (SongSaver, error) {
	return &stdoutSaver{}, nil
}

func (s *stdoutSaver) Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error) {
	time.Sleep(time.Second)
	return &Status{
		MatchQuality: -99,
		FoundTitle:   "NOTHING WAS SAVED",
		SongAdded:    true,
	}, nil
}