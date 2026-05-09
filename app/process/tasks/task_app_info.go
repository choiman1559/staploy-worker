package tasks

import (
	"log"
	"staploy-worker/app/process/pkgs/meta"
	"staploy-worker/app/proto"
)

type TaskAppInfo struct {
	Task
}

func (t *TaskAppInfo) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	fetch := t.packet.GetAppInfoFetch()

	log.Print(t.packet)

	if len(fetch) <= 0 {
		lists, err := meta.GetAllAppMeta()
		if err != nil {
			return nil, err
		}
		workerPacket.GetWorkerInfo().InstalledApp = lists
	} else {
		var lists []*proto.InstalledAppInfo
		for _, appInfo := range fetch {
			if len(appInfo.GetAppVersion()) <= 0 {
				appMeta := meta.GetAppMeta(appInfo.GetApp().AppName)
				data, err := appMeta.FetchAppInfoFS()
				if err != nil {
					return nil, err
				}
				lists = append(lists, data)
			} else {
				var versions []*proto.Version
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
