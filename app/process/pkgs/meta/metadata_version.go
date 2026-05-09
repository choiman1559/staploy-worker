package meta

import (
	"fmt"
	"path/filepath"
	"staploy-worker/app/consts"
	"staploy-worker/app/files"
	"staploy-worker/app/proto"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
)

type VersionMetadata struct {
	VersionName string
	AppName     string
	Lock        sync.Mutex
}

var versionMap = make(map[string]*VersionMetadata)

func GetVersionMeta(appName string, versionName string) *VersionMetadata {
	mapKey := fmt.Sprintf("%s_%s", appName, versionName)
	versionMeta := versionMap[mapKey]
	if versionMeta == nil {
		versionMeta = &VersionMetadata{
			VersionName: versionName,
			AppName:     appName,
			Lock:        sync.Mutex{},
		}
		versionMap[mapKey] = versionMeta
	}
	return versionMeta
}

func (version *VersionMetadata) IsMetadataAvailable() bool {
	path, err := version.GetVersionMetadataFile()
	if err != nil {
		return false
	}
	return files.Exists(path)
}

func (version *VersionMetadata) GetVersionMetadataFile() (string, error) {
	binDir, err := files.GetBinDir()
	if err != nil {
		return "", err
	}

	binAbs, err := filepath.Abs(binDir.Name())
	if err != nil {
		return "", err
	}

	path := filepath.Join(binAbs, version.AppName, version.VersionName, consts.FILENAME_METADATA)
	return path, nil
}

func (version *VersionMetadata) FetchVersionInfoFS() (*proto.Version, error) {
	version.Lock.Lock()
	defer version.Lock.Unlock()

	file, err := version.GetVersionMetadataFile()
	if err != nil {
		return nil, err
	}

	readFile, err := files.ReadFile(file)
	if err != nil {
		return nil, err
	}

	installedVersionInfo := &proto.Version{}
	err = protojson.Unmarshal(readFile, installedVersionInfo)
	if err != nil {
		return nil, err
	}
	return installedVersionInfo, nil
}

func (version *VersionMetadata) CommitVersionInfoFS(info *proto.Version) error {
	version.Lock.Lock()
	defer version.Lock.Unlock()

	file, err := version.GetVersionMetadataFile()
	if err != nil {
		return err
	}

	bytes, err := protojson.Marshal(info)
	err = files.MkdirAll(file)

	if err != nil {
		return err
	}

	err = files.WriteFile(file, bytes)
	if err != nil {
		return err
	}

	return nil
}
