package service

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"staploy-worker/app/consts"
	"staploy-worker/app/files"
	"staploy-worker/app/proto"
	"strings"
	"sync/atomic"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
	gcpu "github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sys/cpu"
)

var atomicWorkerDefaultInfo atomic.Value
var atomicWorkerUniqueId atomic.Value
var uuidLock *flock.Flock

func CreateDefaultWorkerInfo(requireDetail bool) *proto.WorkerInfo {
	if !requireDetail {
		return &proto.WorkerInfo{
			WorkerId: GetWorkerUniqueId(),
		}
	}

	workerInfo := atomicWorkerDefaultInfo.Load()
	if workerInfo == nil {
		info, _ := host.Info()
		workerName := info.Hostname
		if ArgsConfig.OverrideName != "" {
			workerName = ArgsConfig.OverrideName
		}

		workerInfo = &proto.WorkerInfo{
			WorkerId:   GetWorkerUniqueId(),
			WorkerName: workerName,
			WorkerFlags: &proto.WorkerFlags{
				BUFFER_SIZE:            ArgsConfig.BufferSize,
				USE_REMOTE_SHELL:       ArgsConfig.RemoteShell,
				DISABLE_SYMLINK_DIR:    ArgsConfig.DisableSymlinkDir,
				SKIP_HASH_VERIFICATION: ArgsConfig.SkipHashValidCheck,

				CPU_BIG_ENDIAN:   IsCPUBigEndian(),
				CPU_CAPABILITIES: GetCpuExtensions(),
				CPU_FLAGS:        GetCpuFlags(),
			},
		}
		atomicWorkerDefaultInfo.Store(workerInfo)
	}

	workerInfo.(*proto.WorkerInfo).BinLocation = &ArgsConfig.BaseDir
	workerInfo.(*proto.WorkerInfo).CpuArch = new(GetWorkerCpuArch())
	workerInfo.(*proto.WorkerInfo).CpuCoreCount = new(GetCpuCoreCount())
	workerInfo.(*proto.WorkerInfo).MemoryInBytes = new(GetTotalMemorySizeInBytes())

	return workerInfo.(*proto.WorkerInfo)
}

func getUUIDFilePath() (string, error) {
	fil, err := files.GetBaseDir()
	if err != nil {
		log.Fatal(err)
	}

	absPath, _ := filepath.Abs(fil.Name())
	configPath := fmt.Sprintf("%s/%s", absPath, consts.FILENAME_BASE_UUID)
	return configPath, err
}

func GetWorkerUUIDLock() *flock.Flock {
	if uuidLock == nil {
		configPath, err := getUUIDFilePath()
		if err != nil {
			log.Fatal(err)
		}

		uuidLock = flock.New(configPath)
	}
	return uuidLock
}

func GetWorkerUniqueId() string {
	currentId := atomicWorkerUniqueId.Load()
	if currentId == nil || currentId == "" {
		configPath, err := getUUIDFilePath()
		if files.Exists(configPath) {
			content, err := files.ReadFileString(configPath)
			if err != nil {
				log.Fatal(err)
			} else if content != "" {
				atomicWorkerUniqueId.Store(content)
				return content
			}
		}

		newId := uuid.New().String()
		atomicWorkerUniqueId.Store(newId)
		err = files.WriteFileString(configPath, newId)

		if err != nil {
			log.Fatal(err)
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

func GetCpuExtensions() []string {
	var caps []string
	var v reflect.Value

	switch runtime.GOARCH {
	case "386", "amd64":
		v = reflect.ValueOf(cpu.X86)
	case "arm64":
		v = reflect.ValueOf(cpu.ARM64)
	case "arm":
		v = reflect.ValueOf(cpu.ARM)
	case "mips64", "mips64le":
		v = reflect.ValueOf(cpu.MIPS64X)
	case "ppc64", "ppc64le":
		v = reflect.ValueOf(cpu.PPC64)
	case "s390x":
		v = reflect.ValueOf(cpu.S390X)
	case "riscv64":
		v = reflect.ValueOf(cpu.RISCV64)
	case "loong64":
		v = reflect.ValueOf(cpu.Loong64)
	default:
		return []string{}
	}

	prefixRegex := regexp.MustCompile(`^(Has|Is)`)

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.Type.Kind() == reflect.Bool {
			if v.Field(i).Bool() {
				cleanName := prefixRegex.ReplaceAllString(field.Name, "")
				extName := strings.ToLower(cleanName)

				caps = append(caps, extName)
			}
		}
	}
	return caps
}

func GetCpuFlags() []string {
	cpus, _ := gcpu.Info()
	return cpus[0].Flags
}

func IsCPUBigEndian() bool {
	return cpu.IsBigEndian
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
	case "mips":
		return proto.CpuArch_mips
	case "mips64":
		return proto.CpuArch_mips64
	case "ppc64":
		return proto.CpuArch_ppc64
	case "ppc64le":
		return proto.CpuArch_ppc64le
	case "s390x":
		return proto.CpuArch_s390x
	case "loong64":
		return proto.CpuArch_loong64
	default:
		return proto.CpuArch_UNKNOWN
	}
}
