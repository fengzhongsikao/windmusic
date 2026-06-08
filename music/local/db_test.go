package local

import (
	"os"
	"testing"
)

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
