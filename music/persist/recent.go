package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	models "windmusic/internal/music"
	"windmusic/music/appdata"
)

const recentPlayMaxItems = 200

type RecentStore struct {
	path string
	mu   sync.Mutex
}

func (s *RecentStore) List() ([]models.RecentSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *RecentStore) Record(song models.RecentSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	recent, err := s.readLocked()
	if err != nil {
		return err
	}

	entry := normalizeRecentSong(song)
	entry.PlayedAt = time.Now().UTC()

	next := make([]models.RecentSong, 0, len(recent)+1)
	next = append(next, entry)
	for _, item := range recent {
		if !sameRecentSong(item, entry) {
			next = append(next, item)
		}
	}
	if len(next) > recentPlayMaxItems {
		next = next[:recentPlayMaxItems]
	}
	return s.writeLocked(next)
}

func (s *RecentStore) Remove(song models.RecentSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	recent, err := s.readLocked()
	if err != nil {
		return err
	}
	next := make([]models.RecentSong, 0, len(recent))
	for _, item := range recent {
		if !sameRecentSong(item, song) {
			next = append(next, item)
		}
	}
	return s.writeLocked(next)
}

func (s *RecentStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeLocked([]models.RecentSong{})
}

func (s *RecentStore) ensurePathLocked() (string, error) {
	if s.path != "" {
		return s.path, nil
	}
	rootDir, err := appdata.AppDataRootDir()
	if err != nil {
		return "", err
	}
	s.path = filepath.Join(rootDir, "recent.json")
	return s.path, nil
}

func (s *RecentStore) readLocked() ([]models.RecentSong, error) {
	path, err := s.ensurePathLocked()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.RecentSong{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []models.RecentSong{}, nil
	}
	var recent []models.RecentSong
	if err := json.Unmarshal(data, &recent); err != nil {
		return nil, err
	}
	return recent, nil
}

func (s *RecentStore) writeLocked(recent []models.RecentSong) error {
	path, err := s.ensurePathLocked()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	normalized := make([]models.RecentSong, 0, len(recent))
	seen := make(map[string]struct{}, len(recent))
	for _, item := range recent {
		entry := normalizeRecentSong(item)
		key := recentSongKey(entry)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, entry)
	}

	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func normalizeRecentSong(song models.RecentSong) models.RecentSong {
	entry := models.RecentSong{
		ID:       strings.TrimSpace(song.ID),
		Title:    strings.TrimSpace(song.Title),
		Artist:   strings.TrimSpace(song.Artist),
		Album:    strings.TrimSpace(song.Album),
		Duration: strings.TrimSpace(song.Duration),
		CoverURL: strings.TrimSpace(song.CoverURL),
		SourceID: strings.TrimSpace(song.SourceID),
		Platform: strings.TrimSpace(song.Platform),
		MetaJSON: strings.TrimSpace(song.MetaJSON),
	}
	if !song.PlayedAt.IsZero() {
		entry.PlayedAt = song.PlayedAt.UTC()
	}
	return entry
}

func recentSongKey(song models.RecentSong) string {
	entry := normalizeRecentSong(song)
	return strings.Join([]string{entry.ID, entry.SourceID, entry.Platform, entry.MetaJSON}, "|")
}

func sameRecentSong(a, b models.RecentSong) bool {
	left := normalizeRecentSong(a)
	right := normalizeRecentSong(b)

	if left.MetaJSON != "" && right.MetaJSON != "" {
		if left.MetaJSON != right.MetaJSON {
			return false
		}
		if !optionalEqual(left.Platform, right.Platform) {
			return false
		}
		if !optionalEqual(left.SourceID, right.SourceID) {
			return false
		}
		return true
	}

	if left.ID != "" && right.ID != "" {
		if left.ID != right.ID {
			return false
		}
		if !optionalEqual(left.Platform, right.Platform) {
			return false
		}
		if !optionalEqual(left.SourceID, right.SourceID) {
			return false
		}
		return true
	}

	return left.Title != "" &&
		right.Title != "" &&
		left.Artist != "" &&
		right.Artist != "" &&
		left.Title == right.Title &&
		left.Artist == right.Artist
}
