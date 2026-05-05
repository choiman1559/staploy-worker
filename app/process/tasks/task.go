package tasks

import (
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
)

type TaskInvoker interface {
	InvokeTask() (*proto.WorkerPacket, error)
	InitTask(session *service.Session, packet *proto.ServerPacket)
	CreateDefaultMessage() *proto.WorkerPacket
}

type Task struct {
	TaskInvoker
	session *service.Session
	packet  *proto.ServerPacket
}

func (task *Task) InitTask(session *service.Session, packet *proto.ServerPacket) {
	task.session = session
	task.packet = packet
}

func (task *Task) CreateDefaultMessage() *proto.WorkerPacket {
	workerPacket := &proto.WorkerPacket{
		PacketInfo: task.packet.GetPacketInfo(),
		WorkerInfo: service.CreateDefaultWorkerInfo(false),
	}

	workerPacket.PacketInfo.ExtraData = []byte{}
	return workerPacket
}
