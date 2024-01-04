package sources

import (
	"birnenlabs.com/go/lib/icy"
	"context"
	"github.com/golang/glog"
	"strings"
	"time"
)

type icySource struct {
}

func newIcy() SongSource {
	return &icySource{}
}

const timeout = time.Hour * 6

func (s *icySource) Start(ctx context.Context, conf SourceJob, song chan<- Song) error {
	// channel accepted by the icy listener
	title := make(chan string, 10)

	// Thread that is listening to icy stream and pushing data into title channel
	go s.startStreaming(title, song, conf)
	// Thread that is parsing title channel and putting it into songs channel.
	go s.monitorTitleChannel(ctx, title, song, conf)
	return nil
}

func (s *icySource) startStreaming(title chan string, song chan<- Song, conf SourceJob) {
	defer close(title)
	err := icy.OpenWithTimeout(conf.SourceUrl, title, timeout)
	glog.V(1).Infof("%v", err)

	song <- Song{
		Error: err,
	}
}

func (s *icySource) monitorTitleChannel(ctx context.Context, title <-chan string, song chan<- Song, conf SourceJob) {
	defer close(song)
	var t string
	ok := true
	for ok {
		t, ok = <-title
		if len(t) > 5 {
			for substr, replacement := range conf.SubstrMap {
				t = strings.Replace(t, substr, replacement, -1)
			}
			glog.V(2).Infof("Song found: %q", t)
			song <- Song{
				ArtistTitle: t,
				Error:       nil,
			}
		}
	}
}
