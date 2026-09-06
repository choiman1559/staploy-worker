package tasks

import (
	"fmt"
	"log"
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

	var dwCode = t.packet.GetPacketInfo().GetExtraData()
	var saveName = fmt.Sprintf("%s_%s_%s.tar", fetch.App.AppName, fetch.AppVersion[0].VersionName, dwCode)
	log.Printf("Downloading package %s (%s)", fetch.App.AppName, fetch.AppVersion[0].VersionName)

	cacheDir, err := files.GetCacheDir()
	if err != nil {
		return nil, err
	}

	savePath := filepath.Join(cacheDir.Name(), saveName)
	var paths = fmt.Sprintf(consts.APIRouteSchema, "v1", consts.ConnTypeWorker)

	httpPrefix := "http"
	if service.ArgsConfig.Enable_mTLS {
		httpPrefix = "https"
	}

	var addr = fmt.Sprintf("%s://%s:%d%s", httpPrefix, service.ArgsConfig.Address, service.ArgsConfig.Port, paths)
	dwHeader := map[string]string{
		consts.BLOB_REQ_TYPE:          consts.BLOB_REQ_TYPE_DOWNLOAD,
		consts.BLOB_REQ_TYPE_DOWNLOAD: dwCode,
	}

	err = files.DownloadFileWithUrl(addr, savePath, dwHeader, service.TlsConfig)
	if err != nil {
		return nil, err
	}

	log.Printf("Unpacking package %s (%s)", fetch.App.AppName, fetch.AppVersion[0].VersionName)
	appRegistration := binary.NewApp(fetch.App, fetch.AppVersion[0].VersionName, savePath)
	err = appRegistration.InstallAppPack()
	if err != nil {
		return nil, err
	}

	if !service.ArgsConfig.SkipHashValidCheck {
		log.Printf("Verifying integrity of package %s (%s)", fetch.App.AppName, fetch.AppVersion[0].VersionName)
		result, err := appRegistration.VerifyAppPack()
		if err != nil {
			return nil, err
		}

		if !result {
			log.Printf("Package %s (%s) integrity check failed! Rollback...", fetch.App.AppName, fetch.AppVersion[0].VersionName)
			err := appRegistration.UninstallAppPack()
			if err != nil {
				return nil, err
			}
		}
	}

	defer func(path string) {
		err := files.RmdirAll(path)
		if err != nil {
			log.Printf("Failed to remove used-up app artifact: %s", path)
			return
		}
	}(savePath)
	return workerPacket, nil
}
