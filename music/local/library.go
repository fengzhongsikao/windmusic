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
	"windmusic/music/appdata"

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

	folderSongsByKey    map[string][]models.LocalSong
	folderSongsIndexValid bool
}

func (s *LocalLibraryStore) invalidateFolderSongsIndex() {
	s.folderSongsIndexValid = false
	s.folderSongsByKey = nil
}

func (s *LocalLibraryStore) ensureFolderSongsIndexLocked(folders []string, cache *localScanCacheFile) {
	if s.folderSongsIndexValid {
		return
	}
	byKey := make(map[string][]models.LocalSong, len(folders)+1)
	byKey[models.LocalAllTabID] = make([]models.LocalSong, 0)
	for _, folder := range folders {
		byKey[folder] = make([]models.LocalSong, 0)
	}
	for path, entry := range cache.Entries {
		if !fileInMusicFolders(path, folders) {
			continue
		}
		byKey[models.LocalAllTabID] = append(byKey[models.LocalAllTabID], entry.Song)
		for _, folder := range foldersMatchingFile(path, folders) {
			byKey[folder] = append(byKey[folder], entry.Song)
		}
	}
	for key := range byKey {
		sortLocalSongs(byKey[key])
	}
	s.folderSongsByKey = byKey
	s.folderSongsIndexValid = true
}

type localFoldersFile struct {
	Paths   []string          `json:"paths"`
	Aliases map[string]string `json:"aliases,omitempty"`
}

func (f *localFoldersFile) ensureMaps() {
	if f.Aliases == nil {
		f.Aliases = make(map[string]string)
	}
}

func pruneFolderAliases(file *localFoldersFile) {
	file.ensureMaps()
	if len(file.Aliases) == 0 {
		return
	}
	inPaths := make(map[string]struct{}, len(file.Paths))
	for _, p := range file.Paths {
		inPaths[p] = struct{}{}
	}
	for path := range file.Aliases {
		if _, ok := inPaths[path]; !ok {
			delete(file.Aliases, path)
		}
	}
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
	root, err := appdata.AppDataRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "local-folders.json"), nil
}

func (s *LocalLibraryStore) cachePath() (string, error) {
	root, err := appdata.AppDataRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "local-scan-cache.json"), nil
}

func (s *LocalLibraryStore) extrasPath() (string, error) {
	root, err := appdata.AppDataRootDir()
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

func (s *LocalLibraryStore) readFoldersFileLocked() (localFoldersFile, error) {
	path, err := s.foldersPath()
	if err != nil {
		return localFoldersFile{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return localFoldersFile{Paths: []string{}}, nil
		}
		return localFoldersFile{}, err
	}
	if len(data) == 0 {
		return localFoldersFile{Paths: []string{}}, nil
	}
	var payload localFoldersFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return localFoldersFile{}, err
	}
	payload.Paths = normalizeFolderPaths(payload.Paths)
	payload.ensureMaps()
	pruneFolderAliases(&payload)
	return payload, nil
}

func (s *LocalLibraryStore) writeFoldersFileLocked(file localFoldersFile) error {
	path, err := s.foldersPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file.Paths = normalizeFolderPaths(file.Paths)
	file.ensureMaps()
	pruneFolderAliases(&file)
	payload, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	s.invalidateFolderSongsIndex()
	return nil
}

func (s *LocalLibraryStore) readFoldersLocked() ([]string, error) {
	file, err := s.readFoldersFileLocked()
	if err != nil {
		return nil, err
	}
	return file.Paths, nil
}

func (s *LocalLibraryStore) writeFoldersLocked(paths []string) error {
	file, err := s.readFoldersFileLocked()
	if err != nil {
		return err
	}
	file.Paths = paths
	return s.writeFoldersFileLocked(file)
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
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	s.invalidateFolderSongsIndex()
	return nil
}

func (s *LocalLibraryStore) ListFolders() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readFoldersLocked()
}

func (s *LocalLibraryStore) ListFolderAliases() (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	file, err := s.readFoldersFileLocked()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(file.Aliases))
	for path, alias := range file.Aliases {
		trimmed := strings.TrimSpace(alias)
		if trimmed != "" {
			out[path] = trimmed
		}
	}
	return out, nil
}

