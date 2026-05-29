package music

import (
	"strings"
	"sync"
	"time"

	models "windmusic/internal/music"
)

const DiscoverRecommendCacheTTL = 5 * time.Minute

type discoverCacheEntry struct {
	songs    []models.SongItem
	cachedAt time.Time
}

type DiscoverCache struct {
	mu      sync.Mutex
	entries map[string]discoverCacheEntry
}

func NewDiscoverCache() *DiscoverCache {
	return &DiscoverCache{
		entries: map[string]discoverCacheEntry{},
	}
}

func (c *DiscoverCache) Get(tabID string) (models.DiscoverRecommendCache, error) {
	tabID = strings.TrimSpace(tabID)
	if tabID == "" {
		return models.DiscoverRecommendCache{Hit: false, Songs: []models.SongItem{}}, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[tabID]
	if !ok {
		return models.DiscoverRecommendCache{Hit: false, Songs: []models.SongItem{}}, nil
	}
	if time.Since(entry.cachedAt) >= DiscoverRecommendCacheTTL {
		delete(c.entries, tabID)
		return models.DiscoverRecommendCache{Hit: false, Songs: []models.SongItem{}}, nil
	}
	songs := entry.songs
	if songs == nil {
		songs = []models.SongItem{}
	}
	return models.DiscoverRecommendCache{Hit: true, Songs: songs}, nil
}

func (c *DiscoverCache) Set(tabID string, songs []models.SongItem) error {
	tabID = strings.TrimSpace(tabID)
	if tabID == "" {
		return nil
	}
	if songs == nil {
		songs = []models.SongItem{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[tabID] = discoverCacheEntry{
		songs:    songs,
		cachedAt: time.Now(),
	}
	return nil
}
