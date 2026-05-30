package main

import (
	"context"

	models "windmusic/internal/music"
	"windmusic/music/appdata"
	"windmusic/music/cache"
	"windmusic/music/persist"
	localmusic "windmusic/music/local"
	metingmusic "windmusic/music/meting"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	favorites persist.FavoritesStore
	recent    persist.RecentStore
	playlists persist.PlaylistsStore
	local     localmusic.LocalLibraryStore
	player    persist.PlayerSettingsStore
	meting    persist.MetingSettingsStore
	discover  *cache.DiscoverCache
}

func NewApp() *App {
	return &App{
		discover: cache.NewDiscoverCache(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.bootstrapAppData()
	go a.bootstrapLocalLibrary()
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
	return appdata.AppDataRootDir()
}

func (a *App) localLibrarySnapshot() (models.LocalLibrarySnapshot, error) {
	folders, err := a.local.ListFolders()
	if err != nil {
		return models.LocalLibrarySnapshot{}, err
	}
	if folders == nil {
		folders = []string{}
	}
	counts, err := a.local.FolderCounts()
	if err != nil {
		return models.LocalLibrarySnapshot{}, err
	}
	if counts == nil {
		counts = map[string]int{}
	}
	aliases, err := a.local.ListFolderAliases()
	if err != nil {
		return models.LocalLibrarySnapshot{}, err
	}
	if aliases == nil {
		aliases = map[string]string{}
	}
	return models.LocalLibrarySnapshot{
		Folders:       folders,
		FolderAliases: aliases,
		FolderCounts:  counts,
	}, nil
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
