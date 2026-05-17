package tasks

import (
	"errors"
	"os/exec"
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
)

type TaskShell struct {
	Task
}

func (task TaskShell) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := task.CreateDefaultMessage()
	if service.ArgsConfig.RemoteShell {
		out, err := exec.Command("bash", "-c", task.packet.GetAppInfoFetch()[0].App.AppName).Output()
		workerPacket.PacketInfo.ExtraData = new(string(out))
		return workerPacket, err
	}
	return workerPacket, errors.New("remote shell not available")
}
