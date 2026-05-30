package appdata

import (
	"os"
	"path/filepath"
)

// AppDataRootDir returns the application config directory for persisted data.
func AppDataRootDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "windmusic"), nil
}
