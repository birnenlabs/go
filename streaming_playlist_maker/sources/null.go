package sources

import (
	"context"
)

type nullSource struct {
}

func newNull() SongSource {
	return &nullSource{}
}

func (s *nullSource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {
	close(song)
	return nil
}
