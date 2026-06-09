package local

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

type basicAudioTags struct {
	Title       string
	Artist      string
	Album       string
	DurationSec float64
}

func readBasicAudioTags(absPath string) basicAudioTags {
	ext := strings.ToLower(filepath.Ext(absPath))
	base := strings.TrimSuffix(filepath.Base(absPath), ext)

	out := basicAudioTags{
		Title:  base,
		Artist: "未知艺术家",
	}

	file, err := os.Open(absPath)
	if err != nil {
		return out
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return out
	}
	if v := strings.TrimSpace(metadata.Title()); v != "" {
		out.Title = v
	}
	if v := strings.TrimSpace(metadata.Artist()); v != "" {
		out.Artist = v
	}
	if v := strings.TrimSpace(metadata.Album()); v != "" {
		out.Album = v
	}
	out.DurationSec = DurationSecondsFromMetadata(metadata)
	return out
}

func extractEmbeddedCoverData(absPath string) string {
	file, err := os.Open(absPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return ""
	}
	pic := metadata.Picture()
	if pic == nil || len(pic.Data) == 0 {
		return ""
	}
	mime := strings.TrimSpace(pic.MIMEType)
	if mime == "" {
		mime = "image/jpeg"
	}
	encoded := base64.StdEncoding.EncodeToString(pic.Data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded)
}

func probeAudioDuration(absPath string, ext string, size int64) float64 {
	file, err := os.Open(absPath)
	if err != nil {
		return 0
	}
	defer file.Close()
	return ProbeAudioDurationFromFile(file, ext, size)
}
