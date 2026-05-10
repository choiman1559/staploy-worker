package binary

import (
	"fmt"
	"os"
	"path/filepath"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
)

type Symlinks struct {
	AppMeta     *meta.AppMetadata
	VersionMeta *meta.VersionMetadata
}

func CreateSymlinkOps(appName string, versionName string) *Symlinks {
	symlinks := &Symlinks{
		AppMeta:     meta.GetAppMeta(appName),
		VersionMeta: meta.GetVersionMeta(appName, versionName),
	}
	return symlinks
}

func (symlinks *Symlinks) GetAppBinaryPath() (string, error) {
	binDir, err := files.GetBinDir()
	if err != nil {
		return "", err
	}

	binAbs, err := filepath.Abs(binDir.Name())
	if err != nil {
		return "", err
	}

	path := filepath.Join(binAbs, symlinks.AppMeta.AppName, symlinks.VersionMeta.VersionName)
	return path, nil
}

func (symlinks *Symlinks) ExportVersionLink() error {
	binPath, err := symlinks.GetAppBinaryPath()
	if err != nil {
		return err
	}

	exportFile, err := files.GetProfileDir()
	if err != nil {
		return err
	}

	exportPath, err := filepath.Abs(exportFile.Name())
	if err != nil {
		return err
	}

	versionInfo, err := symlinks.VersionMeta.FetchVersionInfoFS()
	if err != nil {
		return err
	}

	for _, binary := range versionInfo.EntryBinaries {
		originalBinPath := filepath.Join(binPath, binary.GetName())
		err := files.SetExecutable(originalBinPath, true)
		if err != nil {
			return err
		}

		err = os.Symlink(originalBinPath, filepath.Join(exportPath, binary.GetName()))
		if err != nil {
			return err
		}
	}

	appInfo, err := symlinks.AppMeta.FetchAppInfoFS()
	if err != nil {
		return err
	}

	var matchVersion *proto.Version
	for _, version := range appInfo.AvailableVersion {
		if version.VersionName == symlinks.VersionMeta.VersionName {
			matchVersion = version
			break
		}
	}

	if matchVersion == nil {
		return fmt.Errorf("version %s not found in available versions", symlinks.VersionMeta.VersionName)
	}

	appInfo.CurrentVersion = matchVersion
	err = symlinks.AppMeta.CommitAppInfoFS(appInfo)
	if err != nil {
		return err
	}
	return nil
}

func (symlinks *Symlinks) RemoveVersionLink() error {
	exportFile, err := files.GetProfileDir()
	if err != nil {
		return err
	}

	exportPath, err := filepath.Abs(exportFile.Name())
	if err != nil {
		return err
	}

	versionInfo, err := symlinks.VersionMeta.FetchVersionInfoFS()
	if err != nil {
		return err
	}

	for _, binary := range versionInfo.EntryBinaries {
		linkPath := filepath.Join(exportPath, binary.GetName())
		fi, err := os.Lstat(linkPath)
		if err == nil && fi.Mode()&os.ModeSymlink != 0 {
			err := os.Remove(linkPath)
			if err != nil {
				return err
			}
		}
	}

	appInfo, err := symlinks.AppMeta.FetchAppInfoFS()
	if err != nil {
		return err
	}

	appInfo.CurrentVersion = &proto.Version{}
	err = symlinks.AppMeta.CommitAppInfoFS(appInfo)
	if err != nil {
		return err
	}
	return nil
}

func (symlinks *Symlinks) CheckIsCurrentVersionLink() (bool, error) {
	appInfo, err := symlinks.AppMeta.FetchAppInfoFS()
	if err != nil {
		return false, err
	}
	return appInfo.CurrentVersion.VersionName == symlinks.VersionMeta.VersionName, nil
}

func CheckWhichVersionEnabled(appName string) (*proto.Version, error) {
	appMeta := meta.GetAppMeta(appName)
	appInfo, err := appMeta.FetchAppInfoFS()
	if err != nil {
		return nil, err
	}
	return appInfo.CurrentVersion, nil
}
