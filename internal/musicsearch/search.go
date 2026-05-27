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

func Search(platform, keyword string, page, limit int) (*music.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

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

	body, err := getJSON("https://music.163.com/api/search/get/web?"+params.Encode(), map[string]string{
		"Referer": "https://music.163.com/",
	})
	if err != nil {
		return nil, err
	}

	result, ok := body["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid netease search response")
	}

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
			Interval: formatDuration(song["duration"]),
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

	body, err := getJSON("https://search.kuwo.cn/r.s?client=kt&encoding=utf8&rformat=json&ver=mbox&vipver=1&pn="+params.Get("pn")+"&rn="+params.Get("rn")+"&all="+url.QueryEscape(keyword), map[string]string{
		"User-Agent": "Mozilla/5.0",
	})
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
			Interval: fmt.Sprint(song["DURATION"]),
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

	body, err := getJSON("https://songsearch.kugou.com/song_search_v2?"+params.Encode(), map[string]string{
		"User-Agent": "Mozilla/5.0",
	})
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
			Interval: fmt.Sprint(song["Duration"]),
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

	body, err := getJSON("https://c.y.qq.com/soso/fcgi-bin/client_search_cp?"+params.Encode(), map[string]string{
		"Referer": "https://y.qq.com/",
	})
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
			Interval: formatDurationSeconds(song["interval"]),
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

	body, err := getJSON("https://m.music.migu.cn/migu/remoting/scr_search_tag?"+params.Encode(), map[string]string{
		"User-Agent": "Mozilla/5.0",
	})
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
			Interval: fmt.Sprint(song["length"]),
			SongMID:  fmt.Sprint(song["copyrightId"]),
			MetaJSON: string(meta),
		})
	}

	total := int(numberValue(body["pgt"]))
	return &music.SearchResult{List: list, Total: total, Page: page, Limit: limit, Source: "mg"}, nil
}

func getJSON(rawURL string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
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

	text := strings.TrimSpace(string(data))
	if strings.HasPrefix(text, "callback(") {
		text = strings.TrimSuffix(strings.TrimPrefix(text, "callback("), ")")
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}
	return body, nil
}

func albumPic(song map[string]interface{}) string {
	if album, ok := song["album"].(map[string]interface{}); ok {
		if pic, ok := album["picUrl"].(string); ok {
			return pic
		}
	}
	return ""
}

func formatDuration(value interface{}) string {
	ms := numberValue(value)
	if ms <= 0 {
		return ""
	}
	seconds := int(ms / 1000)
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

func formatDurationSeconds(value interface{}) string {
	seconds := int(numberValue(value))
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
