package music

import "time"

type PlatformInfo struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	Actions   []string `json:"actions"`
	Qualities []string `json:"qualities"`
}

type SourceInfo struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Author      string         `json:"author"`
	Homepage    string         `json:"homepage"`
	Filename    string         `json:"filename"`
	Enabled     bool           `json:"enabled"`
	Platforms   []PlatformInfo `json:"platforms"`
	ImportedAt  time.Time      `json:"importedAt"`
	Status      string         `json:"status"`
	Error       string         `json:"error,omitempty"`
}

type SongItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Singer   string `json:"singer"`
	Album    string `json:"album"`
	AlbumID  string `json:"albumId,omitempty"`
	Source   string `json:"source"`
	Interval string `json:"interval,omitempty"`
	Img      string `json:"img,omitempty"`
	SongMID  string `json:"songmid"`
	Hash     string `json:"hash,omitempty"`
	MetaJSON string `json:"metaJson"`
}

type SearchResult struct {
	List   []SongItem `json:"list"`
	Total  int        `json:"total"`
	Page   int        `json:"page"`
	Limit  int        `json:"limit"`
	Source string     `json:"source"`
}

type LyricInfo struct {
	Lyric   string `json:"lyric"`
	TLyric  string `json:"tlyric,omitempty"`
	RLyric  string `json:"rlyric,omitempty"`
	LXLyric string `json:"lxlyric,omitempty"`
}

type MusicInfo struct {
	Source    string                 `json:"source"`
	SongMID   string                 `json:"songmid"`
	Hash      string                 `json:"hash,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Singer    string                 `json:"singer,omitempty"`
	AlbumName string                 `json:"albumName,omitempty"`
	Raw       map[string]interface{} `json:"-"`
}

type FavoriteSong struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album,omitempty"`
	Duration string `json:"duration,omitempty"`
	CoverURL string `json:"coverUrl,omitempty"`

	SourceID string `json:"sourceId,omitempty"`
	Platform string `json:"platform,omitempty"`
	MetaJSON string `json:"metaJson,omitempty"`
}

type RecentSong struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Artist   string    `json:"artist"`
	Album    string    `json:"album,omitempty"`
	Duration string    `json:"duration,omitempty"`
	CoverURL string    `json:"coverUrl,omitempty"`
	SourceID string    `json:"sourceId,omitempty"`
	Platform string    `json:"platform,omitempty"`
	MetaJSON string    `json:"metaJson,omitempty"`
	PlayedAt time.Time `json:"playedAt"`
}
