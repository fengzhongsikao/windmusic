package local

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	stdruntime "runtime"

	models "windmusic/internal/music"
	"windmusic/music/appdata"

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

	foldersLoaded bool
	foldersFile   localFoldersFile

	folderSongsByKey      map[string][]models.LocalSong
	folderSongsIndexValid bool

	cacheLoaded  bool
	extrasLoaded bool
	cache        *localScanCacheFile
	extras       *localExtrasFile
	cacheDirty   bool
	extrasDirty  bool

	db         *libraryDB
	coverFiles *coverFileStore
	coverURL   func(string) string
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

func (s *LocalLibraryStore) ensureScanCacheLoaded() error {
	s.mu.RLock()
	if s.cacheLoaded {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadScanCacheFromDiskLocked()
}

func (s *LocalLibraryStore) ensureExtrasLoaded() error {
	s.mu.RLock()
	if s.extrasLoaded {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadExtrasFromDiskLocked()
}

func (s *LocalLibraryStore) SetCoverURLBuilder(fn func(string) string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.coverURL = fn
}

func (s *LocalLibraryStore) loadExtrasFromDiskLocked() error {
	if s.extrasLoaded {
		return nil
	}
	if err := s.ensureLibraryDB(); err != nil {
		return err
	}
	extras, err := s.db.loadExtras()
	if err != nil {
		return err
	}
	s.extras = extras
	s.extrasLoaded = true
	return nil
}

func (s *LocalLibraryStore) loadScanCacheFromDiskLocked() error {
	if s.cacheLoaded {
		return nil
	}
	if err := s.ensureLibraryDB(); err != nil {
		return err
	}
	cache, err := s.db.loadScanCache()
	if err != nil {
		return err
	}
	if cache == nil || len(cache.Entries) == 0 {
		s.cache = newScanCache()
		s.cacheLoaded = true
		return nil
	}
	s.cache = cache
	s.cacheLoaded = true
	return nil
}

func (s *LocalLibraryStore) flushExtrasLocked() error {
	if !s.extrasDirty || s.extras == nil {
		return nil
	}
	if err := s.writeExtrasToDiskLocked(s.extras); err != nil {
		return err
	}
	s.extrasDirty = false
	return nil
}

func (s *LocalLibraryStore) writeExtrasToDiskLocked(extras *localExtrasFile) error {
	if s.db == nil {
		return fmt.Errorf("local library db is not open")
	}
	extras.ensureMaps()
	return s.db.saveExtras(extras)
}

func (s *LocalLibraryStore) flushScanCacheLocked() error {
	if !s.cacheDirty || s.cache == nil {
		return nil
	}
	if err := s.writeScanCacheToDiskLocked(s.cache); err != nil {
		return err
	}
	s.cacheDirty = false
	return nil
}

func (s *LocalLibraryStore) flushCachesLocked() error {
	if err := s.flushScanCacheLocked(); err != nil {
		return err
	}
	return s.flushExtrasLocked()
}


func (s *LocalLibraryStore) readFoldersFileFromDisk() (localFoldersFile, error) {
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

func (s *LocalLibraryStore) ensureFoldersLoadedLocked() error {
	if s.foldersLoaded {
		return nil
	}
	file, err := s.readFoldersFileFromDisk()
	if err != nil {
		return err
	}
	s.foldersFile = file
	s.foldersLoaded = true
	return nil
}

func (s *LocalLibraryStore) readFoldersFileLocked() (localFoldersFile, error) {
	if err := s.ensureFoldersLoadedLocked(); err != nil {
		return localFoldersFile{}, err
	}
	return s.foldersFile, nil
}

func (s *LocalLibraryStore) writeFoldersFileToDisk(file localFoldersFile) error {
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
	return nil
}

func (s *LocalLibraryStore) writeFoldersFileLocked(file localFoldersFile) error {
	file.Paths = normalizeFolderPaths(file.Paths)
	file.ensureMaps()
	pruneFolderAliases(&file)
	if err := s.writeFoldersFileToDisk(file); err != nil {
		return err
	}
	s.foldersFile = file
	s.foldersLoaded = true
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

func newScanCache() *localScanCacheFile {
	return &localScanCacheFile{
		Version: localScanCacheVersion,
		Entries: map[string]localScanCacheEntry{},
	}
}

func (s *LocalLibraryStore) writeScanCacheToDiskLocked(cache *localScanCacheFile) error {
	if s.db == nil {
		return fmt.Errorf("local library db is not open")
	}
	if cache.Entries == nil {
		cache.Entries = map[string]localScanCacheEntry{}
	}
	cache.Version = localScanCacheVersion
	return s.db.saveScanCache(cache)
}

func scanCacheEqual(a, b *localScanCacheFile) bool {
	if a == nil || b == nil {
		return a == b
	}
	if len(a.Entries) != len(b.Entries) {
		return false
	}
	for path, bv := range b.Entries {
		av, ok := a.Entries[path]
		if !ok || av.ModTimeUnix != bv.ModTimeUnix {
			return false
		}
	}
	return true
}

func extrasEqual(a, b *localExtrasFile) bool {
	if a == nil || b == nil {
		return a == b
	}
	if len(a.Entries) != len(b.Entries) {
		return false
	}
	for path, ref := range b.Entries {
		if a.Entries[path] != ref {
			return false
		}
	}
	return true
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

	if err := s.loadScanCacheFromDiskLocked(); err != nil {
		return err
	}
	if err := s.loadExtrasFromDiskLocked(); err != nil {
		return err
	}
	prefix := target + string(os.PathSeparator)
	for path := range s.cache.Entries {
		if path == target || strings.HasPrefix(path, prefix) {
			delete(s.cache.Entries, path)
			delete(s.extras.Entries, path)
		}
	}
	s.cacheDirty = true
	s.extrasDirty = true
	s.invalidateFolderSongsIndex()
	return s.flushCachesLocked()
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

	if err := s.loadScanCacheFromDiskLocked(); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if err := s.loadExtrasFromDiskLocked(); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	workingCache := cloneScanCache(s.cache)
	workingExtras := cloneExtrasFile(s.extras)
	coverFiles := s.coverFiles
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
			songs, paths, err := scanFolder(root, workingCache, workingExtras, coverFiles, &cacheMu)
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

	workingExtras.pruneUnusedCovers()
	if !scanCacheEqual(s.cache, workingCache) {
		s.cache = workingCache
		s.cacheDirty = true
		s.invalidateFolderSongsIndex()
	}
	if !extrasEqual(s.extras, workingExtras) {
		s.extras = workingExtras
		s.extrasDirty = true
	}
	if err := s.flushCachesLocked(); err != nil {
		return nil, err
	}

	sortLocalSongs(allSongs)
	return allSongs, nil
}

// ListCached returns songs from the in-memory scan cache without walking the filesystem.
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

	if err := s.loadScanCacheFromDiskLocked(); err != nil {
		return nil, err
	}

	songs := make([]models.LocalSong, 0, len(s.cache.Entries))
	for path, entry := range s.cache.Entries {
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
	if err := s.ensureScanCacheLoaded(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return nil, err
	}
	if len(folders) == 0 {
		return map[string]int{models.LocalAllTabID: 0}, nil
	}

	s.ensureFolderSongsIndexLocked(folders, s.cache)
	counts := make(map[string]int, len(s.folderSongsByKey))
	for key, songs := range s.folderSongsByKey {
		counts[key] = len(songs)
	}
	return counts, nil
}

// Snapshot returns folder paths, aliases, and per-tab song counts in one lock acquisition.
func (s *LocalLibraryStore) Snapshot() (models.LocalLibrarySnapshot, error) {
	if err := s.ensureScanCacheLoaded(); err != nil {
		return models.LocalLibrarySnapshot{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureFoldersLoadedLocked(); err != nil {
		return models.LocalLibrarySnapshot{}, err
	}

	folders := s.foldersFile.Paths
	if folders == nil {
		folders = []string{}
	}

	counts := map[string]int{models.LocalAllTabID: 0}
	if len(folders) > 0 && s.cache != nil {
		s.ensureFolderSongsIndexLocked(folders, s.cache)
		for key, songs := range s.folderSongsByKey {
			counts[key] = len(songs)
		}
	}

	aliases := make(map[string]string, len(s.foldersFile.Aliases))
	for path, alias := range s.foldersFile.Aliases {
		trimmed := strings.TrimSpace(alias)
		if trimmed != "" {
			aliases[path] = trimmed
		}
	}

	return models.LocalLibrarySnapshot{
		Folders:       folders,
		FolderAliases: aliases,
		FolderCounts:  counts,
	}, nil
}

func cloneLocalSongForList(song models.LocalSong) models.LocalSong {
	song.CoverData = ""
	song.Lyric = ""
	return song
}

// AllFolderSongs returns a copy of the in-memory per-tab song index (built from scan cache).
func (s *LocalLibraryStore) AllFolderSongs() (map[string][]models.LocalSong, error) {
	if err := s.ensureScanCacheLoaded(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return nil, err
	}
	if len(folders) == 0 {
		return map[string][]models.LocalSong{models.LocalAllTabID: {}}, nil
	}

	s.ensureFolderSongsIndexLocked(folders, s.cache)
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

const (
	defaultLocalSongPageSize = 1500
	maxLocalSongPageSize     = 3000
)

func normalizeLocalSongPage(offset, limit int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultLocalSongPageSize
	}
	if limit > maxLocalSongPageSize {
		limit = maxLocalSongPageSize
	}
	return offset, limit
}

// ListCachedForFolder returns cached songs for one tab (models.LocalAllTabID = all folders).
func (s *LocalLibraryStore) ListCachedForFolder(folderKey string) ([]models.LocalSong, error) {
	page, err := s.ListCachedForFolderPage(folderKey, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(page.Songs) == 0 {
		return []models.LocalSong{}, nil
	}
	if page.Total <= len(page.Songs) {
		return page.Songs, nil
	}
	out := make([]models.LocalSong, 0, page.Total)
	out = append(out, page.Songs...)
	for offset := len(page.Songs); offset < page.Total; {
		next, err := s.ListCachedForFolderPage(folderKey, offset, maxLocalSongPageSize)
		if err != nil {
			return nil, err
		}
		if len(next.Songs) == 0 {
			break
		}
		out = append(out, next.Songs...)
		offset += len(next.Songs)
	}
	return out, nil
}

// ListCachedForFolderPage returns one page of cached songs for a folder tab.
func (s *LocalLibraryStore) ListCachedForFolderPage(folderKey string, offset, limit int) (models.LocalSongPage, error) {
	offset, limit = normalizeLocalSongPage(offset, limit)

	if err := s.ensureScanCacheLoaded(); err != nil {
		return models.LocalSongPage{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	folders, err := s.readFoldersLocked()
	if err != nil {
		return models.LocalSongPage{}, err
	}
	if len(folders) == 0 {
		return models.LocalSongPage{Offset: offset}, nil
	}

	s.ensureFolderSongsIndexLocked(folders, s.cache)
	songs := s.folderSongsByKey[folderKey]
	total := len(songs)
	if total == 0 || offset >= total {
		return models.LocalSongPage{Total: total, Offset: offset}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}
	slice := songs[offset:end]
	out := make([]models.LocalSong, len(slice))
	for i, song := range slice {
		out[i] = cloneLocalSongForList(song)
	}
	return models.LocalSongPage{
		Songs:  out,
		Total:  total,
		Offset: offset,
	}, nil
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

type scanCandidate struct {
	absPath string
	modUnix int64
	size    int64
}

func scanWorkerCount() int {
	n := stdruntime.NumCPU()
	if n < 4 {
		return 4
	}
	if n > 12 {
		return 12
	}
	return n
}

func scanFolder(root string, cache *localScanCacheFile, _ *localExtrasFile, _ *coverFileStore, cacheMu *sync.Mutex) ([]models.LocalSong, map[string]struct{}, error) {
	candidates := make([]scanCandidate, 0, 256)
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
		candidates = append(candidates, scanCandidate{
			absPath: absPath,
			modUnix: info.ModTime().Unix(),
			size:    info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	songs := make([]models.LocalSong, 0, len(candidates))
	var songsMu sync.Mutex
	pending := make([]scanCandidate, 0, len(candidates))

	for _, candidate := range candidates {
		cacheMu.Lock()
		cached, ok := cache.Entries[candidate.absPath]
		cacheMu.Unlock()
		if ok && cached.ModTimeUnix == candidate.modUnix {
			songsMu.Lock()
			songs = append(songs, cached.Song)
			songsMu.Unlock()
			continue
		}
		pending = append(pending, candidate)
	}

	if len(pending) == 0 {
		return songs, alive, nil
	}

	workers := scanWorkerCount()
	if workers > len(pending) {
		workers = len(pending)
	}

	jobs := make(chan scanCandidate)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for candidate := range jobs {
				song, err := buildLocalSongFast(candidate.absPath, candidate.size)
				if err != nil {
					continue
				}
				cacheMu.Lock()
				cache.Entries[candidate.absPath] = localScanCacheEntry{
					ModTimeUnix: candidate.modUnix,
					Song:        song,
				}
				cacheMu.Unlock()
				songsMu.Lock()
				songs = append(songs, song)
				songsMu.Unlock()
			}
		}()
	}
	for _, candidate := range pending {
		jobs <- candidate
	}
	close(jobs)
	wg.Wait()

	return songs, alive, nil
}

func persistSongExtra(extras *localExtrasFile, coverFiles *coverFileStore, path, coverData, lyric string) {
	if extras == nil {
		return
	}
	coverKey := ""
	if strings.TrimSpace(coverData) != "" {
		coverKey = coverKeyForData(coverData)
		if coverFiles != nil {
			_ = coverFiles.SaveDataURL(coverKey, coverData)
		}
	}
	extras.assignSongRef(path, coverKey, lyric)
}

func (s *LocalLibraryStore) resolveCoverValue(coverKey string) string {
	coverKey = strings.TrimSpace(coverKey)
	if coverKey == "" {
		return ""
	}
	if s.coverURL != nil {
		if url := strings.TrimSpace(s.coverURL(coverKey)); url != "" {
			return url
		}
	}
	if s.coverFiles != nil {
		if data, err := s.coverFiles.ReadDataURL(coverKey); err == nil && data != "" {
			return data
		}
	}
	if s.extras != nil {
		return strings.TrimSpace(s.extras.Covers[coverKey])
	}
	return ""
}

func (s *LocalLibraryStore) GetSongExtras(filePath string) (models.LocalSongExtras, error) {
	if err := s.ensureExtrasLoaded(); err != nil {
		return models.LocalSongExtras{}, err
	}

	s.mu.RLock()
	folders, err := s.readFoldersLocked()
	if err != nil {
		s.mu.RUnlock()
		return models.LocalSongExtras{}, err
	}
	if err := validatePathWithFolders(filePath, folders); err != nil {
		s.mu.RUnlock()
		return models.LocalSongExtras{}, err
	}
	abs, err := filepath.Abs(strings.TrimSpace(filePath))
	if err != nil {
		s.mu.RUnlock()
		return models.LocalSongExtras{}, err
	}
	entry, ok := s.extras.Entries[abs]
	needsLoad := !ok || strings.TrimSpace(entry.CoverKey) == ""
	if !needsLoad && strings.ToLower(filepath.Ext(abs)) == ".mp3" && strings.TrimSpace(entry.Lyric) == "" {
		needsLoad = true
	}
	s.mu.RUnlock()

	if needsLoad {
		if err := s.ensureSongExtrasForPath(abs); err != nil {
			return models.LocalSongExtras{}, err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok = s.extras.Entries[abs]
	if !ok {
		return models.LocalSongExtras{}, nil
	}
	return models.LocalSongExtras{
		CoverData: s.resolveCoverValue(entry.CoverKey),
		Lyric:     entry.Lyric,
	}, nil
}

// GetCovers returns deduplicated cover blobs for the given library file paths.
func (s *LocalLibraryStore) GetCovers(filePaths []string) (models.LocalCoverBatch, error) {
	if err := s.ensureExtrasLoaded(); err != nil {
		return models.LocalCoverBatch{}, err
	}

	s.mu.RLock()
	folders, err := s.readFoldersLocked()
	if err != nil {
		s.mu.RUnlock()
		return models.LocalCoverBatch{}, err
	}

	allowed := make([]string, 0, len(filePaths))
	missingCovers := make([]string, 0, len(filePaths))
	for _, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}
		abs, err := filepath.Abs(filePath)
		if err != nil {
			abs = filePath
		}
		if !fileInMusicFolders(abs, folders) {
			continue
		}
		allowed = append(allowed, abs)
		entry, ok := s.extras.Entries[abs]
		if !ok || strings.TrimSpace(entry.CoverKey) == "" {
			missingCovers = append(missingCovers, abs)
		}
	}
	s.mu.RUnlock()

	for _, abs := range missingCovers {
		_ = s.ensureSongExtrasForPath(abs)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.extras.buildCoverBatch(allowed, s.coverURL, s.coverFiles), nil
}

func buildLocalSongFast(absPath string, size int64) (models.LocalSong, error) {
	ext := strings.ToLower(filepath.Ext(absPath))
	tags := readBasicAudioTags(absPath)
	duration := formatTrackDuration(tags.DurationSec)

	return models.LocalSong{
		ID:       absPath,
		Title:    tags.Title,
		Artist:   tags.Artist,
		Album:    tags.Album,
		Duration: duration,
		FilePath: absPath,
		Format:   strings.TrimPrefix(strings.ToUpper(ext), "."),
		Size:     formatFileSize(size),
	}, nil
}

func buildLocalSong(absPath string, info fs.FileInfo) (models.LocalSong, localSongExtras, error) {
	ext := strings.ToLower(filepath.Ext(absPath))
	tags := readBasicAudioTags(absPath)
	durationSec := tags.DurationSec
	if durationSec <= 0 {
		durationSec = probeAudioDuration(absPath, ext, info.Size())
	}

	coverData := extractEmbeddedCoverData(absPath)
	lyric := readSidecarLyric(absPath)

	return models.LocalSong{
			ID:       absPath,
			Title:    tags.Title,
			Artist:   tags.Artist,
			Album:    tags.Album,
			Duration: formatTrackDuration(durationSec),
			FilePath: absPath,
			Format:   strings.TrimPrefix(strings.ToUpper(ext), "."),
			Size:     formatFileSize(info.Size()),
		}, localSongExtras{
			CoverData: coverData,
			Lyric:     lyric,
		}, nil
}

func (s *LocalLibraryStore) ensureSongExtrasForPath(absPath string) error {
	if err := s.ensureExtrasLoaded(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.loadExtrasFromDiskLocked(); err != nil {
		return err
	}
	if s.extras == nil {
		s.extras = newExtrasFile()
	}

	entry, ok := s.extras.Entries[absPath]
	needCover := !ok || strings.TrimSpace(entry.CoverKey) == ""
	needLyric := !ok || (strings.TrimSpace(entry.Lyric) == "" && strings.ToLower(filepath.Ext(absPath)) == ".mp3")
	if !needCover && !needLyric {
		return nil
	}

	coverData := ""
	lyric := ""
	if needCover {
		coverData = extractEmbeddedCoverData(absPath)
	}
	if needLyric {
		lyric = readSidecarLyric(absPath)
	}
	if coverData == "" && lyric == "" {
		return nil
	}

	persistSongExtra(s.extras, s.coverFiles, absPath, coverData, lyric)
	s.extrasDirty = true
	return s.flushExtrasLocked()
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

// ClearCache removes all rows from local-library.db and deletes cached cover files.
// Configured music folders (local-folders.json) are not affected.
func (s *LocalLibraryStore) ClearCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLibraryDB(); err != nil {
		return err
	}
	if err := s.db.clearAll(); err != nil {
		return err
	}
	if s.coverFiles != nil {
		if err := s.coverFiles.ClearAll(); err != nil {
			return err
		}
	}

	s.cache = newScanCache()
	s.cacheLoaded = true
	s.cacheDirty = false
	s.extras = newExtrasFile()
	s.extrasLoaded = true
	s.extrasDirty = false
	s.invalidateFolderSongsIndex()
	return nil
}

