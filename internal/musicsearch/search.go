package musicsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"windmusic/internal/music"
)

var client = &http.Client{Timeout: 20 * time.Second}

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

func Search(platform, keyword string, page, limit int) (*music.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	candidates := searchPlatformOrder(platform)
	var lastErr error
	for _, candidate := range candidates {
		result, err := searchPlatform(candidate, keyword, page, limit)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetryableSearchError(err) {
			break
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unsupported platform: %s", platform)
	}
	return nil, lastErr
}

func searchPlatformOrder(platform string) []string {
	fallbacks := []string{"kw", "kg", "tx", "mg", "wy"}
	if platform == "" {
		return fallbacks
	}
	order := []string{platform}
	for _, item := range fallbacks {
		if item != platform {
			order = append(order, item)
		}
	}
	return order
}

func isRetryableSearchError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTML instead of JSON") ||
		strings.Contains(msg, "decode search response") ||
		strings.Contains(msg, "invalid netease search response")
}

func searchPlatform(platform, keyword string, page, limit int) (*music.SearchResult, error) {
	switch platform {
	case "wy":
		return searchNetease(keyword, page, limit)
	case "kw":
		return searchKuwo(keyword, page, limit)
	case "kg":
		return searchKugou(keyword, page, limit)
	case "tx":
		return searchQQ(keyword, page, limit)
	case "mg":
		return searchMigu(keyword, page, limit)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func searchNetease(keyword string, page, limit int) (*music.SearchResult, error) {
	params := url.Values{}
	params.Set("s", keyword)
	params.Set("type", "1")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa((page-1)*limit))

	headers := mergeHeaders(map[string]string{
		"Referer": "https://music.163.com/",
		"Origin":  "https://music.163.com",
	})

	endpoints := []string{
		"https://music.163.com/api/cloudsearch/pc?" + params.Encode(),
		"https://music.163.com/api/search/get/web?" + params.Encode(),
	}

	var lastErr error
	for _, endpoint := range endpoints {
		body, err := getJSON(endpoint, headers)
		if err != nil {
			lastErr = err
			continue
		}
		result, ok := body["result"].(map[string]interface{})
		if !ok {
			lastErr = fmt.Errorf("invalid netease search response")
			continue
		}
		return parseNeteaseResult(result, page, limit)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("netease search failed")
	}
	return nil, lastErr
}

func parseNeteaseResult(result map[string]interface{}, page, limit int) (*music.SearchResult, error) {
	songs, _ := result["songs"].([]interface{})
	list := make([]music.SongItem, 0, len(songs))
	for _, item := range songs {
		song, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		artists := make([]string, 0)
		if rawArtists, ok := song["artists"].([]interface{}); ok {
			for _, artist := range rawArtists {
				if artistMap, ok := artist.(map[string]interface{}); ok {
					artists = append(artists, fmt.Sprint(artistMap["name"]))
				}
			}
		}
		albumName := ""
		if album, ok := song["album"].(map[string]interface{}); ok {
			albumName = fmt.Sprint(album["name"])
		}
		meta, _ := json.Marshal(song)
		list = append(list, music.SongItem{
			ID:       fmt.Sprint(song["id"]),
			Name:     fmt.Sprint(song["name"]),
			Singer:   strings.Join(artists, " / "),
			Album:    albumName,
			Source:   "wy",
			Interval: formatInterval(song["duration"]),
			Img:      albumPic(song),
			SongMID:  fmt.Sprint(song["id"]),
			MetaJSON: string(meta),
		})
	}

	total := int(numberValue(result["songCount"]))
	return &music.SearchResult{
		List:   list,
		Total:  total,
		Page:   page,
		Limit:  limit,
		Source: "wy",
	}, nil
}

func searchKuwo(keyword string, page, limit int) (*music.SearchResult, error) {
	params := url.Values{}
	params.Set("all", keyword)
	params.Set("pn", strconv.Itoa(page))
	params.Set("rn", strconv.Itoa(limit))
	params.Set("ft", "music")

	body, err := getJSON("https://search.kuwo.cn/r.s?client=kt&encoding=utf8&rformat=json&ver=mbox&vipver=1&pn="+params.Get("pn")+"&rn="+params.Get("rn")+"&all="+url.QueryEscape(keyword), mergeHeaders(map[string]string{
		"Referer": "https://www.kuwo.cn/",
	}))
	if err != nil {
		return nil, err
	}

	abslist, _ := body["abslist"].([]interface{})
	list := make([]music.SongItem, 0, len(abslist))
	for _, item := range abslist {
		song, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		meta, _ := json.Marshal(song)
		list = append(list, music.SongItem{
			ID:       fmt.Sprint(song["MUSICRID"]),
			Name:     fmt.Sprint(song["SONGNAME"]),
			Singer:   fmt.Sprint(song["ARTIST"]),
			Album:    fmt.Sprint(song["ALBUM"]),
			Source:   "kw",
			Interval: formatInterval(song["DURATION"]),
			Img:      kuwoPic(song, stringValue(body["BASEPICPATH"])),
			SongMID:  fmt.Sprint(song["MUSICRID"]),
			MetaJSON: string(meta),
		})
	}

	total := int(numberValue(body["TOTAL"]))
	return &music.SearchResult{List: list, Total: total, Page: page, Limit: limit, Source: "kw"}, nil
}

func searchKugou(keyword string, page, limit int) (*music.SearchResult, error) {
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("page", strconv.Itoa(page))
	params.Set("pagesize", strconv.Itoa(limit))
	params.Set("format", "json")

	body, err := getJSON("https://songsearch.kugou.com/song_search_v2?"+params.Encode(), mergeHeaders(map[string]string{
		"Referer": "https://www.kugou.com/",
		"Origin":  "https://www.kugou.com",
	}))
	if err != nil {
		return nil, err
	}

	data, _ := body["data"].(map[string]interface{})
	songs, _ := data["lists"].([]interface{})
	list := make([]music.SongItem, 0, len(songs))
	for _, item := range songs {
		song, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		meta, _ := json.Marshal(song)
		list = append(list, music.SongItem{
			ID:       fmt.Sprint(song["FileHash"]),
			Name:     fmt.Sprint(song["SongName"]),
			Singer:   fmt.Sprint(song["SingerName"]),
			Album:    fmt.Sprint(song["AlbumName"]),
			Source:   "kg",
			Interval: formatInterval(song["Duration"]),
			Img:      kugouPic(song),
			SongMID:  fmt.Sprint(song["FileHash"]),
			Hash:     fmt.Sprint(song["FileHash"]),
			MetaJSON: string(meta),
		})
	}

	total := int(numberValue(data["total"]))
	return &music.SearchResult{List: list, Total: total, Page: page, Limit: limit, Source: "kg"}, nil
}

func searchQQ(keyword string, page, limit int) (*music.SearchResult, error) {
	params := url.Values{}
	params.Set("w", keyword)
	params.Set("p", strconv.Itoa(page))
	params.Set("n", strconv.Itoa(limit))
	params.Set("format", "json")

	body, err := getJSON("https://c.y.qq.com/soso/fcgi-bin/client_search_cp?"+params.Encode(), mergeHeaders(map[string]string{
		"Referer": "https://y.qq.com/",
		"Origin":  "https://y.qq.com",
	}))
	if err != nil {
		return nil, err
	}

	data, _ := body["data"].(map[string]interface{})
	songData, _ := data["song"].(map[string]interface{})
	songs, _ := songData["list"].([]interface{})
	list := make([]music.SongItem, 0, len(songs))
	for _, item := range songs {
		song, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		singers := make([]string, 0)
		if rawSingers, ok := song["singer"].([]interface{}); ok {
			for _, singer := range rawSingers {
				if singerMap, ok := singer.(map[string]interface{}); ok {
					singers = append(singers, fmt.Sprint(singerMap["name"]))
				}
			}
		}
		meta, _ := json.Marshal(song)
		list = append(list, music.SongItem{
			ID:       fmt.Sprint(song["songmid"]),
			Name:     fmt.Sprint(song["songname"]),
			Singer:   strings.Join(singers, " / "),
			Album:    fmt.Sprint(song["albumname"]),
			Source:   "tx",
			Interval: formatInterval(song["interval"]),
			Img:      qqPic(song),
			SongMID:  fmt.Sprint(song["songmid"]),
			MetaJSON: string(meta),
		})
	}

	total := int(numberValue(songData["totalnum"]))
	return &music.SearchResult{List: list, Total: total, Page: page, Limit: limit, Source: "tx"}, nil
}

func searchMigu(keyword string, page, limit int) (*music.SearchResult, error) {
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("pgc", strconv.Itoa(page))
	params.Set("rows", strconv.Itoa(limit))
	params.Set("type", "2")

	body, err := getJSON("https://m.music.migu.cn/migu/remoting/scr_search_tag?"+params.Encode(), mergeHeaders(map[string]string{
		"Referer": "https://music.migu.cn/",
	}))
	if err != nil {
		return nil, err
	}

	musics, _ := body["musics"].([]interface{})
	list := make([]music.SongItem, 0, len(musics))
	for _, item := range musics {
		song, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		meta, _ := json.Marshal(song)
		list = append(list, music.SongItem{
			ID:       fmt.Sprint(song["copyrightId"]),
			Name:     fmt.Sprint(song["songName"]),
			Singer:   fmt.Sprint(song["singerName"]),
			Album:    fmt.Sprint(song["albumName"]),
			Source:   "mg",
			Interval: formatInterval(song["length"]),
			Img:      miguPic(song),
			SongMID:  fmt.Sprint(song["copyrightId"]),
			MetaJSON: string(meta),
		})
	}

	total := int(numberValue(body["pgt"]))
	return &music.SearchResult{List: list, Total: total, Page: page, Limit: limit, Source: "mg"}, nil
}

func mergeHeaders(extra map[string]string) map[string]string {
	headers := map[string]string{
		"User-Agent":      defaultUserAgent,
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
	}
	for key, value := range extra {
		headers[key] = value
	}
	return headers
}

func unwrapJSONPayload(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "\ufeff")

	if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
		return text
	}

	start := strings.IndexByte(text, '(')
	end := strings.LastIndexByte(text, ')')
	if start >= 0 && end > start {
		return strings.TrimSpace(text[start+1 : end])
	}
	return text
}

