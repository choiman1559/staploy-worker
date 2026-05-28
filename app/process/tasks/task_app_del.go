package tasks

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/binary"
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

		if len(appFetch.GetAppVersion()) > 0 {
			for _, version := range appFetch.GetAppVersion() {
				if currentVersion != nil && currentVersion.GetVersionName() == version.GetVersionName() {
					log.Printf("Processing unlinking for package %s (%s)", appFetch.App.AppName, version.GetVersionName())
					symlinker := binary.CreateSymlinkOps(appFetch.App.AppName, version.GetVersionName())
					err := symlinker.RemoveVersionLink()
					if err != nil {
						return nil, err
					}
				}

				log.Printf("Removing package %s (%s)", appFetch.App.AppName, version.VersionName)
				appBinaryManager := binary.NewApp(appFetch.App, version.VersionName, "")
				err := appBinaryManager.UninstallAppPack()
				if err != nil {
					return nil, err
				}
			}
		} else {
			if currentVersion != nil && currentVersion.GetVersionName() != "" {
				symlinker := binary.CreateSymlinkOps(appFetch.App.AppName, currentVersion.GetVersionName())
				err := symlinker.RemoveVersionLink()
				if err != nil {
					return nil, err
				}
			}

			log.Printf("Removing all versions of package %s", appFetch.App.AppName)
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
