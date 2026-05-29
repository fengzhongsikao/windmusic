package music

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	models "windmusic/internal/music"
)

var (
	ErrPlaylistNameEmpty  = errors.New("playlist name is required")
	ErrPlaylistNameExists = errors.New("playlist name already exists")
	ErrPlaylistNotFound   = errors.New("playlist not found")
)

type PlaylistsStore struct {
	path string
	mu   sync.Mutex
}

func (s *PlaylistsStore) List() ([]models.UserPlaylist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *PlaylistsStore) Create(name string) (models.UserPlaylist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return models.UserPlaylist{}, ErrPlaylistNameEmpty
	}

	playlists, err := s.readLocked()
	if err != nil {
		return models.UserPlaylist{}, err
	}
	for _, item := range playlists {
		if strings.EqualFold(strings.TrimSpace(item.Name), trimmed) {
			return models.UserPlaylist{}, ErrPlaylistNameExists
		}
	}

	playlist := models.UserPlaylist{
		ID:        newPlaylistID(),
		Name:      trimmed,
		CreatedAt: time.Now().UTC(),
		Songs:     []models.FavoriteSong{},
	}
	playlists = append(playlists, playlist)
	if err := s.writeLocked(playlists); err != nil {
		return models.UserPlaylist{}, err
	}
	return playlist, nil
}

func (s *PlaylistsStore) Get(id string) (models.UserPlaylist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	playlistID := strings.TrimSpace(id)
	if playlistID == "" {
		return models.UserPlaylist{}, ErrPlaylistNotFound
	}

	playlists, err := s.readLocked()
	if err != nil {
		return models.UserPlaylist{}, err
	}
	for _, item := range playlists {
		if item.ID == playlistID {
			return item, nil
		}
	}
	return models.UserPlaylist{}, ErrPlaylistNotFound
}

func (s *PlaylistsStore) AddSong(playlistID string, song models.FavoriteSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := strings.TrimSpace(playlistID)
	if id == "" {
		return ErrPlaylistNotFound
	}

	playlists, err := s.readLocked()
	if err != nil {
		return err
	}

	entry := normalizeFavoriteSong(song)
	if entry.Title == "" && entry.ID == "" {
		return errors.New("invalid song")
	}

	found := false
	for i := range playlists {
		if playlists[i].ID != id {
			continue
		}
		found = true
		for _, existing := range playlists[i].Songs {
			if sameFavoriteSong(existing, entry) {
				return nil
			}
		}
		playlists[i].Songs = append(playlists[i].Songs, entry)
		break
	}
	if !found {
		return ErrPlaylistNotFound
	}
	return s.writeLocked(playlists)
}

func (s *PlaylistsStore) RemoveSong(playlistID string, song models.FavoriteSong) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := strings.TrimSpace(playlistID)
	if id == "" {
		return ErrPlaylistNotFound
	}

	playlists, err := s.readLocked()
	if err != nil {
		return err
	}

	entry := normalizeFavoriteSong(song)
	found := false
	for i := range playlists {
		if playlists[i].ID != id {
			continue
		}
		found = true
		next := make([]models.FavoriteSong, 0, len(playlists[i].Songs))
		for _, item := range playlists[i].Songs {
			if !sameFavoriteSong(item, entry) {
				next = append(next, item)
			}
		}
		playlists[i].Songs = next
		break
	}
	if !found {
		return ErrPlaylistNotFound
	}
	return s.writeLocked(playlists)
}

func (s *PlaylistsStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	playlistID := strings.TrimSpace(id)
	if playlistID == "" {
		return ErrPlaylistNotFound
	}

	playlists, err := s.readLocked()
	if err != nil {
		return err
	}

	next := make([]models.UserPlaylist, 0, len(playlists))
	found := false
	for _, item := range playlists {
		if item.ID == playlistID {
			found = true
			continue
		}
		next = append(next, item)
	}
	if !found {
		return ErrPlaylistNotFound
	}
	return s.writeLocked(next)
}

func (s *PlaylistsStore) ensurePathLocked() (string, error) {
	if s.path != "" {
		return s.path, nil
	}
	rootDir, err := AppDataRootDir()
	if err != nil {
		return "", err
	}
	s.path = filepath.Join(rootDir, "playlists.json")
	return s.path, nil
}

func (s *PlaylistsStore) readLocked() ([]models.UserPlaylist, error) {
	path, err := s.ensurePathLocked()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.UserPlaylist{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []models.UserPlaylist{}, nil
	}
	var playlists []models.UserPlaylist
	if err := json.Unmarshal(data, &playlists); err != nil {
		return nil, err
	}
	for i := range playlists {
		if playlists[i].Songs == nil {
			playlists[i].Songs = []models.FavoriteSong{}
		}
	}
	return playlists, nil
}

func (s *PlaylistsStore) writeLocked(playlists []models.UserPlaylist) error {
	path, err := s.ensurePathLocked()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	for i := range playlists {
		if playlists[i].Songs == nil {
			playlists[i].Songs = []models.FavoriteSong{}
		}
	}
	data, err := json.MarshalIndent(playlists, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func newPlaylistID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339Nano), ":", "")))
	}
	return hex.EncodeToString(b)
}
