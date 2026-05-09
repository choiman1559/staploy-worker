package tasks

import (
	"errors"
	"log"
	"os/exec"
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
)

type TaskShell struct {
	Task
}

func (task TaskShell) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := task.CreateDefaultMessage()

	log.Print(task.packet)
	if service.ArgsConfig.RemoteShell {
		out, err := exec.Command("bash", "-c", task.packet.GetAppInfoFetch()[0].App.AppName).Output()
		workerPacket.PacketInfo.ExtraData = out
		return workerPacket, err
	}
	return workerPacket, errors.New("remote shell not available")
}
