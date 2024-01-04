package main

import (
	"bytes"
	"fmt"
	"sync"
)

type aggregatedStatus struct {
	added    int64
	notFound int64
	exists   int64
	errors   int64
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
func (s *statistics) NotFound(jobName string, artistTitle string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[jobName].notFound++
}

// Song was found by the saver but it already exists
func (s *statistics) Exists(jobName string, artistTitle string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[jobName].exists++
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
		total := v.added + v.notFound + v.exists + v.errors

		buf.WriteString(fmt.Sprintf("[%15.15s] A %4d, N %5d, E %5d, Err %3d, total: %5d.\n", k, v.added, v.notFound, v.exists, v.errors, total))
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
		if v.added+v.exists == 0 {
			buf.WriteString(k)
			buf.WriteString(": no songs were added or existed before.\n")
		} else if v.errors*100/(v.added+v.exists) > 30 {
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
