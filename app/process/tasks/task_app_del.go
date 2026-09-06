package tasks

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/binary"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
)

type TaskAppDelete struct {
	Task
}

func (t *TaskAppDelete) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	fetch := t.packet.GetAppInfoFetch()

	if len(fetch) <= 0 {
		return nil, errors.New("app info not found")
	}

	for _, appFetch := range fetch {
		if appFetch.App.AppName == "" {
			return nil, errors.New(fmt.Sprintf("app name \"%s\" not found", appFetch.App.AppName))
		}

		currentVersion, err := binary.CheckWhichVersionEnabled(appFetch.App.AppName)
		if err != nil {
			return nil, err
		}

		removeVersion := func(version *proto.Version) error {
			appBinaryManager := binary.NewApp(appFetch.App, version.VersionName, "")
			isUnderProc, err := appBinaryManager.CheckVersionOnProc()
			if err != nil {
				return err
			}

			if isUnderProc {
				return fmt.Errorf("app package \"%s\" (%s) is used by a running process", appFetch.App.AppName, version.GetVersionName())
			}

			if currentVersion != nil && currentVersion.GetVersionName() == version.GetVersionName() {
				log.Printf("Processing unlinking for package %s (%s)", appFetch.App.AppName, version.GetVersionName())
				symlinker := binary.CreateSymlinkOps(appFetch.App.AppName, version.GetVersionName())
				err := symlinker.RemoveVersionLink()
				if err != nil {
					return err
				}
			}

			log.Printf("Removing package %s (%s)", appFetch.App.AppName, version.VersionName)
			err = appBinaryManager.UninstallAppPack()
			if err != nil {
				return err
			}

			return nil
		}

		if len(appFetch.GetAppVersion()) > 0 {
			for _, version := range appFetch.GetAppVersion() {
				err := removeVersion(version)
				if err != nil {
					return nil, err
				}
			}
		} else {
			log.Printf("Removing all versions of package %s", appFetch.App.AppName)
			appMeta, err := meta.GetAppMeta(appFetch.App.AppName).FetchAppInfoFS()
			if err != nil {
				return nil, err
			}

			for _, version := range appMeta.AvailableVersion {
				err := removeVersion(version)
				if err != nil {
					return nil, err
				}
			}

			binDir, err := files.GetBinDir()
			if err != nil {
				return nil, err
			}

			binAbs, err := filepath.Abs(binDir.Name())
			if err != nil {
				return nil, err
			}

			path := filepath.Join(binAbs, appFetch.App.AppName)
			err = files.RmdirAll(path)
			if err != nil {
				return nil, err
			}
		}
	}

	return workerPacket, nil
}
