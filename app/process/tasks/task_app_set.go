package tasks

import "staploy-worker/app/proto"

type TaskAppSet struct {
	Task
}

func (t *TaskAppSet) InvokeTask() (*proto.WorkerPacket, error) {
	return t.CreateDefaultMessage(), nil
}
