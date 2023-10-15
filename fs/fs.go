package fs

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	defaultDirPermissions  = 0755
	defaultFilePermissions = 0755
	appName                = "dumpler"
)

type FilesManager struct {
	config *viper.Viper
}

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

func fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateConfigFile(fileName string) error {
	pathFile, err := GetAppConfigPath()
	if err != nil {
		return err
	}

	filePathComplete := fmt.Sprintf("%s/%s", pathFile, fileName)
	var exists bool
	exists, err = fileExists(filePathComplete)
	if err != nil {
		return err
	}
	var file *os.File
	if exists {
		// Open existing file
		// Create flag to not overwrite this with another method
		file, err = os.OpenFile(filePathComplete, os.O_WRONLY, defaultFilePermissions)
		if err != nil {
			return err
		}
		slog.Debug("Existing session file opened and truncated")
	} else {
		// Create file
		file, err = os.Create(filePathComplete)
		if err != nil {
			return err
		}
		slog.Debug("New session file created")
	}
	defer file.Close()

	return nil
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
