package main

import (
	"context"

	models "windmusic/internal/music"
	appmusic "windmusic/music"
	localmusic "windmusic/music/local"
	metingmusic "windmusic/music/meting"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	favorites appmusic.FavoritesStore
	recent    appmusic.RecentStore
	playlists appmusic.PlaylistsStore
	local     localmusic.LocalLibraryStore
	player    appmusic.PlayerSettingsStore
	meting    appmusic.MetingSettingsStore
	discover  *appmusic.DiscoverCache
}

func NewApp() *App {
	return &App{
		discover: appmusic.NewDiscoverCache(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.bootstrapAppData()
	go a.bootstrapLocalLibrary()
}

func (a *App) localLibrarySnapshot() (models.LocalLibrarySnapshot, error) {
	folders, err := a.local.ListFolders()
	if err != nil {
		return models.LocalLibrarySnapshot{}, err
	}
	songs, err := a.local.ListCached()
	if err != nil {
		return models.LocalLibrarySnapshot{}, err
	}
	if songs == nil {
		songs = []models.LocalSong{}
	}
	if folders == nil {
		folders = []string{}
	}
	return models.LocalLibrarySnapshot{Folders: folders, Songs: songs}, nil
}

func (a *App) emitLocalLibrarySnapshot() {
	if a.ctx == nil {
		return
	}
	snapshot, err := a.localLibrarySnapshot()
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, localmusic.EventLibraryUpdated, snapshot)
}

func (a *App) emitLocalLibraryScanning(scanning bool) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, localmusic.EventLibraryScanning, scanning)
}

func (a *App) bootstrapLocalLibrary() {
	a.emitLocalLibrarySnapshot()

	snapshot, err := a.localLibrarySnapshot()
	if err != nil {
		return
	}
	if len(snapshot.Folders) == 0 || len(snapshot.Songs) > 0 {
		return
	}

	a.emitLocalLibraryScanning(true)
	defer a.emitLocalLibraryScanning(false)

	if _, err := a.local.Scan(); err != nil {
		return
	}
	a.emitLocalLibrarySnapshot()
}

func (a *App) scanAndEmitLocalLibrary() error {
	a.emitLocalLibraryScanning(true)
	defer a.emitLocalLibraryScanning(false)

	if _, err := a.local.Scan(); err != nil {
		return err
	}
	a.emitLocalLibrarySnapshot()
	return nil
}

func (a *App) Search(sourceID, platform, keyword string, page int) (*models.SearchResult, error) {
	return metingmusic.Search(sourceID, platform, keyword, page)
}

func (a *App) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	return metingmusic.GetMusicURL(sourceID, platform, quality, metaJSON)
}

func (a *App) GetLyric(sourceID, platform, metaJSON string) (*models.LyricInfo, error) {
	return metingmusic.GetLyric(sourceID, platform, metaJSON)
}

func (a *App) GetPic(sourceID, platform, metaJSON string) (string, error) {
	return metingmusic.GetPic(sourceID, platform, metaJSON)
}

func (a *App) GetSourceDataDir() (string, error) {
	return appmusic.AppDataRootDir()
}

func (a *App) ListFavorites() ([]models.FavoriteSong, error) {
	return a.favorites.List()
}

func (a *App) IsFavorite(song models.FavoriteSong) (bool, error) {
	return a.favorites.IsFavorite(song)
}

func (a *App) AddFavorite(song models.FavoriteSong) error {
	if err := a.favorites.Add(song); err != nil {
		return err
	}
	a.emitFavorites()
	return nil
}

func (a *App) RemoveFavorite(song models.FavoriteSong) error {
	if err := a.favorites.Remove(song); err != nil {
		return err
	}
	a.emitFavorites()
	return nil
}

func (a *App) ListRecent() ([]models.RecentSong, error) {
	return a.recent.List()
}

func (a *App) RecordRecent(song models.RecentSong) error {
	if err := a.recent.Record(song); err != nil {
		return err
	}
	a.emitRecent()
	return nil
}

func (a *App) RemoveRecent(song models.RecentSong) error {
	if err := a.recent.Remove(song); err != nil {
		return err
	}
	a.emitRecent()
	return nil
}

func (a *App) ClearRecent() error {
	if err := a.recent.Clear(); err != nil {
		return err
	}
	a.emitRecent()
	return nil
}

func (a *App) ListPlaylists() ([]models.UserPlaylist, error) {
	return a.playlists.List()
}

func (a *App) CreatePlaylist(name string) (models.UserPlaylist, error) {
	playlist, err := a.playlists.Create(name)
	if err != nil {
		return models.UserPlaylist{}, err
	}
	a.emitPlaylists()
	return playlist, nil
}

func (a *App) GetPlaylist(id string) (models.UserPlaylist, error) {
	return a.playlists.Get(id)
}

func (a *App) AddPlaylistSong(playlistID string, song models.FavoriteSong) error {
	if err := a.playlists.AddSong(playlistID, song); err != nil {
		return err
	}
	a.emitPlaylists()
	return nil
}

func (a *App) RemovePlaylistSong(playlistID string, song models.FavoriteSong) error {
	if err := a.playlists.RemoveSong(playlistID, song); err != nil {
		return err
	}
	a.emitPlaylists()
	return nil
}

func (a *App) DeletePlaylist(id string) error {
	if err := a.playlists.Delete(id); err != nil {
		return err
	}
	a.emitPlaylists()
	return nil
}

func (a *App) PickLocalMusicFolder() (string, error) {
	dir, err := localmusic.PickMusicFolder(a.ctx)
	if err != nil || dir == "" {
		return dir, err
	}
	if err := a.local.AddFolder(dir); err != nil {
		return "", err
	}
	if err := a.scanAndEmitLocalLibrary(); err != nil {
		return dir, err
	}
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
	return a.scanAndEmitLocalLibrary()
}

func (a *App) ListLocalLibrary() ([]models.LocalSong, error) {
	return a.local.ListCached()
}

func (a *App) ScanLocalLibrary() ([]models.LocalSong, error) {
	a.emitLocalLibraryScanning(true)
	defer a.emitLocalLibraryScanning(false)

	songs, err := a.local.Scan()
	if err != nil {
		return nil, err
	}
	a.emitLocalLibrarySnapshot()
	return songs, nil
}

func (a *App) GetLocalAudioStream(filePath string) (string, error) {
	return localmusic.GetLocalAudioStream(&a.local, filePath)
}

func (a *App) GetLocalSongExtras(filePath string) (models.LocalSongExtras, error) {
	return a.local.GetSongExtras(filePath)
}

func (a *App) GetLocalSongCovers(filePaths []string) (models.LocalCoverBatch, error) {
	return a.local.GetCovers(filePaths)
}
