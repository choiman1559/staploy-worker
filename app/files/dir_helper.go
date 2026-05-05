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

func GetBinDir() (*os.File, error) {
	return os.Open(fmt.Sprintf("%s/%s", GetCleanAbs(dirConfig.BaseDir), consts.DIR_PREFIX_BIN))
}

func GetProfileDir() (*os.File, error) {
	var profileDir string
	if dirConfig.ProfileDir == "" {
		profileDir = fmt.Sprintf("%s/%s", GetCleanAbs(dirConfig.BaseDir), consts.DIR_PREFIX_PROFILE)
	} else {
		profileDir = dirConfig.ProfileDir
	}
	return os.Open(profileDir)
}

func GetCacheDir() (*os.File, error) {
	var cacheDir string
	if dirConfig.CacheDir == "" {
		cacheDir = fmt.Sprintf("%s/%s", GetCleanAbs(dirConfig.BaseDir), consts.DIR_PREFIX_CACHE)
	} else {
		cacheDir = dirConfig.ProfileDir
	}
	return os.Open(cacheDir)
}
