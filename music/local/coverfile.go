package local

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"windmusic/music/appdata"
)

type coverFileStore struct {
	dir string
}

func newCoverFileStore() (*coverFileStore, error) {
	root, err := appdata.AppDataRootDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, "local-covers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &coverFileStore{dir: dir}, nil
}

func (c *coverFileStore) SaveDataURL(key, dataURL string) error {
	key = strings.TrimSpace(key)
	if key == "" || strings.TrimSpace(dataURL) == "" {
		return nil
	}
	mime, raw, err := parseDataURL(dataURL)
	if err != nil {
		return err
	}
	ext := mimeToExt(mime)
	path := filepath.Join(c.dir, key+ext)
	return os.WriteFile(path, raw, 0o644)
}

func (c *coverFileStore) FilePath(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".bin"} {
		path := filepath.Join(c.dir, key+ext)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (c *coverFileStore) ReadDataURL(key string) (string, error) {
	path := c.FilePath(key)
	if path == "" {
		return "", nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	mime := mimeFromExt(filepath.Ext(path))
	encoded := base64.StdEncoding.EncodeToString(raw)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

func parseDataURL(dataURL string) (mime string, raw []byte, err error) {
	dataURL = strings.TrimSpace(dataURL)
	if !strings.HasPrefix(dataURL, "data:") {
		return "", nil, fmt.Errorf("unsupported cover data")
	}
	rest := strings.TrimPrefix(dataURL, "data:")
	comma := strings.Index(rest, ",")
	if comma <= 0 {
		return "", nil, fmt.Errorf("invalid cover data url")
	}
	meta := rest[:comma]
	payload := rest[comma+1:]
	mime = "image/jpeg"
	if semi := strings.Index(meta, ";"); semi >= 0 {
		mime = strings.TrimSpace(meta[:semi])
	} else if meta != "" {
		mime = strings.TrimSpace(meta)
	}
	if strings.HasSuffix(meta, ";base64") || strings.Contains(meta, ";base64") {
		raw, err = base64.StdEncoding.DecodeString(payload)
		return mime, raw, err
	}
	return mime, []byte(payload), nil
}

func mimeToExt(mime string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}
