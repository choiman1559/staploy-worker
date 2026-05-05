package tasks

import (
	"staploy-worker/app/proto"
)

type TaskAppAdd struct {
	Task
}

func (t *TaskAppAdd) InvokeTask() (*proto.WorkerPacket, error) {
	return t.CreateDefaultMessage(), nil
}
