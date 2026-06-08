package local

import (
	"os"
	"testing"

	models "windmusic/internal/music"
)

func TestSaveScanCacheUpsertAndPrune(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WINDMUSIC_APPDATA", root)

	store := &LocalLibraryStore{}
	if err := store.ensureLibraryDB(); err != nil {
		t.Fatalf("ensureLibraryDB: %v", err)
	}

	cache := newScanCache()
	cache.Entries["/music/a.mp3"] = localScanCacheEntry{
		ModTimeUnix: 1,
		Song:        models.LocalSong{ID: "/music/a.mp3", Title: "A", FilePath: "/music/a.mp3"},
	}
	cache.Entries["/music/b.mp3"] = localScanCacheEntry{
		ModTimeUnix: 2,
		Song:        models.LocalSong{ID: "/music/b.mp3", Title: "B", FilePath: "/music/b.mp3"},
	}
	if err := store.db.saveScanCache(cache); err != nil {
		t.Fatalf("saveScanCache: %v", err)
	}

	cache.Entries["/music/a.mp3"] = localScanCacheEntry{
		ModTimeUnix: 10,
		Song:        models.LocalSong{ID: "/music/a.mp3", Title: "A updated", FilePath: "/music/a.mp3"},
	}
	delete(cache.Entries, "/music/b.mp3")
	if err := store.db.saveScanCache(cache); err != nil {
		t.Fatalf("saveScanCache update: %v", err)
	}

	loaded, err := store.db.loadScanCache()
	if err != nil {
		t.Fatalf("loadScanCache: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry after prune, got %d", len(loaded.Entries))
	}
	entry := loaded.Entries["/music/a.mp3"]
	if entry.ModTimeUnix != 10 || entry.Song.Title != "A updated" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestEnsureLibraryDBOpensEmpty(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WINDMUSIC_APPDATA", root)

	store := &LocalLibraryStore{}
	if err := store.ensureLibraryDB(); err != nil {
		t.Fatalf("ensureLibraryDB: %v", err)
	}
	if store.db == nil {
		t.Fatal("expected db")
	}
	if store.coverFiles == nil {
		t.Fatal("expected cover file store")
	}

	cache, err := store.db.loadScanCache()
	if err != nil {
		t.Fatalf("loadScanCache: %v", err)
	}
	if len(cache.Entries) != 0 {
		t.Fatalf("expected empty cache, got %d entries", len(cache.Entries))
	}

	extras, err := store.db.loadExtras()
	if err != nil {
		t.Fatalf("loadExtras: %v", err)
	}
	if len(extras.Entries) != 0 {
		t.Fatalf("expected empty extras, got %d entries", len(extras.Entries))
	}

	if _, err := os.Stat(root + "/local-library.db"); err != nil {
		t.Fatalf("expected sqlite db file: %v", err)
	}
}
