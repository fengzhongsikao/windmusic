package main

import (
	models "windmusic/internal/music"
)

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
