package files

import (
	"fmt"
	"os"
	"path/filepath"
	"staploy-worker/app/consts"
)

type IoConfig struct {
	BaseDir    string
	ProfileDir string
	CacheDir   string
	BufferSize int64
}

var dirConfig IoConfig

func SetConfig(config IoConfig) {
	dirConfig = config
}

func GetBaseDir() (*os.File, error) {
	return os.Open(dirConfig.BaseDir)
}

func GetCleanAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	return abs
}

func mkdirIfNotExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err := MkdirAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetBinDir() (*os.File, error) {
	binPath := fmt.Sprintf("%s/%s", GetCleanAbs(dirConfig.BaseDir), consts.DIR_PREFIX_BIN)
	err := mkdirIfNotExists(binPath)
	if err != nil {
		return nil, err
	}
	return os.Open(binPath)
}

func GetProfileDir() (*os.File, error) {
	var profileDir string
	if dirConfig.ProfileDir == "" {
		profileDir = fmt.Sprintf("%s/%s", GetCleanAbs(dirConfig.BaseDir), consts.DIR_PREFIX_PROFILE)
		err := mkdirIfNotExists(profileDir)
		if err != nil {
			return nil, err
		}
	} else {
		profileDir = dirConfig.ProfileDir
	}
	return os.Open(profileDir)
}

func GetCacheDir() (*os.File, error) {
	var cacheDir string
	if dirConfig.CacheDir == "" {
		cacheDir = fmt.Sprintf("%s/%s", GetCleanAbs(dirConfig.BaseDir), consts.DIR_PREFIX_CACHE)
		err := mkdirIfNotExists(cacheDir)
		if err != nil {
			return nil, err
		}
	} else {
		cacheDir = dirConfig.ProfileDir
	}
	return os.Open(cacheDir)
}
