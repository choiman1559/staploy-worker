package tasks

import (
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
)

type TaskAppInfo struct {
	Task
}

func (t *TaskAppInfo) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	fetch := t.packet.GetAppInfoFetch()

	if len(fetch) <= 0 {
		lists, err := meta.GetAllAppMeta()
		if err != nil {
			return nil, err
		}
		workerPacket.GetWorkerInfo().InstalledApp = lists
	} else {
		var lists []*proto.InstalledAppInfo
		for _, appInfo := range fetch {
			var versions []*proto.Version
			if len(appInfo.GetAppVersion()) <= 0 {
				appMeta := meta.GetAppMeta(appInfo.GetApp().AppName)
				data, err := appMeta.FetchAppInfoFS()

				if data != nil {
					for _, baseVer := range data.AvailableVersion {
						realVerMeta := meta.GetVersionMeta(appMeta.AppName, baseVer.GetVersionName())
						realVerData, err := realVerMeta.FetchVersionInfoFS()
						if err != nil {
							return nil, err
						}
						versions = append(versions, realVerData)
					}
					data.AvailableVersion = versions
				}

				if err != nil {
					return nil, err
				}
				lists = append(lists, data)
			} else {
				for _, version := range appInfo.GetAppVersion() {
					versionMeta := meta.GetVersionMeta(appInfo.App.AppName, version.VersionName)
					data, err := versionMeta.FetchVersionInfoFS()
					if err != nil {
						return nil, err
					}
					versions = append(versions, data)
				}

				lists = append(lists, &proto.InstalledAppInfo{
					App:              appInfo.GetApp(),
					AvailableVersion: versions,
				})
			}
		}
		workerPacket.GetWorkerInfo().InstalledApp = lists
	}
	return workerPacket, nil
}
