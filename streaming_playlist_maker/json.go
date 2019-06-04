package main

import (
	"birnenlabs.com/streaming_playlist_maker/savers"
	"birnenlabs.com/streaming_playlist_maker/sources"
)

type Configuration struct {
	Jobs        []Job
	EmailNotify string
}

type Job struct {
	Name   string
	Active bool
	sources.SourceJob
	savers.SaverJob
}
