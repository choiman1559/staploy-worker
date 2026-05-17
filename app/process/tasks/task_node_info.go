package tasks

import (
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
)

type TaskNodeInfo struct {
	Task
}

func (t *TaskNodeInfo) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := &proto.WorkerPacket{
		PacketInfo: t.packet.GetPacketInfo(),
		WorkerInfo: service.CreateDefaultWorkerInfo(true),
	}
	return workerPacket, nil
}
