package persist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlaylistsStoreCreateDoesNotDeadlock(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WINDMUSIC_APPDATA", root)

	store := &PlaylistsStore{}
	created, err := store.Create("测试歌单")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.Name != "测试歌单" {
		t.Fatalf("Create() = %+v, want non-empty id and trimmed name", created)
	}

	path := filepath.Join(root, "playlists.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("playlists.json not written: %v", err)
	}
}
