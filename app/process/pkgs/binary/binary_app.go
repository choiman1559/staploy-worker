package binary

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
	"strings"

	"golang.org/x/sys/unix"
)

type AppPackageManager struct {
	ArchivePath    string
	AppMeta        *meta.AppMetadata
	VersionMeta    *meta.VersionMetadata
	AppDescription string
}

func NewApp(appInfo *proto.AppInfo, versionName string, archivePath string) *AppPackageManager {
	return &AppPackageManager{
		ArchivePath:    archivePath,
		AppMeta:        meta.GetAppMeta(appInfo.GetAppName()),
		VersionMeta:    meta.GetVersionMeta(appInfo.GetAppName(), versionName),
		AppDescription: appInfo.GetAppDescription(),
	}
}

func (app *AppPackageManager) InstallAppPack() error {
	binPath, err := app.getAppPackPath()
	if err != nil {
		return err
	}

	newVersion := &proto.Version{VersionName: app.VersionMeta.VersionName}
	var appInfoToUpdate *proto.InstalledAppInfo
	needAppending := true

	if app.AppMeta.IsMetadataAvailable() {
		appInfo, err := app.AppMeta.FetchAppInfoFS()
		if err != nil {
			return err
		}

		if appInfo.GetCurrentVersion() != nil {
			currentVersion := appInfo.GetCurrentVersion()
			if currentVersion.VersionName == newVersion.VersionName {
				return fmt.Errorf("requested binary \"%s\" (version %s) is currently activated, over-writing not allowed",
					appInfo.GetApp().GetAppName(), newVersion.VersionName)
			}
		}

		for _, version := range appInfo.AvailableVersion {
			if newVersion.GetVersionName() == version.GetVersionName() {
				needAppending = false
				isOnProc, err := app.CheckVersionOnProc()
				if err != nil {
					return err
				}

				if isOnProc {
					return fmt.Errorf("requested binary \"%s\" (version %s) is used by a running process, overwriting not allowed",
						appInfo.GetApp().GetAppName(), newVersion.VersionName)
				}
			}
		}
		appInfoToUpdate = appInfo
	} else {
		appInfo := &proto.InstalledAppInfo{
			App: &proto.AppInfo{
				AppName: app.AppMeta.AppName,
			},
			CurrentVersion:   nil,
			AvailableVersion: make([]*proto.Version, 0),
		}
		appInfoToUpdate = appInfo
	}

	if needAppending {
		appInfoToUpdate.AvailableVersion = append(appInfoToUpdate.AvailableVersion, newVersion)
	}

	if app.AppDescription != "" {
		appInfoToUpdate.App.AppDescription = &app.AppDescription
	}

	err = files.ExtractTar(app.ArchivePath, binPath)
	if err != nil {
		return err
	}

	err = app.AppMeta.CommitAppInfoFS(appInfoToUpdate)
	if err != nil {
		return err
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

func (app *AppPackageManager) VerifyAppPack() (bool, error) {
	versionInfo, err := app.VersionMeta.FetchVersionInfoFS()
	if err != nil {
		return false, err
	}

	symlinkOps := CreateSymlinkOps(app.AppMeta.AppName, app.VersionMeta.VersionName)
	versionDir, err := symlinkOps.GetAppBinaryPath(true)
	if err != nil {
		return false, err
	}

	for _, binary := range versionInfo.EntryBinaries {
		binaryPath := filepath.Join(versionDir, binary.Name)
		shaValue, err := getSHA1(binaryPath)
		if err != nil {
			return false, err
		}

		if binary.GetHash() == "" || binary.GetHash() != shaValue {
			return false, nil
		}
	}
	return true, nil
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

func (app *AppPackageManager) CheckVersionOnProc() (bool, error) {
	versionMeta, err := app.VersionMeta.FetchVersionInfoFS()
	if err != nil {
		return false, err
	}

	for _, executable := range versionMeta.EntryBinaries {
		executablePath, err := app.getVersionPackFilePath(executable.GetName())
		if err != nil {
			return false, err
		}

		isOnProc, err := IsFileUnderProcess(executablePath)
		if err != nil {
			return false, err
		}

		if isOnProc {
			return true, nil
		}
	}

	return false, nil
}

func (app *AppPackageManager) getVersionPackFilePath(filename string) (string, error) {
	appPackPath, err := app.getAppPackPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(appPackPath, strings.TrimSpace(filename)), nil
}

func getSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func IsFileUnderProcess(filePath string) (bool, error) {
	var target unix.Stat_t
	err := unix.Stat(filePath, &target)
	if err != nil {
		return false, fmt.Errorf("failed to get stat: %s: %w", filePath, err)
	}

	processes, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return false, err
	}

	for _, proc := range processes {
		var exeInfo unix.Stat_t
		err := unix.Stat(filepath.Join(proc, "exe"), &exeInfo)

		if err != nil {
			continue
		}

		if exeInfo.Dev == target.Dev && exeInfo.Ino == target.Ino {
			return true, nil
		}
	}

	return false, nil
}
