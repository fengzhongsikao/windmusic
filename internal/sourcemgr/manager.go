package sourcemgr

import (
	"encoding/json"
	"fmt"
	"log"
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

func (m *Manager) SourceDisplayName(sourceID string) string {
	info, ok := m.store.Get(sourceID)
	shortID := sourceID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	if !ok {
		return fmt.Sprintf("未知音源(%s)", shortID)
	}
	name := info.Name
	if name == "" {
		name = "未命名音源"
	}
	return fmt.Sprintf("%s(%s)", name, shortID)
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
		return music.SourceInfo{}, fmt.Errorf("音源初始化失败: %w", err)
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
	startedAt := time.Now()
	log.Printf("[后端:音源管理] 开始搜索 source=%s platform=%s page=%d keyword=%q", sourceID, platform, page, keyword)

	if err := m.ensureRuntime(sourceID); err != nil {
		log.Printf("[后端:音源管理] 搜索前检查运行时失败 source=%s err=%v elapsed=%s", sourceID, err, time.Since(startedAt))
		return nil, err
	}

	if platform == "" {
		platform = m.defaultPlatform(sourceID)
		log.Printf("[后端:音源管理] 搜索回退到默认平台 source=%s platform=%s", sourceID, platform)
	}
	if platform == "" {
		log.Printf("[后端:音源管理] 搜索失败 source=%s err=无可用平台 elapsed=%s", sourceID, time.Since(startedAt))
		return nil, fmt.Errorf("当前音源没有可用平台")
	}

	result, err := musicsearch.Search(platform, keyword, page, 20)
	if err != nil {
		log.Printf("[后端:音源管理] 搜索失败 source=%s platform=%s err=%v elapsed=%s", sourceID, platform, err, time.Since(startedAt))
		return nil, err
	}

	log.Printf("[后端:音源管理] 搜索完成 source=%s platform=%s list=%d total=%d elapsed=%s", sourceID, result.Source, len(result.List), result.Total, time.Since(startedAt))
	return result, nil
}

func (m *Manager) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	startedAt := time.Now()
	log.Printf("[后端:音源管理] 开始获取播放地址 source=%s platform=%s quality=%s metaBytes=%d", sourceID, platform, quality, len(metaJSON))

	rt, musicInfo, err := m.prepareRequest(sourceID, platform, metaJSON)
	if err != nil {
		log.Printf("[后端:音源管理] 获取播放地址前准备请求失败 source=%s err=%v elapsed=%s", sourceID, err, time.Since(startedAt))
		return "", err
	}
	if quality == "" {
		quality = m.defaultQuality(sourceID, platform)
		log.Printf("[后端:音源管理] 获取播放地址回退到默认音质 source=%s platform=%s quality=%s", sourceID, platform, quality)
	}

	normalizeMusicInfoAliases(musicInfo)
	log.Printf("[后端:音源管理] 调用运行时获取播放地址 source=%s platform=%s quality=%s musicInfoKeys=%d songmid=%q hash=%q", sourceID, platform, quality, len(musicInfo), musicInfoString(musicInfo, "songmid", "songMid", "id"), musicInfoString(musicInfo, "hash", "FileHash"))
	url, err := rt.GetMusicURL(platform, musicInfo, quality)
	if err != nil {
		log.Printf("[后端:音源管理] 运行时获取播放地址失败 source=%s platform=%s quality=%s err=%v elapsed=%s", sourceID, platform, quality, err, time.Since(startedAt))
		return "", err
	}

	log.Printf("[后端:音源管理] 获取播放地址完成 source=%s platform=%s quality=%s urlBytes=%d elapsed=%s musicUrl=%s", sourceID, platform, quality, len(url), time.Since(startedAt), url)
	return url, nil
}

func (m *Manager) GetLyric(sourceID, platform, metaJSON string) (*music.LyricInfo, error) {
	startedAt := time.Now()
	log.Printf("[后端:音源管理] 开始获取歌词 source=%s platform=%s metaBytes=%d", sourceID, platform, len(metaJSON))

	rt, musicInfo, err := m.prepareRequest(sourceID, platform, metaJSON)
	if err != nil {
		log.Printf("[后端:音源管理] 获取歌词前准备请求失败 source=%s err=%v elapsed=%s", sourceID, err, time.Since(startedAt))
		return nil, err
	}
	platformCandidates := m.platformCandidatesForAction(sourceID, platform, "lyric")
	log.Printf("[后端:音源管理] 调用运行时获取歌词 source=%s platformCandidates=%v song=%q singer=%q songmid=%q hash=%q", sourceID, platformCandidates, musicInfoString(musicInfo, "name", "songName"), musicInfoString(musicInfo, "singer", "singerName"), musicInfoString(musicInfo, "songmid", "songMid", "id"), musicInfoString(musicInfo, "hash", "FileHash"))

	var lastErr error
	for _, candidate := range platformCandidates {
		lyric, callErr := rt.GetLyric(candidate, musicInfo)
		if callErr == nil {
			log.Printf("[后端:音源管理] 获取歌词完成 source=%s platform=%s lyricBytes=%d elapsed=%s", sourceID, candidate, len(lyric.Lyric), time.Since(startedAt))
			return lyric, nil
		}
		lastErr = callErr
		log.Printf("[后端:音源管理] 获取歌词重试 source=%s platform=%s err=%v", sourceID, candidate, callErr)
	}

	log.Printf("[后端:音源管理] 运行时获取歌词失败 source=%s platforms=%v err=%v elapsed=%s", sourceID, platformCandidates, lastErr, time.Since(startedAt))
	return nil, lastErr
}

