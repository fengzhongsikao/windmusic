package sourcemgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"windmusic/internal/lxruntime"
	"windmusic/internal/music"
	"windmusic/internal/musicsearch"
	"windmusic/internal/sourcestore"
)

type Manager struct {
	store    *sourcestore.Store
	mu       sync.RWMutex
	runtimes map[string]*lxruntime.Runtime
}

func NewManager(rootDir string) (*Manager, error) {
	store, err := sourcestore.NewStore(rootDir)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		store:    store,
		runtimes: map[string]*lxruntime.Runtime{},
	}
	manager.reloadEnabledSources()
	return manager, nil
}

func DefaultRootDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "windmusic"), nil
}

func (m *Manager) ListSources() []music.SourceInfo {
	return m.store.List()
}

func (m *Manager) ImportSource(sourcePath string) (music.SourceInfo, error) {
	script, err := os.ReadFile(sourcePath)
	if err != nil {
		return music.SourceInfo{}, err
	}

	meta := lxruntime.ParseScriptMeta(string(script))
	info := music.SourceInfo{
		Name:        meta.Name,
		Description: meta.Description,
		Version:     meta.Version,
		Author:      meta.Author,
		Homepage:    meta.Homepage,
	}
	if info.Name == "" {
		info.Name = filepath.Base(sourcePath)
	}

	rt := lxruntime.NewRuntime(string(script))
	initResult, err := rt.Init(30 * time.Second)
	if err != nil {
		return music.SourceInfo{}, fmt.Errorf("source init failed: %w", err)
	}

	platforms := toPlatformInfos(initResult.Platforms)
	record, err := m.store.ImportScript(sourcePath, info, sourcestore.ToStoredPlatforms(platforms))
	if err != nil {
		return music.SourceInfo{}, err
	}

	m.mu.Lock()
	m.runtimes[record.ID] = rt
	m.mu.Unlock()
	return record, nil
}

func (m *Manager) EnableSource(id string) (music.SourceInfo, error) {
	_, err := m.store.SetEnabled(id, true)
	if err != nil {
		return music.SourceInfo{}, err
	}
	if err := m.loadSource(id); err != nil {
		_, _ = m.store.SetEnabled(id, false)
		_ = m.store.UpdateStatus(id, "error", err.Error(), nil)
		return music.SourceInfo{}, err
	}
	updated, _ := m.store.Get(id)
	return updated, nil
}

func (m *Manager) DisableSource(id string) (music.SourceInfo, error) {
	info, err := m.store.SetEnabled(id, false)
	if err != nil {
		return music.SourceInfo{}, err
	}
	m.mu.Lock()
	delete(m.runtimes, id)
	m.mu.Unlock()
	return info, nil
}

func (m *Manager) DeleteSource(id string) error {
	m.mu.Lock()
	delete(m.runtimes, id)
	m.mu.Unlock()
	return m.store.Delete(id)
}

func (m *Manager) Search(sourceID, platform, keyword string, page int) (*music.SearchResult, error) {
	if err := m.ensureRuntime(sourceID); err != nil {
		return nil, err
	}
	if platform == "" {
		platform = m.defaultPlatform(sourceID)
	}
	if platform == "" {
		return nil, fmt.Errorf("no available platform for source")
	}
	return musicsearch.Search(platform, keyword, page, 20)
}

func (m *Manager) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	rt, musicInfo, err := m.prepareRequest(sourceID, platform, metaJSON)
	if err != nil {
		return "", err
	}
	if quality == "" {
		quality = m.defaultQuality(sourceID, platform)
	}
	return rt.GetMusicURL(platform, musicInfo, quality)
}

func (m *Manager) GetLyric(sourceID, platform, metaJSON string) (*music.LyricInfo, error) {
	rt, musicInfo, err := m.prepareRequest(sourceID, platform, metaJSON)
	if err != nil {
		return nil, err
	}
	return rt.GetLyric(platform, musicInfo)
}

func (m *Manager) GetPic(sourceID, platform, metaJSON string) (string, error) {
	rt, musicInfo, err := m.prepareRequest(sourceID, platform, metaJSON)
	if err != nil {
		return "", err
	}
	return rt.GetPic(platform, musicInfo)
}

func (m *Manager) reloadEnabledSources() {
	for _, source := range m.store.List() {
		if !source.Enabled {
			continue
		}
		_ = m.loadSource(source.ID)
	}
}

func (m *Manager) loadSource(id string) error {
	script, err := m.store.ReadScript(m.mustGetFilename(id))
	if err != nil {
		return err
	}

	rt := lxruntime.NewRuntime(script)
	initResult, err := rt.Init(30 * time.Second)
	if err != nil {
		_ = m.store.UpdateStatus(id, "error", err.Error(), nil)
		return err
	}

	platforms := sourcestore.ToStoredPlatforms(toPlatformInfos(initResult.Platforms))
	_ = m.store.UpdateStatus(id, "ready", "", platforms)

	m.mu.Lock()
	m.runtimes[id] = rt
	m.mu.Unlock()
	return nil
}

func (m *Manager) ensureRuntime(sourceID string) error {
	m.mu.RLock()
	_, ok := m.runtimes[sourceID]
	m.mu.RUnlock()
	if ok {
		return nil
	}
	info, found := m.store.Get(sourceID)
	if !found {
		return fmt.Errorf("source not found")
	}
	if !info.Enabled {
		return fmt.Errorf("source is disabled")
	}
	return m.loadSource(sourceID)
}

func (m *Manager) prepareRequest(sourceID, platform, metaJSON string) (*lxruntime.Runtime, map[string]interface{}, error) {
	if err := m.ensureRuntime(sourceID); err != nil {
		return nil, nil, err
	}

	m.mu.RLock()
	rt := m.runtimes[sourceID]
	m.mu.RUnlock()
	if rt == nil {
		return nil, nil, fmt.Errorf("source runtime unavailable")
	}

	musicInfo := map[string]interface{}{}
	if metaJSON != "" {
		if err := json.Unmarshal([]byte(metaJSON), &musicInfo); err != nil {
			return nil, nil, fmt.Errorf("invalid music metadata: %w", err)
		}
	}
	if platform == "" {
		platform = m.defaultPlatform(sourceID)
	}
	return rt, musicInfo, nil
}

func (m *Manager) defaultPlatform(sourceID string) string {
	info, ok := m.store.Get(sourceID)
	if !ok || len(info.Platforms) == 0 {
		return "wy"
	}
	return info.Platforms[0].Key
}

func (m *Manager) defaultQuality(sourceID, platform string) string {
	info, ok := m.store.Get(sourceID)
	if !ok {
		return "128k"
	}
	for _, item := range info.Platforms {
		if item.Key != platform {
			continue
		}
		if len(item.Qualities) > 0 {
			return item.Qualities[0]
		}
	}
	return "128k"
}

func (m *Manager) mustGetFilename(id string) string {
	info, ok := m.store.Get(id)
	if !ok {
		return ""
	}
	return info.Filename
}

func toPlatformInfos(platforms []lxruntime.PlatformMeta) []music.PlatformInfo {
	result := make([]music.PlatformInfo, 0, len(platforms))
	for _, platform := range platforms {
		result = append(result, music.PlatformInfo{
			Key:       platform.Key,
			Name:      platform.Name,
			Actions:   platform.Actions,
			Qualities: platform.Qualities,
		})
	}
	return result
}
