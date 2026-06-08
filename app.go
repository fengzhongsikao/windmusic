package main

import (
	"context"
	"sync"

	models "windmusic/internal/music"
	"windmusic/music/appdata"
	"windmusic/music/cache"
	"windmusic/music/persist"
	localmusic "windmusic/music/local"
	metingmusic "windmusic/music/meting"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	favorites  persist.FavoritesStore
	recent     persist.RecentStore
	playlists  persist.PlaylistsStore
	local      localmusic.LocalLibraryStore
	localAudio *localmusic.AudioServer
	player     persist.PlayerSettingsStore
	meting     persist.MetingSettingsStore
	discover    *cache.DiscoverCache
	lyricCache  *cache.LyricCache
	searchCache *cache.SearchCache
	musicURLCache *cache.MusicURLCache

	localScanMu      sync.Mutex
	localScanRunning bool
	localScanPending bool
}

func NewApp() *App {
	app := &App{
		discover:      cache.NewDiscoverCache(),
		lyricCache:    cache.NewLyricCache(),
		searchCache:   cache.NewSearchCache(),
		musicURLCache: cache.NewMusicURLCache(),
	}
	app.localAudio = localmusic.NewAudioServer(&app.local)
	app.local.SetCoverURLBuilder(func(coverKey string) string {
		url, err := app.localAudio.CoverURL(coverKey)
		if err != nil {
			return ""
		}
		return url
	})
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.bootstrapAppData()
	go a.bootstrapLocalLibrary()
}

func (a *App) Search(sourceID, platform, keyword string, page int) (*models.SearchResult, error) {
	return metingmusic.Search(sourceID, platform, keyword, page, a.searchCache)
}

func (a *App) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	if url, ok := a.musicURLCache.Get(sourceID, platform, quality, metaJSON); ok {
		return url, nil
	}
	url, err := metingmusic.GetMusicURL(sourceID, platform, quality, metaJSON)
	if err != nil {
		return "", err
	}
	a.musicURLCache.Set(sourceID, platform, quality, metaJSON, url)
	return url, nil
}

func (a *App) GetLyric(sourceID, platform, metaJSON string) (*models.LyricInfo, error) {
	if lyric, ok := a.lyricCache.Get(metaJSON); ok {
		return lyric, nil
	}
	lyric, err := metingmusic.GetLyric(sourceID, platform, metaJSON)
	if err != nil {
		return nil, err
	}
	a.lyricCache.Set(metaJSON, lyric)
	return lyric, nil
}

func (a *App) GetPic(sourceID, platform, metaJSON string) (string, error) {
	return metingmusic.GetPic(sourceID, platform, metaJSON)
}

func (a *App) GetSourceDataDir() (string, error) {
	return appdata.AppDataRootDir()
}

func (a *App) localLibrarySnapshot() (models.LocalLibrarySnapshot, error) {
	return a.local.Snapshot()
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
