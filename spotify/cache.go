package spotify

import (
	"fmt"
	"sync"
)

type Cache interface {
	Add(playlistId string, track *ImmutableSpotifyTrack) error
	ReplaceAll(playlistId string, tracks []*ImmutableSpotifyTrack) error
	Replace(playlistId string, oldTrack *ImmutableSpotifyTrack, newTrack *ImmutableSpotifyTrack) error
	Get(playlistId string) []*ImmutableSpotifyTrack
	IsCached(playlistId string) bool
}

type spotifyCache struct {
	playlist          map[string]*playlistCache
	notFoundCache     map[string]*ImmutableSpotifyTrack
	playlistLock      sync.RWMutex
	notFoundCacheLock sync.RWMutex
}

type playlistCache struct {
	tracks     []*ImmutableSpotifyTrack
	tracksLock sync.RWMutex
}

func newCache() Cache {
	playlist := make(map[string]*playlistCache)
	notFoundCache := make(map[string]*ImmutableSpotifyTrack)

	return &spotifyCache{
		playlist:      playlist,
		notFoundCache: notFoundCache,
	}
}

func (s *spotifyCache) getOrCreate(playlistId string) *playlistCache {
	s.playlistLock.Lock()
	defer s.playlistLock.Unlock()

	p, ok := s.playlist[playlistId]
	if ok {
		return p
	} else {
		p = &playlistCache{}
		s.playlist[playlistId] = p
		return p
	}
}

func (s *spotifyCache) Add(playlistId string, track *ImmutableSpotifyTrack) error {
	return s.getOrCreate(playlistId).add(track)
}

func (s *spotifyCache) Get(playlistId string) []*ImmutableSpotifyTrack {
	return s.getOrCreate(playlistId).get()
}

func (s *spotifyCache) ReplaceAll(playlistId string, tracks []*ImmutableSpotifyTrack) error {
	return s.getOrCreate(playlistId).replaceAll(tracks)
}

func (s *spotifyCache) Replace(playlistId string, oldTrack *ImmutableSpotifyTrack, newTrack *ImmutableSpotifyTrack) error {
	return s.getOrCreate(playlistId).replace(oldTrack, newTrack)
}

func (s *spotifyCache) IsCached(playlistId string) bool {
	return s.getOrCreate(playlistId).size() > 0
}

func (p *playlistCache) add(track *ImmutableSpotifyTrack) error {
	if track == nil {
		return fmt.Errorf("Cannot add nil track")
	}
	p.tracksLock.Lock()
	defer p.tracksLock.Unlock()
	p.tracks = append(p.tracks, track)
	return nil
}

func (p *playlistCache) get() []*ImmutableSpotifyTrack {
	p.tracksLock.RLock()
	defer p.tracksLock.RUnlock()

	return append([]*ImmutableSpotifyTrack{}, p.tracks...)
}

func (p *playlistCache) size() int {
        p.tracksLock.RLock()
        defer p.tracksLock.RUnlock()

	return len(p.tracks)
}

func (p *playlistCache) replaceAll(tracks []*ImmutableSpotifyTrack) error {
	p.tracksLock.Lock()
	defer p.tracksLock.Unlock()
	p.tracks = append([]*ImmutableSpotifyTrack{}, tracks...)
	return nil
}

func (p *playlistCache) replace(oldTrack *ImmutableSpotifyTrack, newTrack *ImmutableSpotifyTrack) error {
	if oldTrack == nil {
		return fmt.Errorf("Cannot replace nil track")
	}

	p.tracksLock.Lock()
	defer p.tracksLock.Unlock()

	for i, t := range p.tracks {
		if t.Id() == oldTrack.Id() {
			if newTrack == nil {
				p.tracks = append(p.tracks[:i], p.tracks[i+1:]...)
			} else {
				p.tracks[i] = newTrack
			}
			return nil
		}
	}
	return fmt.Errorf("Track %q to replace was not found", oldTrack)
}
