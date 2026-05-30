package persist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	models "windmusic/internal/music"
	"windmusic/music/appdata"
)

var httpURLPattern = regexp.MustCompile(`(?i)^https?://`)

type MetingSettingsStore struct {
	path string
	mu   sync.Mutex
}

func DefaultMetingSettings() models.MetingSettings {
	return models.MetingSettings{
		URLs:      []string{},
		ActiveURL: "",
		Platform:  "netease",
	}
}

func NormalizeMetingURL(url string) string {
	return strings.TrimRight(strings.TrimSpace(url), "/")
}

func NormalizeMetingPlatform(raw string) string {
	v := strings.TrimSpace(strings.ToLower(raw))
	switch v {
	case "netease", "wy", "163":
		return "netease"
	default:
		return "netease"
	}
}

func normalizeMetingURLs(urls []string) []string {
	seen := make(map[string]struct{}, len(urls))
	out := make([]string, 0, len(urls))
	for _, item := range urls {
		normalized := NormalizeMetingURL(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeMetingSettings(settings models.MetingSettings) models.MetingSettings {
	urls := normalizeMetingURLs(settings.URLs)
	active := NormalizeMetingURL(settings.ActiveURL)
	if active != "" {
		found := false
		for _, url := range urls {
			if url == active {
				found = true
				break
			}
		}
		if !found {
			active = ""
		}
	}
	if active == "" && len(urls) > 0 {
		active = urls[0]
	}
	return models.MetingSettings{
		URLs:      urls,
		ActiveURL: active,
		Platform:  NormalizeMetingPlatform(settings.Platform),
	}
}

func (s *MetingSettingsStore) settingsPath() (string, error) {
	if s.path != "" {
		return s.path, nil
	}
	root, err := appdata.AppDataRootDir()
	if err != nil {
		return "", err
	}
	s.path = filepath.Join(root, "meting-settings.json")
	return s.path, nil
}

func (s *MetingSettingsStore) readLocked() (models.MetingSettings, error) {
	path, err := s.settingsPath()
	if err != nil {
		return models.MetingSettings{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultMetingSettings(), nil
		}
		return models.MetingSettings{}, err
	}
	if len(data) == 0 {
		return DefaultMetingSettings(), nil
	}
	var settings models.MetingSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return models.MetingSettings{}, err
	}
	return normalizeMetingSettings(settings), nil
}

func (s *MetingSettingsStore) writeLocked(settings models.MetingSettings) error {
	path, err := s.settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	normalized := normalizeMetingSettings(settings)
	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func (s *MetingSettingsStore) Get() (models.MetingSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *MetingSettingsStore) Save(settings models.MetingSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeLocked(settings)
}

func (s *MetingSettingsStore) AddURL(url string) (models.MetingSettings, error) {
	normalized := NormalizeMetingURL(url)
	if normalized == "" {
		return models.MetingSettings{}, fmt.Errorf("请输入 Meting 地址")
	}
	if !httpURLPattern.MatchString(normalized) {
		return models.MetingSettings{}, fmt.Errorf("地址需以 http:// 或 https:// 开头")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.readLocked()
	if err != nil {
		return models.MetingSettings{}, err
	}
	for _, existing := range settings.URLs {
		if existing == normalized {
			return models.MetingSettings{}, fmt.Errorf("该地址已存在")
		}
	}
	settings.URLs = append(settings.URLs, normalized)
	settings = normalizeMetingSettings(settings)
	return settings, s.writeLocked(settings)
}

func (s *MetingSettingsStore) RemoveURL(url string) (models.MetingSettings, error) {
	normalized := NormalizeMetingURL(url)

	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.readLocked()
	if err != nil {
		return models.MetingSettings{}, err
	}
	next := make([]string, 0, len(settings.URLs))
	for _, existing := range settings.URLs {
		if existing != normalized {
			next = append(next, existing)
		}
	}
	settings.URLs = next
	settings = normalizeMetingSettings(settings)
	return settings, s.writeLocked(settings)
}

func (s *MetingSettingsStore) SetActiveURL(url string) (models.MetingSettings, error) {
	normalized := NormalizeMetingURL(url)

	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.readLocked()
	if err != nil {
		return models.MetingSettings{}, err
	}
	settings.ActiveURL = normalized
	settings = normalizeMetingSettings(settings)
	return settings, s.writeLocked(settings)
}

func (s *MetingSettingsStore) SetPlatform(platform string) (models.MetingSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.readLocked()
	if err != nil {
		return models.MetingSettings{}, err
	}
	settings.Platform = NormalizeMetingPlatform(platform)
	return settings, s.writeLocked(settings)
}
