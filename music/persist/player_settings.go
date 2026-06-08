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

type PlayerSettingsStore struct {
	path string
	mu   sync.Mutex
}

func DefaultPlayerSettings() models.PlayerSettings {
	return models.PlayerSettings{
		Volume:         30,
		Muted:          false,
		RepeatMode:     "off",
		Shuffled:       false,
		WaveformSpread: "center-out",
	}
}

func normalizeWaveformSpread(value string) string {
	switch strings.TrimSpace(value) {
	case "edges-in", "right-left":
		return strings.TrimSpace(value)
	default:
		return "center-out"
	}
}

func normalizePlayerSettings(settings models.PlayerSettings) models.PlayerSettings {
	volume := settings.Volume
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	repeatMode := strings.TrimSpace(settings.RepeatMode)
	if repeatMode != "all" && repeatMode != "one" {
		repeatMode = "off"
	}
	return models.PlayerSettings{
		Volume:         volume,
		Muted:          settings.Muted,
		RepeatMode:     repeatMode,
		Shuffled:       settings.Shuffled,
		WaveformSpread: normalizeWaveformSpread(settings.WaveformSpread),
	}
}

func (s *PlayerSettingsStore) settingsPath() (string, error) {
	if s.path != "" {
		return s.path, nil
	}
	root, err := appdata.AppDataRootDir()
	if err != nil {
		return "", err
	}
	s.path = filepath.Join(root, "player-settings.json")
	return s.path, nil
}

func (s *PlayerSettingsStore) Get() (models.PlayerSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := s.settingsPath()
	if err != nil {
		return models.PlayerSettings{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultPlayerSettings(), nil
		}
		return models.PlayerSettings{}, err
	}
	if len(data) == 0 {
		return DefaultPlayerSettings(), nil
	}
	var settings models.PlayerSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return models.PlayerSettings{}, err
	}
	return normalizePlayerSettings(settings), nil
}

func (s *PlayerSettingsStore) Save(settings models.PlayerSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := s.settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	normalized := normalizePlayerSettings(settings)
	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func (s *PlayerSettingsStore) saveLocked(settings models.PlayerSettings) error {
	path, err := s.settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}
