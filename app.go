package main

import (
	"context"

	models "windmusic/internal/music"
	appmusic "windmusic/music"
)

type App struct {
	ctx       context.Context
	favorites appmusic.FavoritesStore
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
