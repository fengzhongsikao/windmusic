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

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" {
		return ""
	}
	return text
}

func SearchMeting(baseURL, platform, keyword string, page, limit int) (*music.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	server := metingServer(platform)
	endpoint := fmt.Sprintf("%s/api?server=%s&type=search&s=%s", normalizeMetingBase(baseURL), url.QueryEscape(server), url.QueryEscape(keyword))
	body, err := getJSONList(endpoint)
	if err != nil {
		return nil, err
	}

	start := (page - 1) * limit
	if start > len(body) {
		start = len(body)
	}
	end := start + limit
	if end > len(body) {
		end = len(body)
	}

	list := make([]music.SongItem, 0, end-start)
	for _, row := range body[start:end] {
		id := strings.TrimSpace(stringValue(row["id"]))
		// Meting 标准字段：name / artist；同时兼容历史 title / author
		title := strings.TrimSpace(stringValue(row["name"]))
		if title == "" {
			title = strings.TrimSpace(stringValue(row["title"]))
		}
		author := strings.TrimSpace(stringValue(row["artist"]))
		if author == "" {
			author = strings.TrimSpace(stringValue(row["author"]))
		}
		metaObj := map[string]interface{}{
			"id":      id,
			"name":    title,
			"artist":  author,
			"pic":     strings.TrimSpace(stringValue(row["pic"])),
			"url":     strings.TrimSpace(stringValue(row["url"])),
			"lrc":     strings.TrimSpace(stringValue(row["lrc"])),
			"server":  server,
			"meting":  normalizeMetingBase(baseURL),
			"rawItem": row,
		}
		metaJSON, _ := json.Marshal(metaObj)
		list = append(list, music.SongItem{
			ID:       id,
			Name:     title,
			Singer:   author,
			Album:    "",
			Source:   platform,
			Interval: "",
			Img:      strings.TrimSpace(stringValue(row["pic"])),
			SongMID:  id,
			MetaJSON: string(metaJSON),
		})
	}

	return &music.SearchResult{
		List:   list,
		Total:  len(body),
		Page:   page,
		Limit:  limit,
		Source: platform,
	}, nil
}

func GetMetingMusicURL(baseURL, platform, metaJSON string) (string, error) {
	meta := parseMetingMeta(metaJSON)
	if raw := strings.TrimSpace(stringValue(meta["url"])); raw != "" {
		return raw, nil
	}
	id, resolved := resolveMetingID(baseURL, platform, meta)
	if id == "" {
		return "", fmt.Errorf("meting meta missing id")
	}
	if raw := strings.TrimSpace(stringValue(resolved["url"])); raw != "" {
		return raw, nil
	}
	server := metingServer(platform)
	return fmt.Sprintf("%s/api?server=%s&type=url&id=%s", normalizeMetingBase(baseURL), url.QueryEscape(server), url.QueryEscape(id)), nil
}

func GetMetingPic(baseURL, platform, metaJSON string) (string, error) {
	meta := parseMetingMeta(metaJSON)
	if raw := strings.TrimSpace(stringValue(meta["pic"])); raw != "" {
		return raw, nil
	}
	id, resolved := resolveMetingID(baseURL, platform, meta)
	if id == "" {
		return "", fmt.Errorf("meting meta missing id")
	}
	if raw := strings.TrimSpace(stringValue(resolved["pic"])); raw != "" {
		return raw, nil
	}
	server := metingServer(platform)
	return fmt.Sprintf("%s/api?server=%s&type=pic&id=%s", normalizeMetingBase(baseURL), url.QueryEscape(server), url.QueryEscape(id)), nil
}

func GetMetingLyric(baseURL, platform, metaJSON string) (*music.LyricInfo, error) {
	meta := parseMetingMeta(metaJSON)
	lrcURL := strings.TrimSpace(stringValue(meta["lrc"]))
	if lrcURL == "" {
		id, resolved := resolveMetingID(baseURL, platform, meta)
		if lrcURL == "" {
			lrcURL = strings.TrimSpace(stringValue(resolved["lrc"]))
		}
		if id == "" {
			return nil, fmt.Errorf("meting meta missing lrc/id")
		}
		if lrcURL == "" {
			server := metingServer(platform)
			lrcURL = fmt.Sprintf("%s/api?server=%s&type=lrc&id=%s", normalizeMetingBase(baseURL), url.QueryEscape(server), url.QueryEscape(id))
		}
	}

	text, err := getText(lrcURL)
	if err != nil {
		return nil, err
	}
	return &music.LyricInfo{Lyric: strings.TrimSpace(text)}, nil
}

func resolveMetingID(baseURL, platform string, meta map[string]interface{}) (string, map[string]interface{}) {
	if id := strings.TrimSpace(stringValue(meta["id"])); id != "" {
		return id, meta
	}

	name := extractMetaSongName(meta)
	artist := extractMetaArtist(meta)
	keyword := strings.TrimSpace(strings.Join([]string{name, artist}, " "))
	if keyword == "" {
		return "", meta
	}

	server := metingServer(platform)
	searchURL := fmt.Sprintf("%s/api?server=%s&type=search&s=%s", normalizeMetingBase(baseURL), url.QueryEscape(server), url.QueryEscape(keyword))
	items, err := getJSONList(searchURL)
	if err != nil || len(items) == 0 {
		return "", meta
	}
	best := items[0]
	if id := strings.TrimSpace(stringValue(best["id"])); id != "" {
		return id, best
	}
	return "", meta
}

func extractMetaSongName(meta map[string]interface{}) string {
	for _, key := range []string{"name", "songname", "SONGNAME", "title"} {
		if v := strings.TrimSpace(stringValue(meta[key])); v != "" {
			return v
		}
	}
	return ""
}

func extractMetaArtist(meta map[string]interface{}) string {
	for _, key := range []string{"artist", "author", "singer", "singerName", "ARTIST"} {
		if v := strings.TrimSpace(stringValue(meta[key])); v != "" {
			return v
		}
	}
	if rawArtists, ok := meta["artists"].([]interface{}); ok {
		names := make([]string, 0, len(rawArtists))
		for _, item := range rawArtists {
			if artistMap, ok := item.(map[string]interface{}); ok {
				if name := strings.TrimSpace(stringValue(artistMap["name"])); name != "" {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, " ")
	}
	if rawSingers, ok := meta["singer"].([]interface{}); ok {
		names := make([]string, 0, len(rawSingers))
		for _, item := range rawSingers {
			if singerMap, ok := item.(map[string]interface{}); ok {
				if name := strings.TrimSpace(stringValue(singerMap["name"])); name != "" {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, " ")
	}
	return ""
}

func parseMetingMeta(metaJSON string) map[string]interface{} {
	meta := map[string]interface{}{}
	_ = json.Unmarshal([]byte(metaJSON), &meta)
	return meta
}

func normalizeMetingBase(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return "https://meting.mikus.ink"
	}
	return raw
}

func metingServer(platform string) string {
	return "tencent"
}

func getJSONList(rawURL string) ([]map[string]interface{}, error) {
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var anyList []map[string]interface{}
	if err := json.Unmarshal(data, &anyList); err == nil {
		return anyList, nil
	}

	var one map[string]interface{}
	if err := json.Unmarshal(data, &one); err == nil {
		return []map[string]interface{}{one}, nil
	}

	return nil, fmt.Errorf("decode meting response failed")
}

func getText(rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("meting request failed: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
