package musicsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"windmusic/internal/music"
)

var client = &http.Client{Timeout: 20 * time.Second}

const (
	defaultMetingBase = "https://meting.mikus.ink"
	defaultUserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// metingItem 表示 Meting API 搜索返回的单条歌曲数据（兼容 title/author 与 name/artist）。
type metingItem struct {
	Title  string `json:"title"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Artist string `json:"artist"`
	Pic    string `json:"pic"`
	URL    string `json:"url"`
	Lrc    string `json:"lrc"`
}

func (item metingItem) songTitle() string {
	if title := strings.TrimSpace(item.Title); title != "" {
		return title
	}
	return strings.TrimSpace(item.Name)
}

func (item metingItem) songAuthor() string {
	if author := strings.TrimSpace(item.Author); author != "" {
		return author
	}
	return strings.TrimSpace(item.Artist)
}

// metingTrack 搜索结果的持久化元数据（与 metingItem 字段一致，另含 id/server）。
type metingTrack struct {
	metingItem
	ID     string `json:"id"`
	Server string `json:"server"`
}

// SearchMeting 调用 Meting API 搜索歌曲，返回分页后的搜索结果。
// platform 为前端固定的 tencent 或 netease。
func SearchMeting(baseURL, platform, keyword string, page, limit int) (*music.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	server, err := metingAPIServer(platform)
	if err != nil {
		return nil, err
	}
	if err := validateMetingSearch(server); err != nil {
		return nil, err
	}

	items, err := searchMeting(normalizeMetingBase(baseURL), server, keyword)
	if err != nil {
		return nil, err
	}

	start := (page - 1) * limit
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	list := make([]music.SongItem, 0, end-start)
	for idx, item := range items[start:end] {
		title := item.songTitle()
		author := item.songAuthor()
		id := idFromMetingURL(item.URL)
		if id == "" {
			id = idFromMetingURL(item.Lrc)
		}
		if id == "" {
			id = fmt.Sprintf("meting:%s:%s:%d", title, author, start+idx)
		}

		metaJSON, _ := json.Marshal(metingTrack{
			metingItem: metingItem{
				Title:  title,
				Author: author,
				Pic:    strings.TrimSpace(item.Pic),
				URL:    strings.TrimSpace(item.URL),
				Lrc:    strings.TrimSpace(item.Lrc),
			},
			ID:     id,
			Server: server,
		})

		list = append(list, music.SongItem{
			ID:       id,
			Name:     title,
			Singer:   author,
			Source:   platform,
			Img:      strings.TrimSpace(item.Pic),
			SongMID:  id,
			MetaJSON: string(metaJSON),
		})
	}

	return &music.SearchResult{
		List:   list,
		Total:  len(items),
		Page:   page,
		Limit:  limit,
		Source: platform,
	}, nil
}

// GetMetingMusicURL 返回搜索时写入 metaJSON 的 url 字段（Meting 播放地址）。
func GetMetingMusicURL(metaJSON string) (string, error) {
	meta := parseMetingTrack(metaJSON)
	if u := strings.TrimSpace(meta.URL); u != "" {
		return u, nil
	}
	return "", fmt.Errorf("meting meta missing url")
}

// GetMetingPic 返回搜索时写入 metaJSON 的 pic 字段。
func GetMetingPic(metaJSON string) (string, error) {
	meta := parseMetingTrack(metaJSON)
	if pic := strings.TrimSpace(meta.Pic); pic != "" {
		return pic, nil
	}
	return "", fmt.Errorf("meting meta missing pic")
}

// GetMetingLyric 请求 metaJSON 中的 lrc 地址并返回歌词文本。
func GetMetingLyric(metaJSON string) (*music.LyricInfo, error) {
	lrcURL := strings.TrimSpace(parseMetingTrack(metaJSON).Lrc)
	if lrcURL == "" {
		return nil, fmt.Errorf("meting meta missing lrc")
	}
	text, err := httpGetText(lrcURL)
	if err != nil {
		return nil, err
	}
	return &music.LyricInfo{Lyric: strings.TrimSpace(text)}, nil
}

func parseMetingTrack(metaJSON string) metingTrack {
	meta := metingTrack{}
	_ = json.Unmarshal([]byte(metaJSON), &meta)
	return meta
}

// idFromMetingURL 从 Meting API 返回的 URL 中提取 id 查询参数。
func idFromMetingURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Query().Get("id"))
}

// metingAPIServer 将前端平台标识映射为 Meting API 的 server 参数。
func metingAPIServer(platform string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "", "tx", "tencent", "qq":
		return "tencent", nil
	case "wy", "netease", "163":
		return "netease", nil
	default:
		return "", fmt.Errorf("unsupported meting platform: %s", platform)
	}
}

// validateMetingSearch 校验 Meting 节点是否支持该平台的搜索（mikus 节点仅 netease 支持 search）。
func validateMetingSearch(server string) error {
	if server == "netease" {
		return nil
	}
	return fmt.Errorf("meting search is not supported for %s (QQ音乐仅支持单曲/歌单，请切换到网易云)", server)
}

// searchMeting 向 Meting API 发送搜索请求，返回原始歌曲列表。
func searchMeting(baseURL, server, keyword string) ([]metingItem, error) {
	idURL := fmt.Sprintf(
		"%s/api?server=%s&type=search&id=%s",
		baseURL,
		url.QueryEscape(server),
		url.QueryEscape(keyword),
	)
	items, err := httpGetJSON[[]metingItem](idURL)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}

	sURL := fmt.Sprintf(
		"%s/api?server=%s&type=search&s=%s",
		baseURL,
		url.QueryEscape(server),
		url.QueryEscape(keyword),
	)
	return httpGetJSON[[]metingItem](sURL)
}

// normalizeMetingBase 规范化 Meting API 基础地址，去除尾部斜杠，空值时返回默认节点。
func normalizeMetingBase(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return defaultMetingBase
	}
	return raw
}

// httpGetJSON 发送 GET 请求并将响应体反序列化为指定类型 T。
func httpGetJSON[T any](rawURL string) (T, error) {
	var zero T
	data, err := httpGet(rawURL)
	if err != nil {
		return zero, err
	}
	if err := json.Unmarshal(data, &zero); err != nil {
		return zero, fmt.Errorf("decode meting response failed: %w", err)
	}
	return zero, nil
}

// httpGetText 发送 GET 请求并返回响应体的纯文本内容。
func httpGetText(rawURL string) (string, error) {
	data, err := httpGet(rawURL)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// httpGet 执行底层 HTTP GET 请求，设置 UA 和 Accept 头，返回响应字节。
func httpGet(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("meting request failed: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
