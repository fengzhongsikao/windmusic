package main

import models "windmusic/internal/music"

func (a *App) GetPlayerSettings() (models.PlayerSettings, error) {
	return a.player.Get()
}

func (a *App) UpdatePlayerSettings(settings models.PlayerSettings) error {
	if err := a.player.Save(settings); err != nil {
		return err
	}
	a.emitPlayerSettings()
	return nil
}

func (a *App) GetMetingSettings() (models.MetingSettings, error) {
	return a.meting.Get()
}

func (a *App) AddMetingURL(url string) error {
	if _, err := a.meting.AddURL(url); err != nil {
		return err
	}
	a.emitMetingSettings()
	return nil
}

func (a *App) RemoveMetingURL(url string) error {
	if _, err := a.meting.RemoveURL(url); err != nil {
		return err
	}
	a.emitMetingSettings()
	return nil
}

func (a *App) SetActiveMetingURL(url string) error {
	if _, err := a.meting.SetActiveURL(url); err != nil {
		return err
	}
	a.emitMetingSettings()
	return nil
}

func (a *App) SetMetingPlatform(platform string) error {
	if _, err := a.meting.SetPlatform(platform); err != nil {
		return err
	}
	a.emitMetingSettings()
	return nil
}

func (a *App) GetDiscoverRecommendCache(tabID string) (models.DiscoverRecommendCache, error) {
	return a.discover.Get(tabID)
}

func (a *App) SetDiscoverRecommendCache(tabID string, songs []models.SongItem) error {
	return a.discover.Set(tabID, songs)
}