func getJSON(rawURL string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range mergeHeaders(headers) {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := unwrapJSONPayload(string(data))
	if strings.HasPrefix(text, "<") {
		snippet := text
		if len(snippet) > 120 {
			snippet = snippet[:120]
		}
		return nil, fmt.Errorf("search API returned HTML instead of JSON (status %d): %s", resp.StatusCode, snippet)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}
	return body, nil
}

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

func normalizeImageURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	if strings.HasPrefix(raw, "http://") {
		return "https://" + strings.TrimPrefix(raw, "http://")
	}
	return raw
}

func joinImageBase(base, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "//") {
		return normalizeImageURL(path)
	}
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return normalizeImageURL(path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return normalizeImageURL(base + path)
}

func albumPic(song map[string]interface{}) string {
	for _, key := range []string{"album", "al"} {
		album, ok := song[key].(map[string]interface{})
		if !ok {
			continue
		}
		if pic := stringValue(album["picUrl"]); pic != "" {
			return normalizeImageURL(pic)
		}
	}
	return ""
}

func kuwoPic(song map[string]interface{}, basePath string) string {
	for _, key := range []string{"web_albumpic_short", "web_albumpic", "MVPIC", "PIC"} {
		if pic := stringValue(song[key]); pic != "" {
			return joinImageBase(basePath, pic)
		}
	}
	return ""
}

