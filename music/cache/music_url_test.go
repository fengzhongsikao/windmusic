package cache

import (
	"testing"
	"time"
)

func TestMusicURLCacheGetSet(t *testing.T) {
	c := NewMusicURLCache()
	if _, ok := c.Get("src", "netease", "", `{"id":"1"}`); ok {
		t.Fatal("expected cache miss")
	}

	c.Set("src", "netease", "", `{"id":"1"}`, "https://example.com/track.mp3")
	url, ok := c.Get("src", "netease", "", `{"id":"1"}`)
	if !ok || url != "https://example.com/track.mp3" {
		t.Fatalf("cache hit = %v url = %q", ok, url)
	}
}

func TestMusicURLCacheTTL(t *testing.T) {
	c := NewMusicURLCache()
	key := MusicURLCacheKey("src", "netease", "", `{"id":"1"}`)
	c.mu.Lock()
	c.entries[key] = musicURLCacheEntry{
		url:      "https://example.com/old.mp3",
		cachedAt: time.Now().Add(-MusicURLCacheTTL),
	}
	c.mu.Unlock()

	if _, ok := c.Get("src", "netease", "", `{"id":"1"}`); ok {
		t.Fatal("expected expired entry to miss")
	}
}
