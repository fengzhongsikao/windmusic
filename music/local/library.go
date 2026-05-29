package local

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	models "windmusic/internal/music"
	"windmusic/music"

	"github.com/dhowden/tag"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const LocalPlatform = "local"

var audioExtensions = map[string]struct{}{
	".mp3":  {},
	".flac": {},
	".m4a":  {},
	".aac":  {},
	".ogg":  {},
	".wav":  {},
	".wma":  {},
}

type LocalLibraryStore struct {
	mu sync.RWMutex
}

type localFoldersFile struct {
	Paths []string `json:"paths"`
}

type localScanCacheEntry struct {
	ModTimeUnix int64            `json:"modTimeUnix"`
	Song        models.LocalSong `json:"song"`
}

type localScanCacheFile struct {
	Version int                            `json:"version"`
	Entries map[string]localScanCacheEntry `json:"entries"`
}

const localScanCacheVersion = 3

type localSongExtras struct {
	CoverData string
	Lyric     string
}

func (s *LocalLibraryStore) foldersPath() (string, error) {
	root, err := music.AppDataRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "local-folders.json"), nil
}

func (s *LocalLibraryStore) cachePath() (string, error) {
	root, err := music.AppDataRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "local-scan-cache.json"), nil
}

func (s *LocalLibraryStore) extrasPath() (string, error) {
	root, err := music.AppDataRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "local-scan-extras.json"), nil
}

func (s *LocalLibraryStore) readExtrasLocked() (*localExtrasFile, error) {
	path, err := s.extrasPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return newExtrasFile(), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return newExtrasFile(), nil
	}
	var extras localExtrasFile
	if err := json.Unmarshal(data, &extras); err != nil {
		return nil, err
	}
	extras.ensureMaps()
	if extras.normalize() {
		if err := s.writeExtrasLocked(&extras); err != nil {
			return nil, err
		}
	}
	return &extras, nil
}

func (s *LocalLibraryStore) writeExtrasLocked(extras *localExtrasFile) error {
	path, err := s.extrasPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	extras.ensureMaps()
	payload, err := json.MarshalIndent(extras, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func migrateLegacyCache(cache *localScanCacheFile, extras *localExtrasFile) {
	extras.ensureMaps()
	for path, entry := range cache.Entries {
		if entry.Song.CoverData == "" && entry.Song.Lyric == "" {
			continue
		}
		extras.assignSong(path, entry.Song.CoverData, entry.Song.Lyric)
		entry.Song.CoverData = ""
		entry.Song.Lyric = ""
		cache.Entries[path] = entry
	}
}

func (s *LocalLibraryStore) readFoldersLocked() ([]string, error) {
	path, err := s.foldersPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []string{}, nil
	}
	var payload localFoldersFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return normalizeFolderPaths(payload.Paths), nil
}

func (s *LocalLibraryStore) writeFoldersLocked(paths []string) error {
	path, err := s.foldersPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(localFoldersFile{Paths: normalizeFolderPaths(paths)}, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func normalizeFolderPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	}
	sort.Strings(out)
	return out
}

func (s *LocalLibraryStore) readCacheLocked() (*localScanCacheFile, error) {
	path, err := s.cachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return newScanCache(), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return newScanCache(), nil
	}
	var cache localScanCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	if cache.Entries == nil {
		cache.Entries = map[string]localScanCacheEntry{}
	}
	if cache.Version != localScanCacheVersion {
		extras, err := s.readExtrasLocked()
		if err != nil {
			return nil, err
		}
		if cache.Version == 2 {
			migrateLegacyCache(&cache, extras)
			cache.Version = localScanCacheVersion
			if err := s.writeExtrasLocked(extras); err != nil {
				return nil, err
			}
			if err := s.writeCacheLocked(&cache); err != nil {
				return nil, err
			}
		} else {
			return newScanCache(), nil
		}
	}
	return &cache, nil
}

