package main

import (
	glog "birnenlabs.com/lib/alog"
	"birnenlabs.com/lib/conf"
	"birnenlabs.com/streaming_playlist_maker/savers"
	"birnenlabs.com/streaming_playlist_maker/sources"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const appName = "Streaming playlist maker"

var config = flag.String("config", "streaming-playlist-maker", "Configuration")

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	defer glog.Flush()
	ctx := context.Background()

	glog.UseFormattedPayload(appName)

	var jobs []Job
	err := conf.LoadConfigFromJson(*config, &jobs)
	if err != nil {
		glog.Exit("Could not load config: ", err)
	}
	glog.V(3).Infof("Loaded configuration: %+v", jobs)

	sourcesMap, saversMap, err := createSourcesAndSavers(ctx, jobs)
	if err != nil {
		glog.Exit("Could not create sources and savers: ", err)
	}

	glog.Infof("Cleaning savers")
	err = cleanSavers(ctx, saversMap, jobs)
	if err != nil {
		glog.Exit("Could not clean savers: ", err)
	}

	glog.Infof("Starting jobs")
	stats := &statistics{}
	var wg sync.WaitGroup
	for _, conf := range jobs {
		if !conf.Active {
			continue
		}

		err = stats.Init(conf.Name)
		if err != nil {
			glog.Exitf("Could initialize stats: %v", err)
		}

		wg.Add(1)
		go func(ctx context.Context, conf Job, source sources.SongSource, saver savers.SongSaver, stats *statistics) {
			defer wg.Done()
			startJob(ctx, conf, source, saver, stats)
		}(ctx, conf, sourcesMap[conf.SourceType], saversMap[conf.SaverType], stats)
	}

	handleCtrlC(stats)
	go printStatsSometimes(stats)

	wg.Wait()
	glog.Infof("Jobs completed")

	issues := stats.FindIssues()
	glog.Infof("Statistics:\n%v%v", issues, stats)

	glog.InfoSend("\n" + stats.String())
	if len(issues) > 0 {
		glog.Error(issues)
	}
}

func createSourcesAndSavers(ctx context.Context, jobs []Job) (map[string]sources.SongSource, map[string]savers.SongSaver, error) {
	sourcesMap := make(map[string]sources.SongSource)
	saversMap := make(map[string]savers.SongSaver)
	for _, conf := range jobs {
		_, ok := sourcesMap[conf.SourceType]
		if !ok {
			source, err := sources.Create(ctx, conf.SourceType)
			if err != nil {
				return nil, nil, err
			}
			sourcesMap[conf.SourceType] = source
		}

		_, ok = saversMap[conf.SaverType]
		if !ok {
			saver, err := savers.Create(ctx, conf.SaverType)
			if err != nil {
				return nil, nil, err
			}
			saversMap[conf.SaverType] = saver
		}

	}
	return sourcesMap, saversMap, nil

}

func handleCtrlC(stats *statistics) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(stats *statistics) {
		<-c
		glog.Infof("Statistics:\n%v", stats)
		glog.Flush()
		os.Exit(0)
	}(stats)
}

func printStatsSometimes(stats *statistics) {
	for true {
		time.Sleep(time.Second * 60)
		glog.Infof("Statistics:\n%v", stats)
	}
}

func cleanSavers(ctx context.Context, s map[string]savers.SongSaver, jobs []Job) error {
	errors := make(chan error)

	for _, conf := range jobs {
		saver, ok := s[conf.SaverType]
		if !ok {
			return fmt.Errorf("Saver not found for %v", conf)
		}

		go func(ctx context.Context, s savers.SongSaver, conf Job, errors chan<- error) {
			start := time.Now()
			status, err := s.Clean(ctx, conf.SaverJob)
			if err != nil {
				glog.Errorf("[%15.15s] Cleaned after %v, error: %v", conf.Name, time.Now().Sub(start), err)
			} else {
				glog.Infof("[%15.15s] Cleaned after %v, stats:\n%v", conf.Name, time.Now().Sub(start), status)
			}
			errors <- err
		}(ctx, saver, conf, errors)
	}

	var err error
	for _ = range jobs {
		err = <-errors
		if err != nil {
			return err
		}
	}
	return nil
}

func startJob(ctx context.Context, conf Job, source sources.SongSource, saver savers.SongSaver, stats *statistics) {
	glog.Infof("[%15.15s] Starting: %v -> %v (%T -> %T).", conf.Name, conf.SourceType, conf.SaverType, source, saver)
	ch := make(chan sources.Song, 10)
	err := source.Start(ctx, conf.SourceJob, ch)
	if err != nil {
		glog.Errorf("[%15.15s] Could not start job: %v.", conf.Name, err)
		return
	}

	var song sources.Song
	ok := true
	for ok {
		song, ok = <-ch
		if song.Error == nil && ok {
			status, err := saver.Save(ctx, conf.SaverJob, song.ArtistTitle)

			if err != nil {
				glog.Errorf("[%15.15s] ERROR %q: %v", conf.Name, song.ArtistTitle, err)
				stats.Error(conf.Name, song.ArtistTitle, err)
			} else {
				if status.SongAdded {
					// Song added
					glog.Infof("[%15.15s] A %3d %q -> %q added", conf.Name, status.MatchQuality, song.ArtistTitle, status.FoundTitle)
					stats.Added(conf.Name, song.ArtistTitle)
				} else if status.SongExists {
					stats.Exists(conf.Name, song.ArtistTitle)
					glog.Infof("[%15.15s] E %3d %q -> %q exists", conf.Name, status.MatchQuality, song.ArtistTitle, status.FoundTitle)
				} else {
					// not added and not exists -> not found
					stats.NotFound(conf.Name, song.ArtistTitle)
					glog.Infof("[%15.15s] N %3d %q -> %q not added", conf.Name, status.MatchQuality, song.ArtistTitle, status.FoundTitle)
				}
			}
		} else if song.Error != nil {
			glog.Infof("[%15.15s] Error: %v", conf.Name, song.Error)
		}
		// TODO include channel size in stats
	}
	glog.Infof("[%15.15s] Source stopped.", conf.Name)
}
