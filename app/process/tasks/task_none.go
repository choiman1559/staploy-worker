package tasks

import (
	"staploy-worker/app/proto"
)

type TaskNone struct {
	Task
}

func (t *TaskNone) InvokeTask() (*proto.WorkerPacket, error) {
	return t.CreateDefaultMessage(), nil
}