func newScanCache() *localScanCacheFile {
	return &localScanCacheFile{
		Version: localScanCacheVersion,
		Entries: map[string]localScanCacheEntry{},
	}
}

func (s *LocalLibraryStore) writeCacheLocked(cache *localScanCacheFile) error {
	path, err := s.cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if cache.Entries == nil {
		cache.Entries = map[string]localScanCacheEntry{}
	}
	cache.Version = localScanCacheVersion
	payload, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func (s *LocalLibraryStore) ListFolders() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readFoldersLocked()
}

func (s *LocalLibraryStore) AddFolder(folderPath string) error {
	folderPath = strings.TrimSpace(folderPath)
	if folderPath == "" {
		return fmt.Errorf("folder path is empty")
	}
	info, err := os.Stat(folderPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", folderPath)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(folderPath)
	if err != nil {
		abs = folderPath
	}
	for _, existing := range folders {
		if existing == abs {
			return nil
		}
	}
	folders = append(folders, abs)
	return s.writeFoldersLocked(folders)
}

func (s *LocalLibraryStore) RemoveFolder(folderPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return err
	}
	target, err := filepath.Abs(strings.TrimSpace(folderPath))
	if err != nil {
		target = strings.TrimSpace(folderPath)
	}
	next := make([]string, 0, len(folders))
	for _, folder := range folders {
		if folder != target {
			next = append(next, folder)
		}
	}
	if len(next) == len(folders) {
		return fmt.Errorf("folder not found")
	}
	if err := s.writeFoldersLocked(next); err != nil {
		return err
	}

	cache, err := s.readCacheLocked()
	if err != nil {
		return err
	}
	extras, err := s.readExtrasLocked()
	if err != nil {
		return err
	}
	prefix := target + string(os.PathSeparator)
	for path := range cache.Entries {
		if path == target || strings.HasPrefix(path, prefix) {
			delete(cache.Entries, path)
			delete(extras.Entries, path)
		}
	}
	if err := s.writeCacheLocked(cache); err != nil {
		return err
	}
	return s.writeExtrasLocked(extras)
}

func PickMusicFolder(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("application context is not ready")
	}
	dir, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: "选择音乐文件夹",
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(dir), nil
}

func cloneScanCache(cache *localScanCacheFile) *localScanCacheFile {
	out := newScanCache()
	out.Version = cache.Version
	for path, entry := range cache.Entries {
		out.Entries[path] = entry
	}
	return out
}

func cloneExtrasFile(extras *localExtrasFile) *localExtrasFile {
	out := newExtrasFile()
	out.Version = extras.Version
	for path, entry := range extras.Entries {
		out.Entries[path] = entry
	}
	for key, cover := range extras.Covers {
		out.Covers[key] = cover
	}
	return out
}

