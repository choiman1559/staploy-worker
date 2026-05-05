package service

import (
	"fmt"
	"path/filepath"
	"runtime"
	"staploy-worker/app/consts"
	"staploy-worker/app/files"
	"staploy-worker/app/proto"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

var atomicWorkerDefaultInfo atomic.Value
var atomicWorkerUniqueId atomic.Value

func CreateDefaultWorkerInfo(requireDetail bool) *proto.WorkerInfo {
	workerInfo := atomicWorkerDefaultInfo.Load().(*proto.WorkerInfo)
	if workerInfo == nil {
		info, _ := host.Info()
		workerInfo = &proto.WorkerInfo{
			WorkerId:   GetWorkerUniqueId(),
			WorkerName: info.Hostname,
			WorkerFlags: &proto.WorkerFlags{
				BUFFER_SIZE:      ArgsConfig.BufferSize,
				USE_REMOTE_SHELL: ArgsConfig.RemoteShell,
			},
		}
		atomicWorkerDefaultInfo.Store(workerInfo)
	}

	if requireDetail {
		workerInfo.BinLocation = &ArgsConfig.BaseDir
		workerInfo.CpuArch = new(GetWorkerCpuArch())
		workerInfo.CpuCoreCount = new(GetCpuCoreCount())
		workerInfo.MemoryInBytes = new(GetTotalMemorySizeInBytes())
	}

	return workerInfo
}

func GetWorkerUniqueId() string {
	currentId := atomicWorkerUniqueId.Load()
	if currentId == nil || currentId == "" {
		fil, err := files.GetBaseDir()
		if err != nil {
			panic(err)
		}

		absPath, _ := filepath.Abs(fil.Name())
		configPath := fmt.Sprintf("%s/%s", absPath, consts.FILE_BASE_UUID)

		if files.Exists(configPath) {
			content, err := files.ReadFileString(configPath)
			if err != nil {
				panic(err)
			} else if content != "" {
				atomicWorkerUniqueId.Store(content)
				return content
			}
		}

		newId := uuid.New().String()
		atomicWorkerUniqueId.Store(newId)
		err = files.WriteFileString(configPath, newId)

		if err != nil {
			panic(err)
		}
		return newId
	}
	return currentId.(string)
}

func GetTotalMemorySizeInBytes() int64 {
	v, _ := mem.VirtualMemory()
	return int64(v.Total)
}

func GetCpuCoreCount() int64 {
	return int64(runtime.NumCPU())
}

func GetWorkerCpuArch() proto.CpuArch {
	switch runtime.GOARCH {
	case "386":
		return proto.CpuArch_i386
	case "amd64":
		return proto.CpuArch_x86_64
	case "arm":
		return proto.CpuArch_arm
	case "arm64":
		return proto.CpuArch_aarch64
	case "riscv32":
		return proto.CpuArch_riscv32
	case "riscv64":
		return proto.CpuArch_riscv64
	case "mipsle":
		return proto.CpuArch_mipsel
	case "mips64le":
		return proto.CpuArch_mips64el
	default:
		return proto.CpuArch_UNKNOWN
	}
}
