package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"windmusic/internal/music"
	"windmusic/internal/sourcemgr"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	sources *sourcemgr.Manager
}

func (a *App) sourceDisplayName(sourceID string) string {
	if a.sources == nil {
		return sourceID
	}
	return a.sources.SourceDisplayName(sourceID)
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	rootDir, err := sourcemgr.DefaultRootDir()
	if err != nil {
		fmt.Println("解析音源目录失败:", err)
		return
	}

	manager, err := sourcemgr.NewManager(rootDir)
	if err != nil {
		fmt.Println("初始化音源管理器失败:", err)
		return
	}
	a.sources = manager
}

func (a *App) ensureSources() error {
	if a.sources == nil {
		return fmt.Errorf("音源管理器未初始化")
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
		return music.SourceInfo{}, fmt.Errorf("已取消导入")
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
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	log.Printf("[后端] 开始搜索 source=%s platform=%s page=%d keyword=%q", source, platform, page, keyword)

	if err := a.ensureSources(); err != nil {
		log.Printf("[后端] 搜索失败 source=%s err=%v elapsed=%s", source, err, time.Since(startedAt))
		return nil, err
	}

	result, err := a.sources.Search(sourceID, platform, keyword, page)
	if err != nil {
		log.Printf("[后端] 搜索失败 source=%s platform=%s err=%v elapsed=%s", source, platform, err, time.Since(startedAt))
		return nil, err
	}

	log.Printf("[后端] 搜索完成 source=%s platform=%s total=%d list=%d elapsed=%s", source, result.Source, result.Total, len(result.List), time.Since(startedAt))
	return result, nil
}

func (a *App) GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	log.Printf("[后端] 开始获取播放地址 source=%s platform=%s quality=%s metaBytes=%d", source, platform, quality, len(metaJSON))

	if err := a.ensureSources(); err != nil {
		log.Printf("[后端] 获取播放地址失败 source=%s err=%v elapsed=%s", source, err, time.Since(startedAt))
		return "", err
	}

	url, err := a.sources.GetMusicURL(sourceID, platform, quality, metaJSON)
	if err != nil {
		log.Printf("[后端] 获取播放地址失败 source=%s platform=%s quality=%s err=%v elapsed=%s", source, platform, quality, err, time.Since(startedAt))
		return "", err
	}

	log.Printf("[后端] 获取播放地址完成 source=%s platform=%s quality=%s urlBytes=%d elapsed=%s musicUrl=%s", source, platform, quality, len(url), time.Since(startedAt), url)
	return url, nil
}

func (a *App) GetLyric(sourceID, platform, metaJSON string) (*music.LyricInfo, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	log.Printf("[后端] 开始获取歌词 source=%s platform=%s metaBytes=%d", source, platform, len(metaJSON))

	if err := a.ensureSources(); err != nil {
		log.Printf("[后端] 获取歌词失败 source=%s err=%v elapsed=%s", source, err, time.Since(startedAt))
		return nil, err
	}

	lyric, err := a.sources.GetLyric(sourceID, platform, metaJSON)
	if err != nil {
		log.Printf("[后端] 获取歌词失败 source=%s platform=%s err=%v elapsed=%s", source, platform, err, time.Since(startedAt))
		return nil, err
	}

	log.Printf("[后端] 获取歌词完成 source=%s platform=%s lyricBytes=%d elapsed=%s", source, platform, len(lyric.Lyric), time.Since(startedAt))
	return lyric, nil
}

func (a *App) GetPic(sourceID, platform, metaJSON string) (string, error) {
	startedAt := time.Now()
	source := a.sourceDisplayName(sourceID)
	log.Printf("[后端] 开始获取封面 source=%s platform=%s metaBytes=%d", source, platform, len(metaJSON))

	if err := a.ensureSources(); err != nil {
		log.Printf("[后端] 获取封面失败 source=%s err=%v elapsed=%s", source, err, time.Since(startedAt))
		return "", err
	}

	picURL, err := a.sources.GetPic(sourceID, platform, metaJSON)
	if err != nil {
		log.Printf("[后端] 获取封面失败 source=%s platform=%s err=%v elapsed=%s", source, platform, err, time.Since(startedAt))
		return "", err
	}

	log.Printf("[后端] 获取封面完成 source=%s platform=%s urlBytes=%d elapsed=%s", source, platform, len(picURL), time.Since(startedAt))
	return picURL, nil
}

func (a *App) GetSourceDataDir() (string, error) {
	rootDir, err := sourcemgr.DefaultRootDir()
	if err != nil {
		return "", err
	}
	return rootDir, nil
}
