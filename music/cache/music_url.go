package cache

import (
	"strings"
	"sync"
	"time"
)

const MusicURLCacheTTL = 20 * time.Minute

type musicURLCacheEntry struct {
	url      string
	cachedAt time.Time
}

type MusicURLCache struct {
	mu      sync.RWMutex
	entries map[string]musicURLCacheEntry
}

func NewMusicURLCache() *MusicURLCache {
	return &MusicURLCache{
		entries: map[string]musicURLCacheEntry{},
	}
}

func MusicURLCacheKey(sourceID, platform, quality, metaJSON string) string {
	return strings.TrimSpace(sourceID) + "\x00" +
		strings.TrimSpace(platform) + "\x00" +
		strings.TrimSpace(quality) + "\x00" +
		strings.TrimSpace(metaJSON)
}

func (c *MusicURLCache) Get(sourceID, platform, quality, metaJSON string) (string, bool) {
	key := MusicURLCacheKey(sourceID, platform, quality, metaJSON)
	if key == "\x00\x00\x00" {
		return "", false
	}

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Since(entry.cachedAt) >= MusicURLCacheTTL {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return "", false
	}
	return entry.url, true
}

func (c *MusicURLCache) Set(sourceID, platform, quality, metaJSON, url string) {
	key := MusicURLCacheKey(sourceID, platform, quality, metaJSON)
	url = strings.TrimSpace(url)
	if key == "\x00\x00\x00" || url == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = musicURLCacheEntry{
		url:      url,
		cachedAt: time.Now(),
	}
}
