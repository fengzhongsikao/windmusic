package main

import (
	models "windmusic/internal/music"
	"windmusic/music/events"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) bootstrapAppData() {
	a.emitPlayerSettings()
	a.emitMetingSettings()
	a.emitFavorites()
	a.emitPlaylists()
	a.emitRecent()
}

func (a *App) emitPlayerSettings() {
	if a.ctx == nil {
		return
	}
	settings, err := a.player.Get()
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, events.EventPlayerSettingsUpdated, settings)
}

func (a *App) emitMetingSettings() {
	if a.ctx == nil {
		return
	}
	settings, err := a.meting.Get()
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, events.EventMetingSettingsUpdated, settings)
}

func (a *App) emitFavorites() {
	if a.ctx == nil {
		return
	}
	items, err := a.favorites.List()
	if err != nil {
		return
	}
	if items == nil {
		items = []models.FavoriteSong{}
	}
	runtime.EventsEmit(a.ctx, events.EventFavoritesUpdated, items)
}

func (a *App) emitRecent() {
	if a.ctx == nil {
		return
	}
	items, err := a.recent.List()
	if err != nil {
		return
	}
	if items == nil {
		items = []models.RecentSong{}
	}
	runtime.EventsEmit(a.ctx, events.EventRecentUpdated, items)
}

func (a *App) emitPlaylists() {
	if a.ctx == nil {
		return
	}
	items, err := a.playlists.List()
	if err != nil {
		return
	}
	if items == nil {
		items = []models.UserPlaylist{}
	}
	runtime.EventsEmit(a.ctx, events.EventPlaylistsUpdated, items)
}
