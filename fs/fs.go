package fs

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	defaultDirPermissions = 755
	appName               = "dumpler"
)

func createPath(dirs ...string) (string, error) {
	appPath := filepath.Join(dirs...)
	// if app dir does not exist, create it
	if _, err := os.Stat(appPath); errors.Is(err, os.ErrNotExist) {
		slog.Debug("Creating directory: " + appPath)
		if err = os.MkdirAll(appPath, defaultDirPermissions); err != nil {
			return "", err
		}
	}
	return appPath, nil
}

func GetAppConfigPath() (string, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return createPath(configPath, appName)
}

func GetAppCacheDir() (string, error) {
	// Get user specific config dir
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return createPath(cacheDir, appName)
}
