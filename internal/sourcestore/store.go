package sourcestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"windmusic/internal/music"

	"github.com/google/uuid"
)

type storedPlatform struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	Actions   []string `json:"actions"`
	Qualities []string `json:"qualities"`
}

type storedSource struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Author      string           `json:"author"`
	Homepage    string           `json:"homepage"`
	Filename    string           `json:"filename"`
	Enabled     bool             `json:"enabled"`
	Platforms   []storedPlatform `json:"platforms"`
	ImportedAt  time.Time        `json:"importedAt"`
	Status      string           `json:"status"`
	Error       string           `json:"error,omitempty"`
}

type configFile struct {
	Sources []storedSource `json:"sources"`
}

type Store struct {
	rootDir    string
	sourcesDir string
	configPath string
	mu         sync.Mutex
	config     configFile
}

func NewStore(rootDir string) (*Store, error) {
	sourcesDir := filepath.Join(rootDir, "sources")
	configPath := filepath.Join(rootDir, "sources.json")

	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		return nil, err
	}

	store := &Store{
		rootDir:    rootDir,
		sourcesDir: sourcesDir,
		configPath: configPath,
	}

	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) RootDir() string {
	return s.rootDir
}

func (s *Store) SourcesDir() string {
	return s.sourcesDir
}

func (s *Store) List() []music.SourceInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]music.SourceInfo, 0, len(s.config.Sources))
	for _, item := range s.config.Sources {
		result = append(result, toSourceInfo(item))
	}
	return result
}

func (s *Store) Get(id string) (music.SourceInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.config.Sources {
		if item.ID == id {
			return toSourceInfo(item), true
		}
	}
	return music.SourceInfo{}, false
}

func (s *Store) ScriptPath(filename string) string {
	return filepath.Join(s.sourcesDir, filename)
}

func (s *Store) ReadScript(filename string) (string, error) {
	data, err := os.ReadFile(s.ScriptPath(filename))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Store) ImportScript(sourcePath string, meta music.SourceInfo, platforms []storedPlatform) (music.SourceInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	filename := id + ".js"
	targetPath := filepath.Join(s.sourcesDir, filename)

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return music.SourceInfo{}, err
	}
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return music.SourceInfo{}, err
	}

	record := storedSource{
		ID:          id,
		Name:        meta.Name,
		Description: meta.Description,
		Version:     meta.Version,
		Author:      meta.Author,
		Homepage:    meta.Homepage,
		Filename:    filename,
		Enabled:     true,
		Platforms:   platforms,
		ImportedAt:  time.Now(),
		Status:      "ready",
	}
	s.config.Sources = append(s.config.Sources, record)
	if err := s.saveLocked(); err != nil {
		return music.SourceInfo{}, err
	}
	return toSourceInfo(record), nil
}

func (s *Store) SetEnabled(id string, enabled bool) (music.SourceInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, item := range s.config.Sources {
		if item.ID != id {
			continue
		}
		s.config.Sources[i].Enabled = enabled
		if enabled {
			s.config.Sources[i].Status = "ready"
			s.config.Sources[i].Error = ""
		}
		if err := s.saveLocked(); err != nil {
			return music.SourceInfo{}, err
		}
		return toSourceInfo(s.config.Sources[i]), nil
	}
	return music.SourceInfo{}, fmt.Errorf("source not found")
}

func (s *Store) UpdateStatus(id, status, errMsg string, platforms []storedPlatform) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, item := range s.config.Sources {
		if item.ID != id {
			continue
		}
		s.config.Sources[i].Status = status
		s.config.Sources[i].Error = errMsg
		if len(platforms) > 0 {
			s.config.Sources[i].Platforms = platforms
		}
		return s.saveLocked()
	}
	return fmt.Errorf("source not found")
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, item := range s.config.Sources {
		if item.ID != id {
			continue
		}
		_ = os.Remove(s.ScriptPath(item.Filename))
		s.config.Sources = append(s.config.Sources[:i], s.config.Sources[i+1:]...)
		return s.saveLocked()
	}
	return fmt.Errorf("source not found")
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.config = configFile{Sources: []storedSource{}}
			return s.saveLocked()
		}
		return err
	}
	return json.Unmarshal(data, &s.config)
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0o644)
}

func toSourceInfo(item storedSource) music.SourceInfo {
	platforms := make([]music.PlatformInfo, 0, len(item.Platforms))
	for _, platform := range item.Platforms {
		platforms = append(platforms, music.PlatformInfo{
			Key:       platform.Key,
			Name:      platform.Name,
			Actions:   platform.Actions,
			Qualities: platform.Qualities,
		})
	}
	return music.SourceInfo{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Version:     item.Version,
		Author:      item.Author,
		Homepage:    item.Homepage,
		Filename:    item.Filename,
		Enabled:     item.Enabled,
		Platforms:   platforms,
		ImportedAt:  item.ImportedAt,
		Status:      item.Status,
		Error:       item.Error,
	}
}

func ToStoredPlatforms(platforms []music.PlatformInfo) []storedPlatform {
	result := make([]storedPlatform, 0, len(platforms))
	for _, platform := range platforms {
		result = append(result, storedPlatform{
			Key:       platform.Key,
			Name:      platform.Name,
			Actions:   platform.Actions,
			Qualities: platform.Qualities,
		})
	}
	return result
}
