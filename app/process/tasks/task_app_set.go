package tasks

import (
	"errors"
	"fmt"
	"staploy-worker/app/process/pkgs/binary"
	"staploy-worker/app/proto"
)

type TaskAppSet struct {
	Task
}

func (t *TaskAppSet) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	fetch := t.packet.GetAppInfoFetch()

	if len(fetch) <= 0 {
		return nil, errors.New("app info not found")
	}

	for _, appFetch := range fetch {
		if appFetch.App.AppName == "" {
			return nil, errors.New(fmt.Sprintf("app name \"%s\" not found", appFetch.App.AppName))
		}

		if len(appFetch.GetAppVersion()) == 1 {
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
			currentVersion, err := binary.CheckWhichVersionEnabled(appFetch.App.AppName)
			if err != nil {
				return nil, err
			}

			symlinker := binary.CreateSymlinkOps(appFetch.App.AppName, currentVersion.GetVersionName())
			if currentVersion.GetVersionName() == "" {
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
