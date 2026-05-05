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
		out, err := exec.Command(string(task.packet.PacketInfo.GetExtraData())).Output()
		workerPacket.PacketInfo.ExtraData = out
		return workerPacket, err
	}
	return workerPacket, errors.New("remote shell not available")
}
