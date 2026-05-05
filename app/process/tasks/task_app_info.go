package tasks

import "staploy-worker/app/proto"

type TaskAppInfo struct {
	Task
}

func (t *TaskAppInfo) InvokeTask() (*proto.WorkerPacket, error) {
	workerPacket := t.CreateDefaultMessage()
	return workerPacket, nil
}
