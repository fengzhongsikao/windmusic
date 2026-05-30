package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	models "windmusic/internal/music"
	"windmusic/music/appdata"
)

type FavoritesStore struct {
	path string
	mu   sync.Mutex
}

func (s *FavoritesStore) List() ([]models.FavoriteSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *FavoritesStore) IsFavorite(song models.FavoriteSong) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	favorites, err := s.readLocked()
	if err != nil {
		return false, err
	}
	for _, item := range favorites {
		if sameFavoriteSong(item, song) {
			return true, nil
		}
	}
	return false, nil
}

func (s *FavoritesStore) Add(song models.FavoriteSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	favorites, err := s.readLocked()
	if err != nil {
		return err
	}
	for _, item := range favorites {
		if sameFavoriteSong(item, song) {
			return nil
		}
	}
	favorites = append(favorites, normalizeFavoriteSong(song))
	return s.writeLocked(favorites)
}

func (s *FavoritesStore) Remove(song models.FavoriteSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	favorites, err := s.readLocked()
	if err != nil {
		return err
	}
	next := make([]models.FavoriteSong, 0, len(favorites))
	for _, item := range favorites {
		if !sameFavoriteSong(item, song) {
			next = append(next, item)
		}
	}
	return s.writeLocked(next)
}

func (s *FavoritesStore) ensurePathLocked() (string, error) {
	if s.path != "" {
		return s.path, nil
	}
	rootDir, err := appdata.AppDataRootDir()
	if err != nil {
		return "", err
	}
	s.path = filepath.Join(rootDir, "favorites.json")
	return s.path, nil
}

func (s *FavoritesStore) readLocked() ([]models.FavoriteSong, error) {
	path, err := s.ensurePathLocked()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.FavoriteSong{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []models.FavoriteSong{}, nil
	}
	var favorites []models.FavoriteSong
	if err := json.Unmarshal(data, &favorites); err != nil {
		return nil, err
	}
	return dedupeFavorites(favorites), nil
}

func dedupeFavorites(favorites []models.FavoriteSong) []models.FavoriteSong {
	normalized := make([]models.FavoriteSong, 0, len(favorites))
	seen := make(map[string]struct{}, len(favorites))
	for _, item := range favorites {
		entry := normalizeFavoriteSong(item)
		key := favoriteSongKey(entry)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, entry)
	}
	return normalized
}

func (s *FavoritesStore) writeLocked(favorites []models.FavoriteSong) error {
	path, err := s.ensurePathLocked()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	normalized := dedupeFavorites(favorites)

	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func normalizeFavoriteSong(song models.FavoriteSong) models.FavoriteSong {
	return models.FavoriteSong{
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
}

func favoriteSongKey(song models.FavoriteSong) string {
	entry := normalizeFavoriteSong(song)
	return strings.Join([]string{entry.ID, entry.SourceID, entry.Platform, entry.MetaJSON}, "|")
}

func sameFavoriteSong(a, b models.FavoriteSong) bool {
	left := normalizeFavoriteSong(a)
	right := normalizeFavoriteSong(b)

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

func optionalEqual(a, b string) bool {
	return a == "" || b == "" || a == b
}
