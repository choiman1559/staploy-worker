package tasks

import (
	"fmt"
	"path/filepath"
	"staploy-worker/app/consts"
	"staploy-worker/app/files"
	"staploy-worker/app/process/pkgs/binary"
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
)

type TaskAppAdd struct {
	Task
}

func (t *TaskAppAdd) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	fetch := t.packet.GetAppInfoFetch()[0]

	var dwCode = string(t.packet.GetPacketInfo().ExtraData)
	var saveName = fmt.Sprintf("%s_%s_%s.tar", fetch.App.AppName, fetch.AppVersion[0].VersionName, dwCode)

	cacheDir, err := files.GetCacheDir()
	if err != nil {
		return nil, err
	}

	savePath := filepath.Join(cacheDir.Name(), saveName)
	var paths = fmt.Sprintf(consts.APIRouteSchema, "v1", consts.ConnTypeWorker)

	//goland:noinspection ALL (disable http warning)
	var addr = fmt.Sprintf("http://%s:%d%s", service.ArgsConfig.Address, service.ArgsConfig.Port, paths)

	dwHeader := map[string]string{
		consts.BLOB_REQ_TYPE:          consts.BLOB_REQ_TYPE_DOWNLOAD,
		consts.BLOB_REQ_TYPE_DOWNLOAD: dwCode,
	}

	err = files.DownloadFromUrl(addr, savePath, dwHeader)
	if err != nil {
		return nil, err
	}

	appRegistration := binary.NewApp(fetch.App.AppName, fetch.AppVersion[0].VersionName, savePath)
	err = appRegistration.InstallAppPack()
	if err != nil {
		return nil, err
	}

	defer func(path string) {
		err := files.RmdirAll(path)
		if err != nil {
			return
		}
	}(savePath)
	return workerPacket, nil
}
