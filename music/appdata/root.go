package appdata

import (
	"os"
	"path/filepath"
	"strings"
)

// AppDataRootDir returns the application config directory for persisted data.
func AppDataRootDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("WINDMUSIC_APPDATA")); override != "" {
		return override, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "windmusic"), nil
}
