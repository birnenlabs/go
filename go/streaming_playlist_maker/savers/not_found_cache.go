package savers

import (
	"sync"
)

type nfCache struct {
	cache     map[string]*Status
	cacheLock sync.RWMutex
}

func newCache() *nfCache {
	c := make(map[string]*Status)

	return &nfCache{
		cache: c,
	}
}

func (n *nfCache) IsNotFound(artistTitle string) *Status {
	n.cacheLock.RLock()
	defer n.cacheLock.RUnlock()

	return n.cache[artistTitle]
}

func (n *nfCache) AddNotFound(artistTitle string, status *Status) {
	n.cacheLock.Lock()
	defer n.cacheLock.Unlock()

	n.cache[artistTitle] = status
}
