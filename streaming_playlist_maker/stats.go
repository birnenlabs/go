package main

import (
	"bytes"
	"fmt"
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

func (s *statistics) String() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var buf bytes.Buffer
	for k, v := range s.m {
		notFoundTotal := v.notFound + v.notFoundCached
		notFoundPercent := v.notFoundCached * 100 / max(notFoundTotal, 1)
		existsTotal := v.exists + v.existsCached
		existsPercent := v.existsCached * 100 / max(existsTotal, 1)
		total := v.added + notFoundTotal + existsTotal + v.errors

		buf.WriteString(fmt.Sprintf("[%15.15s] A %4d, N %5d (NC %3d%%), E %5d (EC %3d%%), Error %3d, total: %5d.\n", k, v.added, notFoundTotal, notFoundPercent, existsTotal, existsPercent, v.errors, total))
	}
	return buf.String()
}

func (s *statistics) FindIssues() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var buf bytes.Buffer
	if len(s.m) == 0 {
		buf.WriteString("Zero sources were tracked.")
	}

	for k, v := range s.m {
		if v.added+v.exists+v.existsCached == 0 {
			buf.WriteString(k)
			buf.WriteString(": no songs were added or existed before.\n")
		} else if v.errors*100/(v.added+v.exists+v.existsCached) > 30 {
			buf.WriteString(k)
			buf.WriteString(": more than 30% of errors.\n")
		}
	}
	return buf.String()

}

func max(a int64, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}
