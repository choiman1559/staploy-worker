package meta

import (
	"path/filepath"
	"staploy-worker/app/consts"
	"staploy-worker/app/files"
	"staploy-worker/app/proto"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
)

type AppMetadata struct {
	AppName string
	Lock    sync.Mutex
}

var appMap = make(map[string]*AppMetadata)

func GetAppMeta(appName string) *AppMetadata {
	appMeta := appMap[appName]
	if appMeta == nil {
		appMeta = &AppMetadata{
			AppName: appName,
			Lock:    sync.Mutex{},
		}
		appMap[appName] = appMeta
	}
	return appMeta
}

func (app *AppMetadata) IsMetadataAvailable() bool {
	path, err := app.GetAppMetadataFile()
	if err != nil {
		return false
	}
	return files.Exists(path)
}

func (app *AppMetadata) GetAppMetadataFile() (string, error) {
	binDir, err := files.GetBinDir()
	if err != nil {
		return "", err
	}

	binAbs, err := filepath.Abs(binDir.Name())
	if err != nil {
		return "", err
	}

	path := filepath.Join(binAbs, app.AppName, consts.FILENAME_METADATA)
	return path, nil
}

func (app *AppMetadata) FetchAppInfoFS() (*proto.InstalledAppInfo, error) {
	app.Lock.Lock()
	defer app.Lock.Unlock()

	file, err := app.GetAppMetadataFile()
	if err != nil {
		return nil, err
	}

	readFile, err := files.ReadFile(file)
	if err != nil {
		return nil, err
	}

	installedAppInfo := &proto.InstalledAppInfo{}
	err = protojson.Unmarshal(readFile, installedAppInfo)
	if err != nil {
		return nil, err
	}
	return installedAppInfo, nil
}

func (app *AppMetadata) CommitAppInfoFS(info *proto.InstalledAppInfo) error {
	app.Lock.Lock()
	defer app.Lock.Unlock()

	file, err := app.GetAppMetadataFile()
	if err != nil {
		return err
	}

	info.CurrentVersion = ClearVersion(info.GetCurrentVersion())
	availableVersions := info.GetAvailableVersion()

	for i := 0; i < len(availableVersions); i += 1 {
		availableVersions[i] = ClearVersion(availableVersions[i])
	}

	info.AvailableVersion = availableVersions
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

func ClearVersion(version *proto.Version) *proto.Version {
	return &proto.Version{
		VersionName: version.VersionName,
	}
}
