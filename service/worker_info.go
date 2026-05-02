package service

import (
	"runtime"
	"staploy-worker/proto"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

func CreateDefaultWorkerInfo(requireDetail bool) *proto.WorkerInfo {
	info, _ := host.Info()

	workerInfo := &proto.WorkerInfo{
		WorkerId:   GetWorkerUniqueId(),
		WorkerName: info.Hostname,
	}

	if requireDetail {
		workerInfo.BinLocation = &ArgsConfig.BinDir
		workerInfo.CpuArch = new(GetWorkerCpuArch())
		workerInfo.CpuCoreCount = new(GetCpuCoreCount())
		workerInfo.MemoryInBytes = new(GetTotalMemorySizeInBytes())
	}

	return workerInfo
}

func GetWorkerUniqueId() string {
	return "test-uid-go" //uuid.New().String()
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