func kugouPic(song map[string]interface{}) string {
	for _, key := range []string{"Image", "AlbumImg", "album_img", "img"} {
		pic := stringValue(song[key])
		if pic == "" {
			continue
		}
		pic = strings.ReplaceAll(pic, "{size}", "240")
		return normalizeImageURL(pic)
	}
	return ""
}

func qqPic(song map[string]interface{}) string {
	albummid := stringValue(song["albummid"])
	if albummid == "" {
		return ""
	}
	return fmt.Sprintf("https://y.gtimg.cn/music/photo_new/T002R300x300M000%s.jpg", albummid)
}

func miguPic(song map[string]interface{}) string {
	for _, key := range []string{"cover", "picM", "musicPic", "albumImg", "imgSrc"} {
		if pic := stringValue(song[key]); pic != "" {
			return normalizeImageURL(pic)
		}
	}
	return ""
}

func formatInterval(value interface{}) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		text = strings.TrimSpace(text)
		if text == "" {
			return ""
		}
		if formatted := parseDurationText(text); formatted != "" {
			return formatted
		}
		if num := numberValue(text); num > 0 {
			return formatIntervalFromNumber(num)
		}
		return ""
	}
	num := numberValue(value)
	if num <= 0 {
		return ""
	}
	return formatIntervalFromNumber(num)
}

func parseDurationText(text string) string {
	text = strings.TrimSpace(text)
	if !strings.Contains(text, ":") {
		return ""
	}
	parts := strings.Split(text, ":")
	if len(parts) == 2 {
		minutes, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		seconds, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 == nil && err2 == nil && minutes >= 0 && seconds >= 0 && seconds < 60 {
			return formatIntervalFromSeconds(minutes*60 + seconds)
		}
	}
	if len(parts) == 3 {
		hours, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		minutes, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		seconds, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err1 == nil && err2 == nil && err3 == nil && hours >= 0 && minutes >= 0 && seconds >= 0 && seconds < 60 {
			return formatIntervalFromSeconds(hours*3600 + minutes*60 + seconds)
		}
	}
	return ""
}

func formatIntervalFromNumber(num float64) string {
	if num >= 10000 {
		return formatIntervalFromSeconds(int(num / 1000))
	}
	return formatIntervalFromSeconds(int(num))
}

func formatIntervalFromSeconds(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

func numberValue(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}
