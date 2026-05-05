package tasks

import "staploy-worker/app/proto"

type TaskAppDelete struct {
	Task
}

func (t *TaskAppDelete) InvokeTask() (*proto.WorkerPacket, error) {
	return t.CreateDefaultMessage(), nil
}
