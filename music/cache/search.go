package cache

import (
	"strings"
	"sync"
	"time"

	models "windmusic/internal/music"
)

const SearchCacheTTL = 5 * time.Minute

type searchCacheEntry struct {
	items    []models.SongItem
	source   string
	cachedAt time.Time
}

type SearchCache struct {
	mu      sync.RWMutex
	entries map[string]searchCacheEntry
}

func NewSearchCache() *SearchCache {
	return &SearchCache{
		entries: map[string]searchCacheEntry{},
	}
}

func SearchCacheKey(baseURL, platform, keyword string) string {
	return strings.ToLower(strings.TrimSpace(baseURL)) + "\x00" +
		strings.ToLower(strings.TrimSpace(platform)) + "\x00" +
		strings.TrimSpace(keyword)
}

func (c *SearchCache) Get(baseURL, platform, keyword string, page, limit int) (*models.SearchResult, bool) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	key := SearchCacheKey(baseURL, platform, keyword)
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Since(entry.cachedAt) >= SearchCacheTTL {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}

	items := entry.items
	total := len(items)
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	list := make([]models.SongItem, end-start)
	copy(list, items[start:end])
	return &models.SearchResult{
		List:   list,
		Total:  total,
		Page:   page,
		Limit:  limit,
		Source: entry.source,
	}, true
}

func (c *SearchCache) Set(baseURL, platform, keyword string, items []models.SongItem, source string) {
	key := SearchCacheKey(baseURL, platform, keyword)
	if items == nil {
		items = []models.SongItem{}
	}
	stored := make([]models.SongItem, len(items))
	copy(stored, items)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = searchCacheEntry{
		items:    stored,
		source:   source,
		cachedAt: time.Now(),
	}
}
