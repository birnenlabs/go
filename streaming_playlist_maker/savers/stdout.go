package savers

import (
	"context"
	"time"
)

type stdoutSaver struct{}

func newStdout() (SongSaver, error) {
	return &stdoutSaver{}, nil
}

func (s *stdoutSaver) Init(ctx context.Context, conf SaverJob) error {
	return nil
}

func (s *stdoutSaver) Save(ctx context.Context, conf SaverJob, artistTitle string) (*Status, error) {
	time.Sleep(time.Millisecond * 100)
	return &Status{
		MatchQuality: -99,
		FoundTitle:   "NOTHING WAS SAVED",
		SongAdded:    true,
	}, nil
}
