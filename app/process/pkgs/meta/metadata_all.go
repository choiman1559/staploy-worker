package meta

import (
	"os"
	"path/filepath"

	"staploy-worker/app/files"
	"staploy-worker/app/proto"
)

func GetAllAppMeta() ([]*proto.InstalledAppInfo, error) {

	dir, err := files.GetBinDir()
	if err != nil {
		return nil, err
	}

	path, err := filepath.Abs(dir.Name())
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var lists []*proto.InstalledAppInfo
	for _, entry := range entries {
		if entry.IsDir() {
			appMeta := GetAppMeta(entry.Name())
			if appMeta.IsMetadataAvailable() {
				data, err := appMeta.FetchAppInfoFS()
				if err != nil {
					continue
				}
				lists = append(lists, data)
			}
		}
	}
	return lists, nil
}
