package main

type SourceJob struct {
	SourceUrl  string
	SourceType string
	SubstrMap  map[string]string
}

type SaverJob struct {
	Playlist  string
	SaverType string
}

type Job struct {
	Name   string
	Active bool
	SourceJob
	SaverJob
}