const maxLocalFolderAliasLen = 80

func (s *LocalLibraryStore) SetFolderAlias(folderPath, alias string) error {
	folderPath = strings.TrimSpace(folderPath)
	if folderPath == "" {
		return fmt.Errorf("folder path is empty")
	}
	alias = strings.TrimSpace(alias)
	if len(alias) > maxLocalFolderAliasLen {
		return fmt.Errorf("alias is too long (max %d characters)", maxLocalFolderAliasLen)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.readFoldersFileLocked()
	if err != nil {
		return err
	}
	target, err := filepath.Abs(folderPath)
	if err != nil {
		target = folderPath
	}
	found := false
	for _, folder := range file.Paths {
		if folder == target {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("folder not found")
	}
	file.ensureMaps()
	if alias == "" {
		delete(file.Aliases, target)
	} else {
		file.Aliases[target] = alias
	}
	return s.writeFoldersFileLocked(file)
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

	sortLocalSongs(allSongs)
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

	sortLocalSongs(songs)
	return songs, nil
}

func fileInFolder(filePath, folder string) bool {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		abs = filePath
	}
	folderAbs, err := filepath.Abs(folder)
	if err != nil {
		folderAbs = folder
	}
	prefix := folderAbs + string(os.PathSeparator)
	return abs == folderAbs || strings.HasPrefix(abs, prefix)
}

func foldersMatchingFile(filePath string, folders []string) []string {
	matched := make([]string, 0, 1)
	for _, folder := range folders {
		if fileInFolder(filePath, folder) {
			matched = append(matched, folder)
		}
	}
	return matched
}

func sortLocalSongs(songs []models.LocalSong) {
	sort.Slice(songs, func(i, j int) bool {
		left := strings.ToLower(songs[i].Title)
		right := strings.ToLower(songs[j].Title)
		if left == right {
			return songs[i].FilePath < songs[j].FilePath
		}
		return left < right
	})
}

// FolderCounts returns song counts per folder tab (including LocalAllTabID).
func (s *LocalLibraryStore) FolderCounts() (map[string]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return nil, err
	}
	if len(folders) == 0 {
		return map[string]int{models.LocalAllTabID: 0}, nil
	}

	cache, err := s.readCacheLocked()
	if err != nil {
		return nil, err
	}

	s.ensureFolderSongsIndexLocked(folders, cache)
	counts := make(map[string]int, len(s.folderSongsByKey))
	for key, songs := range s.folderSongsByKey {
		counts[key] = len(songs)
	}
	return counts, nil
}

func cloneLocalSongForList(song models.LocalSong) models.LocalSong {
	song.CoverData = ""
	song.Lyric = ""
	return song
}

// AllFolderSongs returns a copy of the in-memory per-tab song index (built from scan cache).
func (s *LocalLibraryStore) AllFolderSongs() (map[string][]models.LocalSong, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return nil, err
	}
	if len(folders) == 0 {
		return map[string][]models.LocalSong{models.LocalAllTabID: {}}, nil
	}

	cache, err := s.readCacheLocked()
	if err != nil {
		return nil, err
	}

	s.ensureFolderSongsIndexLocked(folders, cache)
	out := make(map[string][]models.LocalSong, len(s.folderSongsByKey))
	for key, songs := range s.folderSongsByKey {
		cloned := make([]models.LocalSong, len(songs))
		for i, song := range songs {
			cloned[i] = cloneLocalSongForList(song)
		}
		out[key] = cloned
	}
	return out, nil
}

// ListCachedForFolder returns cached songs for one tab (models.LocalAllTabID = all folders).
func (s *LocalLibraryStore) ListCachedForFolder(folderKey string) ([]models.LocalSong, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	s.ensureFolderSongsIndexLocked(folders, cache)
	songs := s.folderSongsByKey[folderKey]
	if len(songs) == 0 {
		return []models.LocalSong{}, nil
	}
	out := make([]models.LocalSong, len(songs))
	copy(out, songs)
	return out, nil
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
