package local

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudioServerStreamURLAndServe(t *testing.T) {
	root := t.TempDir()
	musicDir := filepath.Join(root, "music")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		t.Fatal(err)
	}

	audioPath := filepath.Join(musicDir, "track.mp3")
	if err := os.WriteFile(audioPath, []byte("fake-mp3-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	foldersPath := filepath.Join(root, "local-folders.json")
	foldersPayload, err := json.Marshal(map[string][]string{"paths": {musicDir}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foldersPath, foldersPayload, 0o644); err != nil {
		t.Fatal(err)
	}

	origRoot := os.Getenv("WINDMUSIC_APPDATA")
	t.Setenv("WINDMUSIC_APPDATA", root)
	defer os.Setenv("WINDMUSIC_APPDATA", origRoot)

	store := &LocalLibraryStore{}
	server := NewAudioServer(store)

	streamURL, err := server.StreamURL(audioPath)
	if err != nil {
		t.Fatalf("StreamURL: %v", err)
	}
	if !strings.HasPrefix(streamURL, "http://127.0.0.1:") || !strings.Contains(streamURL, localAudioRoutePrefix) {
		t.Fatalf("unexpected stream URL: %q", streamURL)
	}

	resp, err := http.Get(streamURL)
	if err != nil {
		t.Fatalf("GET stream URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "fake-mp3-bytes" {
		t.Fatalf("body = %q", body)
	}
	if resp.Header.Get("Accept-Ranges") == "" {
		t.Fatal("expected Accept-Ranges header from ServeContent")
	}
}

func TestAudioServerRejectsOutsideLibrary(t *testing.T) {
	root := t.TempDir()
	musicDir := filepath.Join(root, "music")
	otherDir := filepath.Join(root, "other")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatal(err)
	}

	inLib := filepath.Join(musicDir, "in.mp3")
	outside := filepath.Join(otherDir, "out.mp3")
	for _, path := range []string{inLib, outside} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	foldersPath := filepath.Join(root, "local-folders.json")
	foldersPayload, err := json.Marshal(map[string][]string{"paths": {musicDir}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foldersPath, foldersPayload, 0o644); err != nil {
		t.Fatal(err)
	}

	origRoot := os.Getenv("WINDMUSIC_APPDATA")
	t.Setenv("WINDMUSIC_APPDATA", root)
	defer os.Setenv("WINDMUSIC_APPDATA", origRoot)

	store := &LocalLibraryStore{}
	server := NewAudioServer(store)

	if _, err := server.StreamURL(outside); err == nil {
		t.Fatal("expected error for file outside library")
	}

	inURL, err := server.StreamURL(inLib)
	if err != nil {
		t.Fatalf("StreamURL in library: %v", err)
	}

	resp, err := http.Get(inURL + "x")
	if err != nil {
		t.Fatalf("GET tampered URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("tampered token status = %d, want 404", resp.StatusCode)
	}
}
