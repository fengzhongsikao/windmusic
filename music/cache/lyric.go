package cache

import (
	"strings"
	"sync"
	"time"

	models "windmusic/internal/music"
)

const LyricCacheTTL = 30 * time.Minute

type lyricCacheEntry struct {
	lyric    models.LyricInfo
	cachedAt time.Time
}

type LyricCache struct {
	mu      sync.RWMutex
	entries map[string]lyricCacheEntry
}

func NewLyricCache() *LyricCache {
	return &LyricCache{
		entries: map[string]lyricCacheEntry{},
	}
}

func (c *LyricCache) Get(metaJSON string) (*models.LyricInfo, bool) {
	key := strings.TrimSpace(metaJSON)
	if key == "" {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Since(entry.cachedAt) >= LyricCacheTTL {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}
	lyric := entry.lyric
	return &lyric, true
}

func (c *LyricCache) Set(metaJSON string, lyric *models.LyricInfo) {
	key := strings.TrimSpace(metaJSON)
	if key == "" || lyric == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = lyricCacheEntry{
		lyric:    *lyric,
		cachedAt: time.Now(),
	}
}
