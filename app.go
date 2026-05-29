package main

import (
	"context"

	models "windmusic/internal/music"
	appmusic "windmusic/music"
)

type App struct {
	ctx       context.Context
	favorites appmusic.FavoritesStore
	recent    appmusic.RecentStore
	playlists appmusic.PlaylistsStore
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Search(sourceID, platform, keyword string, page int) (*models.SearchResult, error) {
	return appmusic.Search(sourceID, platform, keyword, page)
}

func (a *App) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	return appmusic.GetMusicURL(sourceID, platform, quality, metaJSON)
}

func (a *App) GetLyric(sourceID, platform, metaJSON string) (*models.LyricInfo, error) {
	return appmusic.GetLyric(sourceID, platform, metaJSON)
}

func (a *App) GetPic(sourceID, platform, metaJSON string) (string, error) {
	return appmusic.GetPic(sourceID, platform, metaJSON)
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
	return a.favorites.Add(song)
}

func (a *App) RemoveFavorite(song models.FavoriteSong) error {
	return a.favorites.Remove(song)
}

func (a *App) ListRecent() ([]models.RecentSong, error) {
	return a.recent.List()
}

func (a *App) RecordRecent(song models.RecentSong) error {
	return a.recent.Record(song)
}

func (a *App) RemoveRecent(song models.RecentSong) error {
	return a.recent.Remove(song)
}

func (a *App) ClearRecent() error {
	return a.recent.Clear()
}

func (a *App) ListPlaylists() ([]models.UserPlaylist, error) {
	return a.playlists.List()
}

func (a *App) CreatePlaylist(name string) (models.UserPlaylist, error) {
	return a.playlists.Create(name)
}

func (a *App) GetPlaylist(id string) (models.UserPlaylist, error) {
	return a.playlists.Get(id)
}

func (a *App) AddPlaylistSong(playlistID string, song models.FavoriteSong) error {
	return a.playlists.AddSong(playlistID, song)
}

func (a *App) RemovePlaylistSong(playlistID string, song models.FavoriteSong) error {
	return a.playlists.RemoveSong(playlistID, song)
}

func (a *App) DeletePlaylist(id string) error {
	return a.playlists.Delete(id)
}
