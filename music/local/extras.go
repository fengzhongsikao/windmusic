package local

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	models "windmusic/internal/music"
)

const localExtrasVersion = 2

type localSongExtraRef struct {
	CoverKey  string `json:"coverKey,omitempty"`
	Lyric     string `json:"lyric,omitempty"`
	CoverData string `json:"coverData,omitempty"` // legacy inline storage
}

type localExtrasFile struct {
	Version int                          `json:"version"`
	Covers  map[string]string            `json:"covers,omitempty"`
	Entries map[string]localSongExtraRef `json:"entries"`
}

func newExtrasFile() *localExtrasFile {
	return &localExtrasFile{
		Version: localExtrasVersion,
		Covers:  map[string]string{},
		Entries: map[string]localSongExtraRef{},
	}
}

func (e *localExtrasFile) ensureMaps() {
	if e.Covers == nil {
		e.Covers = map[string]string{}
	}
	if e.Entries == nil {
		e.Entries = map[string]localSongExtraRef{}
	}
}

func coverKeyForData(coverData string) string {
	sum := sha256.Sum256([]byte(coverData))
	return hex.EncodeToString(sum[:12])
}

func (e *localExtrasFile) assignSongRef(path, coverKey, lyric string) {
	e.ensureMaps()
	if coverKey == "" && lyric == "" {
		delete(e.Entries, path)
		return
	}
	e.Entries[path] = localSongExtraRef{
		CoverKey: strings.TrimSpace(coverKey),
		Lyric:    lyric,
	}
}

func (e *localExtrasFile) assignSong(path, coverData, lyric string) {
	e.ensureMaps()
	if coverData == "" && lyric == "" {
		delete(e.Entries, path)
		return
	}
	ref := localSongExtraRef{Lyric: lyric}
	if coverData != "" {
		key := coverKeyForData(coverData)
		e.Covers[key] = coverData
		ref.CoverKey = key
	}
	e.Entries[path] = ref
}

func (e *localExtrasFile) coverKeyAndData(path string) (string, string) {
	entry, ok := e.Entries[path]
	if !ok {
		return "", ""
	}
	if entry.CoverKey != "" {
		return entry.CoverKey, e.Covers[entry.CoverKey]
	}
	if entry.CoverData != "" {
		key := coverKeyForData(entry.CoverData)
		return key, entry.CoverData
	}
	return "", ""
}

func (e *localExtrasFile) coverForPath(path string) string {
	_, cover := e.coverKeyAndData(path)
	return cover
}

func (e *localExtrasFile) lyricForPath(path string) string {
	entry, ok := e.Entries[path]
	if !ok {
		return ""
	}
	return entry.Lyric
}

func (e *localExtrasFile) normalize() bool {
	e.ensureMaps()
	changed := false
	for path, entry := range e.Entries {
		if entry.CoverData == "" {
			continue
		}
		key := coverKeyForData(entry.CoverData)
		e.Covers[key] = entry.CoverData
		entry.CoverKey = key
		entry.CoverData = ""
		e.Entries[path] = entry
		changed = true
	}
	if e.Version < localExtrasVersion {
		e.Version = localExtrasVersion
		changed = true
	}
	return changed
}

func (e *localExtrasFile) pruneUnusedCovers() {
	e.ensureMaps()
	used := make(map[string]struct{}, len(e.Entries))
	for _, entry := range e.Entries {
		if entry.CoverKey != "" {
			used[entry.CoverKey] = struct{}{}
		}
	}
	for key := range e.Covers {
		if _, ok := used[key]; !ok {
			delete(e.Covers, key)
		}
	}
}

func (e *localExtrasFile) buildCoverBatch(paths []string, urlForKey func(string) string, files *coverFileStore) models.LocalCoverBatch {
	batch := models.LocalCoverBatch{
		Covers: make(map[string]string),
		Paths:  make(map[string]string, len(paths)),
	}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		entry, ok := e.Entries[path]
		if !ok {
			continue
		}
		key := strings.TrimSpace(entry.CoverKey)
		if key == "" {
			continue
		}
		cover := ""
		if urlForKey != nil {
			cover = strings.TrimSpace(urlForKey(key))
		}
		if cover == "" && files != nil {
			if data, err := files.ReadDataURL(key); err == nil {
				cover = data
			}
		}
		if cover == "" {
			cover = strings.TrimSpace(e.Covers[key])
		}
		if cover == "" {
			continue
		}
		batch.Covers[key] = cover
		batch.Paths[path] = key
	}
	return batch
}
