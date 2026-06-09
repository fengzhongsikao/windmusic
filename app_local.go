package main

import (
	models "windmusic/internal/music"
	localmusic "windmusic/music/local"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) startLocalLibraryScan() {
	a.localScanMu.Lock()
	if a.localScanRunning {
		a.localScanPending = true
		a.localScanMu.Unlock()
		return
	}
	a.localScanRunning = true
	a.localScanMu.Unlock()

	go a.runLocalLibraryScan()
}

func (a *App) runLocalLibraryScan() {
	a.emitLocalLibraryScanning(true)
	defer a.emitLocalLibraryScanning(false)

	if _, err := a.local.Scan(); err != nil {
		runtime.LogErrorf(a.ctx, "local library scan failed: %v", err)
	}
	a.emitLocalLibrarySnapshot()

	a.localScanMu.Lock()
	a.localScanRunning = false
	pending := a.localScanPending
	a.localScanPending = false
	a.localScanMu.Unlock()

	if pending {
		a.startLocalLibraryScan()
	}
}

func (a *App) bootstrapLocalLibrary() {
	a.emitLocalLibrarySnapshot()

	snapshot, err := a.localLibrarySnapshot()
	if err != nil {
		return
	}
	if len(snapshot.Folders) == 0 || snapshot.FolderCounts[models.LocalAllTabID] > 0 {
		return
	}

	a.startLocalLibraryScan()
}

func (a *App) PickLocalMusicFolder() (string, error) {
	dir, err := localmusic.PickMusicFolder(a.ctx)
	if err != nil || dir == "" {
		return dir, err
	}
	if err := a.local.AddFolder(dir); err != nil {
		return "", err
	}
	a.emitLocalLibrarySnapshot()
	a.startLocalLibraryScan()
	return dir, nil
}

func (a *App) GetLocalLibrarySnapshot() (models.LocalLibrarySnapshot, error) {
	return a.localLibrarySnapshot()
}

func (a *App) ListLocalFolders() ([]string, error) {
	return a.local.ListFolders()
}

func (a *App) RemoveLocalMusicFolder(folderPath string) error {
	if err := a.local.RemoveFolder(folderPath); err != nil {
		return err
	}
	a.emitLocalLibrarySnapshot()
	return nil
}

func (a *App) SetLocalFolderAlias(folderPath, alias string) error {
	if err := a.local.SetFolderAlias(folderPath, alias); err != nil {
		return err
	}
	a.emitLocalLibrarySnapshot()
	return nil
}

func (a *App) ListLocalLibrary() ([]models.LocalSong, error) {
	return a.local.ListCached()
}

func (a *App) GetLocalFolderSongs(folderKey string) ([]models.LocalSong, error) {
	songs, err := a.local.ListCachedForFolder(folderKey)
	if err != nil {
		return nil, err
	}
	if songs == nil {
		return []models.LocalSong{}, nil
	}
	return songs, nil
}

func (a *App) GetLocalFolderSongsPage(folderKey string, offset, limit int) (models.LocalSongPage, error) {
	return a.local.ListCachedForFolderPage(folderKey, offset, limit)
}

func (a *App) GetLocalLibraryTracksIndex() (map[string][]models.LocalSong, error) {
	index, err := a.local.AllFolderSongs()
	if err != nil {
		return nil, err
	}
	if index == nil {
		return map[string][]models.LocalSong{}, nil
	}
	return index, nil
}

func (a *App) ScanLocalLibrary() error {
	a.startLocalLibraryScan()
	return nil
}

func (a *App) GetLocalAudioStream(filePath string) (string, error) {
	return a.localAudio.StreamURL(filePath)
}

func (a *App) GetLocalSongExtras(filePath string) (models.LocalSongExtras, error) {
	return a.local.GetSongExtras(filePath)
}

func (a *App) GetLocalSongCovers(filePaths []string) (models.LocalCoverBatch, error) {
	return a.local.GetCovers(filePaths)
}
