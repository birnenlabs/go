package spotify

import (
	"fmt"
	"sync"
)

type Cache interface {
	// Playlist cache methods
	Add(playlistId string, track *ImmutableSpotifyTrack) error
	ReplaceAll(playlistId string, tracks []*ImmutableSpotifyTrack) error
	Remove(playlistId string, track *ImmutableSpotifyTrack) error
	Get(playlistId string) []*ImmutableSpotifyTrack
	IsCached(playlistId string) bool
}

type spotifyCache struct {
	playlist     map[string]*playlistCache
	playlistLock sync.RWMutex
}

type playlistCache struct {
	tracks     []*ImmutableSpotifyTrack
	tracksLock sync.RWMutex
}

func newCache() Cache {
	playlist := make(map[string]*playlistCache)

	return &spotifyCache{
		playlist: playlist,
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

func (s *spotifyCache) Remove(playlistId string, track *ImmutableSpotifyTrack) error {
	return s.getOrCreate(playlistId).remove(track)
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

func (p *playlistCache) remove(track *ImmutableSpotifyTrack) error {
	if track == nil {
		return fmt.Errorf("Cannot remove nil track")
	}

	p.tracksLock.Lock()
	defer p.tracksLock.Unlock()

	i := 0
	for _, t := range p.tracks {
		// if ids are different in place copy to the new array
		if t.Id() != track.Id() {
			p.tracks[i] = t
			i++
		}
	}
	// if the length did not change noting was discarded
	if len(p.tracks) == i {
		return fmt.Errorf("Track %q to remove was not found", track)
	}

	p.tracks = p.tracks[:i]
	return nil
}