func (m *Manager) GetPic(sourceID, platform, metaJSON string) (string, error) {
	startedAt := time.Now()
	log.Printf("[后端:音源管理] 开始获取封面 source=%s platform=%s metaBytes=%d", sourceID, platform, len(metaJSON))

	rt, musicInfo, err := m.prepareRequest(sourceID, platform, metaJSON)
	if err != nil {
		log.Printf("[后端:音源管理] 获取封面前准备请求失败 source=%s err=%v elapsed=%s", sourceID, err, time.Since(startedAt))
		return "", err
	}
	log.Printf("[后端:音源管理] 调用运行时获取封面 source=%s platform=%s song=%q singer=%q songmid=%q hash=%q", sourceID, platform, musicInfoString(musicInfo, "name", "songName"), musicInfoString(musicInfo, "singer", "singerName"), musicInfoString(musicInfo, "songmid", "songMid", "id"), musicInfoString(musicInfo, "hash", "FileHash"))

	picURL, err := rt.GetPic(platform, musicInfo)
	if err != nil {
		log.Printf("[后端:音源管理] 运行时获取封面失败 source=%s platform=%s err=%v elapsed=%s", sourceID, platform, err, time.Since(startedAt))
		return "", err
	}

	log.Printf("[后端:音源管理] 获取封面完成 source=%s platform=%s urlBytes=%d elapsed=%s", sourceID, platform, len(picURL), time.Since(startedAt))
	return picURL, nil
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
		return fmt.Errorf("未找到音源")
	}
	if !info.Enabled {
		return fmt.Errorf("音源已禁用")
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
		return nil, nil, fmt.Errorf("音源运行时不可用")
	}

	musicInfo := map[string]interface{}{}
	if metaJSON != "" {
		if err := json.Unmarshal([]byte(metaJSON), &musicInfo); err != nil {
			return nil, nil, fmt.Errorf("歌曲元数据格式无效: %w", err)
		}
	}
	if platform == "" {
		platform = m.defaultPlatform(sourceID)
	}
	log.Printf("[后端:音源管理] 准备请求 source=%s platform=%s song=%q singer=%q songmid=%q hash=%q", sourceID, platform, musicInfoString(musicInfo, "name", "songName"), musicInfoString(musicInfo, "singer", "singerName"), musicInfoString(musicInfo, "songmid", "songMid", "id"), musicInfoString(musicInfo, "hash", "FileHash"))
	return rt, musicInfo, nil
}

func musicInfoString(musicInfo map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := musicInfo[key]
		if !ok || value == nil {
			continue
		}
		text := fmt.Sprint(value)
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func normalizeMusicInfoAliases(musicInfo map[string]interface{}) {
	setMusicInfoAlias(musicInfo, "hash", "FileHash", "fileHash")
	setMusicInfoAlias(musicInfo, "songmid", "songMid", "id")
}

func setMusicInfoAlias(musicInfo map[string]interface{}, target string, candidates ...string) {
	if musicInfo == nil {
		return
	}
	if value := musicInfoString(musicInfo, target); value != "" {
		return
	}
	for _, key := range candidates {
		if value := musicInfoString(musicInfo, key); value != "" {
			musicInfo[target] = value
			return
		}
	}
}

func (m *Manager) platformCandidatesForAction(sourceID, preferred, action string) []string {
	info, ok := m.store.Get(sourceID)
	if !ok {
		if preferred != "" {
			return []string{preferred}
		}
		return []string{m.defaultPlatform(sourceID)}
	}

	seen := map[string]bool{}
	candidates := make([]string, 0, len(info.Platforms))
	appendCandidate := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		candidates = append(candidates, p)
	}

	appendCandidate(preferred)

	for _, item := range info.Platforms {
		if supportsAction(item.Actions, action) {
			appendCandidate(item.Key)
		}
	}
	for _, item := range info.Platforms {
		appendCandidate(item.Key)
	}

	if len(candidates) == 0 {
		appendCandidate(m.defaultPlatform(sourceID))
	}
	return candidates
}

func supportsAction(actions []string, action string) bool {
	for _, item := range actions {
		if item == action {
			return true
		}
	}
	return false
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
