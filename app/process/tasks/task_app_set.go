package tasks

import (
	"errors"
	"fmt"
	"log"
	"staploy-worker/app/process/pkgs/binary"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
)

type TaskAppSet struct {
	Task
}

func (t *TaskAppSet) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	fetch := t.packet.GetAppInfoFetch()

	if len(fetch) <= 0 {
		return nil, errors.New("app info not found in argument")
	}

	for _, appFetch := range fetch {
		if appFetch.App.AppName == "" {
			return nil, errors.New(fmt.Sprintf("app name \"%s\" not found in argument", appFetch.App.AppName))
		}

		appMeta := meta.GetAppMeta(appFetch.App.AppName)
		if !appMeta.IsMetadataAvailable() {
			return nil, errors.New(fmt.Sprintf("app metadata %s is not available on local filesystem", appFetch.App.AppName))
		}

		if len(appFetch.GetAppVersion()) == 1 {
			versionMeta := meta.GetVersionMeta(appFetch.App.AppName, appFetch.AppVersion[0].GetVersionName())
			if !versionMeta.IsMetadataAvailable() {
				return nil, fmt.Errorf("requested set package \"%s\" (%s) not available on local storage", appFetch.App.AppName, appFetch.AppVersion[0].GetVersionName())
			}

			log.Printf("Processing triggers for package %s (%s)", appFetch.App.AppName, appFetch.AppVersion[0].VersionName)
			symlinker := binary.CreateSymlinkOps(appFetch.App.AppName, appFetch.AppVersion[0].VersionName)
			isSymlinked, err := symlinker.CheckIsCurrentVersionLink()
			if err != nil {
				return nil, err
			}

			if !isSymlinked {
				err := symlinker.ExportVersionLink()
				if err != nil {
					return nil, err
				}
			}
		} else if len(appFetch.GetAppVersion()) == 0 {
			log.Printf("Processing unlinking for package %s", appFetch.App.AppName)
			currentVersion, err := binary.CheckWhichVersionEnabled(appFetch.App.AppName)
			if err != nil {
				return nil, err
			}

			symlinker := binary.CreateSymlinkOps(appFetch.App.AppName, currentVersion.GetVersionName())
			if currentVersion.GetVersionName() != "" {
				err := symlinker.RemoveVersionLink()
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("no enabled app version for specified app")
			}
		} else {
			return nil, errors.New("app version length error, must have either 0 or 1 of version specified")
		}
	}

	return workerPacket, nil
}
