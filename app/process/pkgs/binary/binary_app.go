package binary

import (
	"path/filepath"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
)

type AppPackageManager struct {
	ArchivePath string
	AppMeta     *meta.AppMetadata
	VersionMeta *meta.VersionMetadata
}

func NewApp(appName string, versionName string, archivePath string) *AppPackageManager {
	return &AppPackageManager{
		ArchivePath: archivePath,
		AppMeta:     meta.GetAppMeta(appName),
		VersionMeta: meta.GetVersionMeta(appName, versionName),
	}
}

func (app *AppPackageManager) InstallAppPack() error {
	binPath, err := app.getAppPackPath()
	if err != nil {
		return err
	}

	err = files.ExtractTar(app.ArchivePath, binPath)
	if err != nil {
		return err
	}

	newVersion := &proto.Version{VersionName: app.VersionMeta.VersionName}
	if app.AppMeta.IsMetadataAvailable() {
		appInfo, err := app.AppMeta.FetchAppInfoFS()
		if err != nil {
			return err
		}

		needAppending := true
		for _, version := range appInfo.AvailableVersion {
			if newVersion.GetVersionName() == version.GetVersionName() {
				needAppending = false
			}
		}

		if needAppending {
			appInfo.AvailableVersion = append(appInfo.AvailableVersion, newVersion)
		}

		err = app.AppMeta.CommitAppInfoFS(appInfo)
		if err != nil {
			return err
		}
	} else {
		appInfo := &proto.InstalledAppInfo{
			App: &proto.AppInfo{
				AppName: app.AppMeta.AppName,
			},
			CurrentVersion:   nil,
			AvailableVersion: make([]*proto.Version, 0),
		}

		appInfo.AvailableVersion = append(appInfo.AvailableVersion, newVersion)
		err := app.AppMeta.CommitAppInfoFS(appInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *AppPackageManager) UninstallAppPack() error {
	binPath, err := app.getAppPackPath()
	if err != nil {
		return err
	}

	err = files.RmdirAll(binPath)
	if err != nil {
		return err
	}

	appInfo, err := app.AppMeta.FetchAppInfoFS()
	if err != nil {
		return err
	}

	newAvailableVersion := make([]*proto.Version, 0)
	for _, version := range appInfo.AvailableVersion {
		if version.VersionName != app.VersionMeta.VersionName {
			newAvailableVersion = append(newAvailableVersion, version)
		}
	}

	appInfo.AvailableVersion = newAvailableVersion
	err = app.AppMeta.CommitAppInfoFS(appInfo)
	if err != nil {
		return err
	}

	return nil
}

func (app *AppPackageManager) getAppPackPath() (string, error) {
	binDir, err := files.GetBinDir()
	if err != nil {
		return "", err
	}

	binAbs, err := filepath.Abs(binDir.Name())
	if err != nil {
		return "", err
	}

	path := filepath.Join(binAbs, app.AppMeta.AppName, app.VersionMeta.VersionName)
	return path, nil
}
