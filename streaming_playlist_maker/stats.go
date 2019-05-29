package main

import (
	"fmt"
	"github.com/golang/glog"

	"sync"
)

type aggregatedStatus struct {
	added          int64
	notFoundCached int64
	notFound       int64
	existsCached   int64
	exists         int64
	errors         int64
}

type statistics struct {
	m    map[string]*aggregatedStatus
	lock sync.RWMutex
}

func (s *statistics) Init(jobName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.m == nil {
		s.m = make(map[string]*aggregatedStatus)
	}
	_, ok := s.m[jobName]
	if ok {
		return fmt.Errorf("%v already initialized", jobName)
	}

	s.m[jobName] = &aggregatedStatus{}
	return nil
}

// Song was added to the playlist
func (s *statistics) Added(jobName string, artistTitle string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[jobName].added++
}

// Song was not found by saver.
func (s *statistics) NotFound(jobName string, artistTitle string, cached bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if cached {
		s.m[jobName].notFoundCached++
	} else {
		s.m[jobName].notFound++
	}
}

// Song was found by the saver but it already exists
func (s *statistics) Exists(jobName string, artistTitle string, cached bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if cached {
		s.m[jobName].existsCached++
	} else {
		s.m[jobName].exists++
	}

}

func (s *statistics) Error(jobName string, artistTitle string, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[jobName].errors++
}

func (s *statistics) Print() {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for k, v := range s.m {
		notFoundTotal := v.notFound + v.notFoundCached
		notFoundPercent := v.notFoundCached * 100 / max(notFoundTotal, 1)
		existsTotal := v.exists + v.existsCached
		existsPercent := v.existsCached * 100 / max(existsTotal, 1)
		total := v.added + notFoundTotal + existsTotal + v.errors

		glog.Infof("[%15.15s] Stats: A %4d, N %5d (NC %3d%%), E %5d (EC %3d%%), Error %3d, total: %5d.", k, v.added, notFoundTotal, notFoundPercent, existsTotal, existsPercent, v.errors, total)
	}
}

func max(a int64, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}
