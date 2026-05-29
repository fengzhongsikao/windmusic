package meting

import (
	"log"
	"time"

	models "windmusic/internal/music"
	"windmusic/internal/musicsearch"
)

func Search(sourceID, platform, keyword string, page int) (*models.SearchResult, error) {
	startedAt := time.Now()
	source := SourceDisplayName(sourceID)
	logPrefix := BackendLogPrefix(sourceID)
	log.Printf("%s 开始搜索 source=%s platform=%s page=%d keyword=%q", logPrefix, source, platform, page, keyword)

	base := ResolveMetingBase(sourceID)
	result, err := musicsearch.SearchMeting(base, platform, keyword, page, 20)
	if err != nil {
		log.Printf("%s 搜索失败 source=%s platform=%s err=%v elapsed=%s", logPrefix, source, platform, err, time.Since(startedAt))
		return nil, err
	}
	log.Printf("%s 搜索完成 source=%s platform=%s total=%d list=%d elapsed=%s", logPrefix, source, result.Source, result.Total, len(result.List), time.Since(startedAt))
	return result, nil
}

func GetMusicURL(sourceID, platform, quality, metaJSON string) (string, error) {
	startedAt := time.Now()
	source := SourceDisplayName(sourceID)
	logPrefix := BackendLogPrefix(sourceID)
	log.Printf("%s 开始获取播放地址 source=%s platform=%s quality=%s metaBytes=%d", logPrefix, source, platform, quality, len(metaJSON))

	url, err := musicsearch.GetMetingMusicURL(metaJSON)
	if err != nil {
		log.Printf("%s 获取播放地址失败 source=%s platform=%s quality=%s err=%v elapsed=%s", logPrefix, source, platform, quality, err, time.Since(startedAt))
		return "", err
	}
	log.Printf("%s 获取播放地址完成 source=%s platform=%s quality=%s urlBytes=%d elapsed=%s musicUrl=%s", logPrefix, source, platform, quality, len(url), time.Since(startedAt), url)
	return url, nil
}

func GetLyric(sourceID, platform, metaJSON string) (*models.LyricInfo, error) {
	startedAt := time.Now()
	source := SourceDisplayName(sourceID)
	logPrefix := BackendLogPrefix(sourceID)
	log.Printf("%s 开始获取歌词 source=%s platform=%s metaBytes=%d", logPrefix, source, platform, len(metaJSON))

	lyric, err := musicsearch.GetMetingLyric(metaJSON)
	if err != nil {
		log.Printf("%s 获取歌词失败 source=%s platform=%s err=%v elapsed=%s", logPrefix, source, platform, err, time.Since(startedAt))
		return nil, err
	}
	log.Printf("%s 获取歌词完成 source=%s platform=%s lyricBytes=%d elapsed=%s", logPrefix, source, platform, len(lyric.Lyric), time.Since(startedAt))
	return lyric, nil
}

func GetPic(sourceID, platform, metaJSON string) (string, error) {
	startedAt := time.Now()
	source := SourceDisplayName(sourceID)
	logPrefix := BackendLogPrefix(sourceID)
	log.Printf("%s 开始获取封面 source=%s platform=%s metaBytes=%d", logPrefix, source, platform, len(metaJSON))

	picURL, err := musicsearch.GetMetingPic(metaJSON)
	if err != nil {
		log.Printf("%s 获取封面失败 source=%s platform=%s err=%v elapsed=%s", logPrefix, source, platform, err, time.Since(startedAt))
		return "", err
	}
	log.Printf("%s 获取封面完成 source=%s platform=%s urlBytes=%d elapsed=%s", logPrefix, source, platform, len(picURL), time.Since(startedAt))
	return picURL, nil
}