func (s *LocalLibraryStore) Scan() ([]models.LocalSong, error) {
	s.mu.Lock()
	folders, err := s.readFoldersLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if len(folders) == 0 {
		s.mu.Unlock()
		return []models.LocalSong{}, nil
	}

	cache, err := s.readCacheLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	extras, err := s.readExtrasLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	workingCache := cloneScanCache(cache)
	workingExtras := cloneExtrasFile(extras)
	s.mu.Unlock()

	type folderResult struct {
		songs []models.LocalSong
		paths map[string]struct{}
	}

	results := make([]folderResult, len(folders))
	var wg sync.WaitGroup
	var scanErr error
	var scanErrMu sync.Mutex
	var cacheMu sync.Mutex

	for i, folder := range folders {
		wg.Add(1)
		go func(idx int, root string) {
			defer wg.Done()
			songs, paths, err := scanFolder(root, workingCache, workingExtras, &cacheMu)
			if err != nil {
				scanErrMu.Lock()
				if scanErr == nil {
					scanErr = err
				}
				scanErrMu.Unlock()
				return
			}
			results[idx] = folderResult{songs: songs, paths: paths}
		}(i, folder)
	}
	wg.Wait()
	if scanErr != nil {
		return nil, scanErr
	}

	alive := make(map[string]struct{})
	allSongs := make([]models.LocalSong, 0)
	for _, result := range results {
		for path := range result.paths {
			alive[path] = struct{}{}
		}
		allSongs = append(allSongs, result.songs...)
	}

	for path := range workingCache.Entries {
		if _, ok := alive[path]; !ok {
			delete(workingCache.Entries, path)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writeCacheLocked(workingCache); err != nil {
		return nil, err
	}
	workingExtras.pruneUnusedCovers()
	if err := s.writeExtrasLocked(workingExtras); err != nil {
		return nil, err
	}

	sort.Slice(allSongs, func(i, j int) bool {
		left := strings.ToLower(allSongs[i].Title)
		right := strings.ToLower(allSongs[j].Title)
		if left == right {
			return allSongs[i].FilePath < allSongs[j].FilePath
		}
		return left < right
	})

	return allSongs, nil
}

// ListCached returns songs from the on-disk scan cache without walking the filesystem.
func (s *LocalLibraryStore) ListCached() ([]models.LocalSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return nil, err
	}
	if len(folders) == 0 {
		return []models.LocalSong{}, nil
	}

	cache, err := s.readCacheLocked()
	if err != nil {
		return nil, err
	}

	songs := make([]models.LocalSong, 0, len(cache.Entries))
	for path, entry := range cache.Entries {
		if !fileInMusicFolders(path, folders) {
			continue
		}
		songs = append(songs, entry.Song)
	}

	sort.Slice(songs, func(i, j int) bool {
		left := strings.ToLower(songs[i].Title)
		right := strings.ToLower(songs[j].Title)
		if left == right {
			return songs[i].FilePath < songs[j].FilePath
		}
		return left < right
	})

	return songs, nil
}

func fileInMusicFolders(filePath string, folders []string) bool {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		abs = filePath
	}
	for _, folder := range folders {
		folderAbs, err := filepath.Abs(folder)
		if err != nil {
			folderAbs = folder
		}
		prefix := folderAbs + string(os.PathSeparator)
		if strings.HasPrefix(abs, prefix) {
			return true
		}
	}
	return false
}

func scanFolder(root string, cache *localScanCacheFile, extras *localExtrasFile, cacheMu *sync.Mutex) ([]models.LocalSong, map[string]struct{}, error) {
	songs := make([]models.LocalSong, 0)
	alive := make(map[string]struct{})

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := audioExtensions[ext]; !ok {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}
		alive[absPath] = struct{}{}

		modUnix := info.ModTime().Unix()
		cacheMu.Lock()
		cached, ok := cache.Entries[absPath]
		cacheMu.Unlock()
		if ok && cached.ModTimeUnix == modUnix {
			songs = append(songs, cached.Song)
			return nil
		}

		song, songExtras, err := buildLocalSong(absPath, info)
		if err != nil {
			return nil
		}
		cacheMu.Lock()
		cache.Entries[absPath] = localScanCacheEntry{
			ModTimeUnix: modUnix,
			Song:        song,
		}
		extras.assignSong(absPath, songExtras.CoverData, songExtras.Lyric)
		cacheMu.Unlock()
		songs = append(songs, song)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return songs, alive, nil
}

func (s *LocalLibraryStore) GetSongExtras(filePath string) (models.LocalSongExtras, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return models.LocalSongExtras{}, err
	}
	if err := validatePathWithFolders(filePath, folders); err != nil {
		return models.LocalSongExtras{}, err
	}
	abs, err := filepath.Abs(strings.TrimSpace(filePath))
	if err != nil {
		return models.LocalSongExtras{}, err
	}

	extras, err := s.readExtrasLocked()
	if err != nil {
		return models.LocalSongExtras{}, err
	}
	if _, ok := extras.Entries[abs]; !ok {
		return models.LocalSongExtras{}, nil
	}
	return models.LocalSongExtras{
		CoverData: extras.coverForPath(abs),
		Lyric:     extras.lyricForPath(abs),
	}, nil
}

