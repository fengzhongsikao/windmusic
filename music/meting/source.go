package meting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const metingSourcePrefix = "meting::"
const defaultMetingBaseURL = "https://meting.mikus.ink"

func SourceDisplayName(sourceID string) string {
	if base, ok := ParseMetingSourceID(sourceID); ok {
		return fmt.Sprintf("Meting(%s)", base)
	}
	return sourceID
}

func ParseMetingSourceID(sourceID string) (string, bool) {
	if !strings.HasPrefix(sourceID, metingSourcePrefix) {
		return "", false
	}
	base := strings.TrimSpace(strings.TrimPrefix(sourceID, metingSourcePrefix))
	if base == "" {
		return "", false
	}
	return strings.TrimSuffix(base, "/"), true
}

func ResolveMetingBase(sourceID string) string {
	if base, ok := ParseMetingSourceID(sourceID); ok {
		return base
	}
	return defaultMetingBaseURL
}

func BackendLogPrefix(sourceID string) string {
	if _, ok := ParseMetingSourceID(sourceID); ok {
		return "[后端meting]"
	}
	return "[后端meting]"
}

func AppDataRootDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "windmusic"), nil
}
