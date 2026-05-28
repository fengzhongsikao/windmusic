package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"windmusic/internal/music"
	"windmusic/internal/musicsearch"
)

type App struct {
	ctx           context.Context
	favoritesPath string
	favoritesMu   sync.Mutex
}

const metingSourcePrefix = "meting::"
const defaultMetingBaseURL = "https://meting.mikus.ink"

func (a *App) sourceDisplayName(sourceID string) string {
	if base, ok := parseMetingSourceID(sourceID); ok {
		return fmt.Sprintf("Meting(%s)", base)
	}
	return sourceID
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Search(sourceID, platform, keyword string, page int) (*music.SearchResult, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	logPrefix := backendLogPrefix(sourceID)
	log.Printf("%s 开始搜索 source=%s platform=%s page=%d keyword=%q", logPrefix, source, platform, page, keyword)

	base := resolveMetingBase(sourceID)
	result, err := musicsearch.SearchMeting(base, platform, keyword, page, 20)
	if err != nil {
		log.Printf("%s 搜索失败 source=%s platform=%s err=%v elapsed=%s", logPrefix, source, platform, err, time.Since(startedAt))
		return nil, err
	}
	log.Printf("%s 搜索完成 source=%s platform=%s total=%d list=%d elapsed=%s", logPrefix, source, result.Source, result.Total, len(result.List), time.Since(startedAt))
	return result, nil
}

func (a *App) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	logPrefix := backendLogPrefix(sourceID)
	log.Printf("%s 开始获取播放地址 source=%s platform=%s quality=%s metaBytes=%d", logPrefix, source, platform, quality, len(metaJSON))

	base := resolveMetingBase(sourceID)
	url, err := musicsearch.GetMetingMusicURL(base, platform, metaJSON)
	if err != nil {
		log.Printf("%s 获取播放地址失败 source=%s platform=%s quality=%s err=%v elapsed=%s", logPrefix, source, platform, quality, err, time.Since(startedAt))
		return "", err
	}
	log.Printf("%s 获取播放地址完成 source=%s platform=%s quality=%s urlBytes=%d elapsed=%s musicUrl=%s", logPrefix, source, platform, quality, len(url), time.Since(startedAt), url)
	return url, nil
}

func (a *App) GetLyric(sourceID, platform, metaJSON string) (*music.LyricInfo, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	logPrefix := backendLogPrefix(sourceID)
	log.Printf("%s 开始获取歌词 source=%s platform=%s metaBytes=%d", logPrefix, source, platform, len(metaJSON))

	base := resolveMetingBase(sourceID)
	lyric, err := musicsearch.GetMetingLyric(base, platform, metaJSON)
	if err != nil {
		log.Printf("%s 获取歌词失败 source=%s platform=%s err=%v elapsed=%s", logPrefix, source, platform, err, time.Since(startedAt))
		return nil, err
	}
	log.Printf("%s 获取歌词完成 source=%s platform=%s lyricBytes=%d elapsed=%s", logPrefix, source, platform, len(lyric.Lyric), time.Since(startedAt))
	return lyric, nil
}

func (a *App) GetPic(sourceID, platform, metaJSON string) (string, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	logPrefix := backendLogPrefix(sourceID)
	log.Printf("%s 开始获取封面 source=%s platform=%s metaBytes=%d", logPrefix, source, platform, len(metaJSON))

	base := resolveMetingBase(sourceID)
	picURL, err := musicsearch.GetMetingPic(base, platform, metaJSON)
	if err != nil {
		log.Printf("%s 获取封面失败 source=%s platform=%s err=%v elapsed=%s", logPrefix, source, platform, err, time.Since(startedAt))
		return "", err
	}
	log.Printf("%s 获取封面完成 source=%s platform=%s urlBytes=%d elapsed=%s", logPrefix, source, platform, len(picURL), time.Since(startedAt))
	return picURL, nil
}

func (a *App) GetSourceDataDir() (string, error) {
	rootDir, err := appDataRootDir()
	if err != nil {
		return "", err
	}
	return rootDir, nil
}

func (a *App) ListFavorites() ([]music.FavoriteSong, error) {
	a.favoritesMu.Lock()
	defer a.favoritesMu.Unlock()

	return a.readFavoritesLocked()
}

func (a *App) IsFavorite(song music.FavoriteSong) (bool, error) {
	a.favoritesMu.Lock()
	defer a.favoritesMu.Unlock()

	favorites, err := a.readFavoritesLocked()
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

func (a *App) AddFavorite(song music.FavoriteSong) error {
	a.favoritesMu.Lock()
	defer a.favoritesMu.Unlock()

	favorites, err := a.readFavoritesLocked()
	if err != nil {
		return err
	}

	for _, item := range favorites {
		if sameFavoriteSong(item, song) {
			return nil
		}
	}

	favorites = append(favorites, normalizeFavoriteSong(song))
	return a.writeFavoritesLocked(favorites)
}

func (a *App) RemoveFavorite(song music.FavoriteSong) error {
	a.favoritesMu.Lock()
	defer a.favoritesMu.Unlock()

	favorites, err := a.readFavoritesLocked()
	if err != nil {
		return err
	}

	next := make([]music.FavoriteSong, 0, len(favorites))
	for _, item := range favorites {
		if !sameFavoriteSong(item, song) {
			next = append(next, item)
		}
	}
	return a.writeFavoritesLocked(next)
}

func (a *App) ensureFavoritesPath() (string, error) {
	if a.favoritesPath != "" {
		return a.favoritesPath, nil
	}

	rootDir, err := appDataRootDir()
	if err != nil {
		return "", err
	}
	a.favoritesPath = filepath.Join(rootDir, "favorites.json")
	return a.favoritesPath, nil
}

func (a *App) readFavoritesLocked() ([]music.FavoriteSong, error) {
	path, err := a.ensureFavoritesPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []music.FavoriteSong{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []music.FavoriteSong{}, nil
	}

	var favorites []music.FavoriteSong
	if err := json.Unmarshal(data, &favorites); err != nil {
		return nil, err
	}
	return favorites, nil
}

func (a *App) writeFavoritesLocked(favorites []music.FavoriteSong) error {
	path, err := a.ensureFavoritesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	normalized := make([]music.FavoriteSong, 0, len(favorites))
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

	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func normalizeFavoriteSong(song music.FavoriteSong) music.FavoriteSong {
	return music.FavoriteSong{
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

func favoriteSongKey(song music.FavoriteSong) string {
	entry := normalizeFavoriteSong(song)
	return strings.Join([]string{entry.ID, entry.SourceID, entry.Platform, entry.MetaJSON}, "|")
}

func sameFavoriteSong(a, b music.FavoriteSong) bool {
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

func parseMetingSourceID(sourceID string) (string, bool) {
	if !strings.HasPrefix(sourceID, metingSourcePrefix) {
		return "", false
	}
	base := strings.TrimSpace(strings.TrimPrefix(sourceID, metingSourcePrefix))
	if base == "" {
		return "", false
	}
	return strings.TrimSuffix(base, "/"), true
}

func resolveMetingBase(sourceID string) string {
	if base, ok := parseMetingSourceID(sourceID); ok {
		return base
	}
	return defaultMetingBaseURL
}

func backendLogPrefix(sourceID string) string {
	if _, ok := parseMetingSourceID(sourceID); ok {
		return "[后端meting]"
	}
	return "[后端网络源]"
}

func appDataRootDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "windmusic"), nil
}