// GetCovers returns deduplicated cover blobs for the given library file paths.
func (s *LocalLibraryStore) GetCovers(filePaths []string) (models.LocalCoverBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return models.LocalCoverBatch{}, err
	}
	extras, err := s.readExtrasLocked()
	if err != nil {
		return models.LocalCoverBatch{}, err
	}

	allowed := make([]string, 0, len(filePaths))
	for _, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}
		abs, err := filepath.Abs(filePath)
		if err != nil {
			abs = filePath
		}
		if fileInMusicFolders(abs, folders) {
			allowed = append(allowed, abs)
		}
	}
	return extras.buildCoverBatch(allowed), nil
}

func buildLocalSong(absPath string, info fs.FileInfo) (models.LocalSong, localSongExtras, error) {
	ext := strings.ToLower(filepath.Ext(absPath))
	base := strings.TrimSuffix(filepath.Base(absPath), ext)

	title := base
	artist := "未知艺术家"
	album := ""
	coverData := ""

	file, err := os.Open(absPath)
	if err == nil {
		defer file.Close()
		metadata, err := tag.ReadFrom(file)
		if err == nil {
			if v := strings.TrimSpace(metadata.Title()); v != "" {
				title = v
			}
			if v := strings.TrimSpace(metadata.Artist()); v != "" {
				artist = v
			}
			if v := strings.TrimSpace(metadata.Album()); v != "" {
				album = v
			}
			if pic := metadata.Picture(); pic != nil && len(pic.Data) > 0 {
				mime := strings.TrimSpace(pic.MIMEType)
				if mime == "" {
					mime = "image/jpeg"
				}
				encoded := base64.StdEncoding.EncodeToString(pic.Data)
				coverData = fmt.Sprintf("data:%s;base64,%s", mime, encoded)
			}
		}
	}

	lyric := readSidecarLyric(absPath)
	duration := formatTrackDuration(probeAudioDurationSeconds(absPath, ext))

	return models.LocalSong{
			ID:       absPath,
			Title:    title,
			Artist:   artist,
			Album:    album,
			Duration: duration,
			FilePath: absPath,
			Format:   strings.TrimPrefix(strings.ToUpper(ext), "."),
			Size:     formatFileSize(info.Size()),
		}, localSongExtras{
			CoverData: coverData,
			Lyric:     lyric,
		}, nil
}

func readSidecarLyric(audioPath string) string {
	if strings.ToLower(filepath.Ext(audioPath)) != ".mp3" {
		return ""
	}
	lrcPath := strings.TrimSuffix(audioPath, filepath.Ext(audioPath)) + ".lrc"
	data, err := os.ReadFile(lrcPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(size) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB"}[exp]
	return fmt.Sprintf("%.1f %s", value, suffix)
}

func audioMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	case ".m4a":
		return "audio/mp4"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	case ".wma":
		return "audio/x-ms-wma"
	default:
		return "application/octet-stream"
	}
}

func validatePathWithFolders(filePath string, folders []string) error {
	abs, err := filepath.Abs(strings.TrimSpace(filePath))
	if err != nil {
		return fmt.Errorf("invalid file path")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}
	if !fileInMusicFolders(abs, folders) {
		return fmt.Errorf("file is outside configured music folders")
	}
	return nil
}

func (s *LocalLibraryStore) ValidateLibraryPath(filePath string) error {
	s.mu.RLock()
	folders, err := s.readFoldersLocked()
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return validatePathWithFolders(filePath, folders)
}

func GetLocalAudioStream(store *LocalLibraryStore, filePath string) (string, error) {
	if store == nil {
		return "", fmt.Errorf("local library is not initialized")
	}
	if err := store.ValidateLibraryPath(filePath); err != nil {
		return "", err
	}

	abs, err := filepath.Abs(strings.TrimSpace(filePath))
	if err != nil {
		return "", err
	}

	file, err := os.Open(abs)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	mime := audioMIME(filepath.Ext(abs))
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}
