package main

import (
	"birnenlabs.com/streaming_playlist_maker/savers"
	"birnenlabs.com/streaming_playlist_maker/sources"
)

type Job struct {
	Name   string
	Active bool
	sources.SourceJob
	savers.SaverJob
}
