package main

import (
	"context"
	"fmt"

	"windmusic/internal/music"
	"windmusic/internal/sourcemgr"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	sources *sourcemgr.Manager
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	rootDir, err := sourcemgr.DefaultRootDir()
	if err != nil {
		fmt.Println("failed to resolve source directory:", err)
		return
	}

	manager, err := sourcemgr.NewManager(rootDir)
	if err != nil {
		fmt.Println("failed to initialize source manager:", err)
		return
	}
	a.sources = manager
}

func (a *App) ensureSources() error {
	if a.sources == nil {
		return fmt.Errorf("source manager not initialized")
	}
	return nil
}

func (a *App) ImportSource() (music.SourceInfo, error) {
	if err := a.ensureSources(); err != nil {
		return music.SourceInfo{}, err
	}

	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入音源",
		Filters: []runtime.FileFilter{
			{DisplayName: "JavaScript 音源 (*.js)", Pattern: "*.js"},
		},
	})
	if err != nil {
		return music.SourceInfo{}, err
	}
	if path == "" {
		return music.SourceInfo{}, fmt.Errorf("import cancelled")
	}
	return a.sources.ImportSource(path)
}

func (a *App) ListSources() ([]music.SourceInfo, error) {
	if err := a.ensureSources(); err != nil {
		return nil, err
	}
	return a.sources.ListSources(), nil
}

func (a *App) EnableSource(sourceID string) (music.SourceInfo, error) {
	if err := a.ensureSources(); err != nil {
		return music.SourceInfo{}, err
	}
	return a.sources.EnableSource(sourceID)
}

func (a *App) DisableSource(sourceID string) (music.SourceInfo, error) {
	if err := a.ensureSources(); err != nil {
		return music.SourceInfo{}, err
	}
	return a.sources.DisableSource(sourceID)
}

func (a *App) DeleteSource(sourceID string) error {
	if err := a.ensureSources(); err != nil {
		return err
	}
	return a.sources.DeleteSource(sourceID)
}

func (a *App) Search(sourceID, platform, keyword string, page int) (*music.SearchResult, error) {
	if err := a.ensureSources(); err != nil {
		return nil, err
	}
	return a.sources.Search(sourceID, platform, keyword, page)
}

func (a *App) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	if err := a.ensureSources(); err != nil {
		return "", err
	}
	return a.sources.GetMusicURL(sourceID, platform, quality, metaJSON)
}

func (a *App) GetLyric(sourceID, platform, metaJSON string) (*music.LyricInfo, error) {
	if err := a.ensureSources(); err != nil {
		return nil, err
	}
	return a.sources.GetLyric(sourceID, platform, metaJSON)
}

func (a *App) GetPic(sourceID, platform, metaJSON string) (string, error) {
	if err := a.ensureSources(); err != nil {
		return "", err
	}
	return a.sources.GetPic(sourceID, platform, metaJSON)
}

func (a *App) GetSourceDataDir() (string, error) {
	rootDir, err := sourcemgr.DefaultRootDir()
	if err != nil {
		return "", err
	}
	return rootDir, nil
}
