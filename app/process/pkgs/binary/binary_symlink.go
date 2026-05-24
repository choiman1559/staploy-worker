package binary

import (
	"fmt"
	"os"
	"path/filepath"
	"staploy-worker/app/consts"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
	"strings"
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

func DualDifference(s []*proto.BinaryInfo, other []*proto.BinaryInfo) ([]*proto.BinaryInfo, []*proto.BinaryInfo, []*proto.BinaryInfo) {
	var aMinusB []*proto.BinaryInfo
	var bMinusA []*proto.BinaryInfo
	var aAndB []*proto.BinaryInfo

	checkContain := func(d []*proto.BinaryInfo, f *proto.BinaryInfo) bool {
		for _, i := range d {
			if i.Name == f.Name {
				return true
			}
		}
		return false
	}

	for _, data := range s {
		if !checkContain(other, data) {
			aMinusB = append(aMinusB, data)
		}
	}

	for _, data := range other {
		if !checkContain(s, data) {
			bMinusA = append(bMinusA, data)
		}
	}

	for _, data := range other {
		if checkContain(s, data) {
			aAndB = append(aAndB, data)
		}
	}

	return aMinusB, bMinusA, aAndB
}

func CreateLinks(binaries []*proto.BinaryInfo, binPath string, exportPath string) error {
	for _, binary := range binaries {
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
	return nil
}

func RemoveLink(linkPath string) error {
	fi, err := os.Lstat(linkPath)
	if err == nil && fi.Mode()&os.ModeSymlink != 0 {
		err := os.Remove(linkPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveLinks(binaries []*proto.BinaryInfo, exportPath string) error {
	for _, binary := range binaries {
		linkPath := filepath.Join(exportPath, binary.GetName())
		err := os.Remove(linkPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsLinkPointToActual(binary *proto.BinaryInfo, exportPath string) bool {
	linkPath := filepath.Join(exportPath, binary.GetName())
	target, err := os.Readlink(linkPath)
	if err != nil {
		return false
	}
	return !strings.Contains(target, fmt.Sprintf("/%s/%s", consts.FILENAME_SYMLINK_DIR, binary.GetName()))
}

func (symlinks *Symlinks) GetAppBinaryPath(isVersion bool) (string, error) {
	binDir, err := files.GetBinDir()
	if err != nil {
		return "", err
	}

	binAbs, err := filepath.Abs(binDir.Name())
	if err != nil {
		return "", err
	}

	var path string
	if isVersion {
		path = filepath.Join(binAbs, symlinks.AppMeta.AppName, symlinks.VersionMeta.VersionName)
	} else {
		path = filepath.Join(binAbs, symlinks.AppMeta.AppName)
	}
	return path, nil
}

func (symlinks *Symlinks) GetAppActiveSymlinkDir() (string, error) {
	linkBaseDir, err := symlinks.GetAppBinaryPath(false)
	if err != nil {
		return "", err
	}
	return filepath.Join(linkBaseDir, consts.FILENAME_SYMLINK_DIR), nil
}

func (symlinks *Symlinks) ExportVersionLink() error {
	binPath, err := symlinks.GetAppBinaryPath(true)
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

	appInfo, err := symlinks.AppMeta.FetchAppInfoFS()
	if err != nil {
		return err
	}

	appSymlinkDir, err := symlinks.GetAppActiveSymlinkDir()
	if err != nil {
		return err
	}

	err = RemoveLink(appSymlinkDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Symlink(binPath, appSymlinkDir)
	if err != nil {
		return err
	}

	if service.ArgsConfig.DisableSymlinkDir {
		err := CreateLinks(versionInfo.EntryBinaries, binPath, exportPath)
		if err != nil {
			return err
		}
	} else {
		var binaryToAdd []*proto.BinaryInfo
		preVersionName := appInfo.GetCurrentVersion().GetVersionName()

		if preVersionName != "" {
			preVersionMeta := meta.GetVersionMeta(appInfo.GetApp().GetAppName(), preVersionName)
			preVersionInfo, err := preVersionMeta.FetchVersionInfoFS()

			if err != nil {
				return err
			}

			binaryToAddTemp, binaryToRm, binaryRemain := DualDifference(versionInfo.EntryBinaries, preVersionInfo.EntryBinaries)
			binaryToAdd = binaryToAddTemp

			for _, binary := range binaryRemain {
				if IsLinkPointToActual(binary, exportPath) {
					binaryToRm = append(binaryToRm, binary)
					binaryToAdd = append(binaryToAdd, binary)
				}
			}

			err = RemoveLinks(binaryToRm, exportPath)
			if err != nil {
				return err
			}
		} else {
			binaryToAdd = versionInfo.EntryBinaries
		}

		err = CreateLinks(binaryToAdd, appSymlinkDir, exportPath)
		if err != nil {
			return err
		}
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

	err = RemoveLinks(versionInfo.EntryBinaries, exportPath)
	if err != nil {
		return err
	}

	activeDir, err := symlinks.GetAppActiveSymlinkDir()
	if err != nil {
		return err
	}
	err = RemoveLink(activeDir)
	if err != nil && !os.IsNotExist(err) {
		return err
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
