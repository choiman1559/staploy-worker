package process

import (
	"staploy-worker/app/consts"
	"staploy-worker/app/process/tasks"
	"staploy-worker/app/proto"
	"staploy-worker/app/service"
)

var currentTaskMap = SafeMap{}
var executableTaskMap = map[proto.ActionProcedure]tasks.TaskInvoker{
	proto.ActionProcedure_PROCEDURE_NONE:               &tasks.TaskNone{},
	proto.ActionProcedure_PROCEDURE_REQUEST_APP_INFO:   &tasks.TaskAppInfo{},
	proto.ActionProcedure_PROCEDURE_PUSH_APP_BINARY:    &tasks.TaskNone{}, // Reserved, But no used by worker (Via http file request header)
	proto.ActionProcedure_PROCEDURE_ADD_APP_VERSION:    &tasks.TaskAppAdd{},
	proto.ActionProcedure_PROCEDURE_SET_APP_VERSION:    &tasks.TaskAppSet{},
	proto.ActionProcedure_PROCEDURE_DELETE_APP_VERSION: &tasks.TaskAppDelete{},
	proto.ActionProcedure_PROCEDURE_EXECUTE_SHELL:      &tasks.TaskShell{},
}

func processAppAction(session *service.Session, packet *proto.ServerPacket) {
	resultObj := proto.Result{
		TaskStarted:    true,
		ResultFinished: false,
	}

	currentTaskMap.Set(packet.GetPacketInfo().GetChallengeCode(), &resultObj)
	targetTask := executableTaskMap[packet.GetPacketInfo().GetActionProcedure()]
	if targetTask == nil {
		resultObj.ResultFinished = false
		resultObj.ResultSuccessful = false
		resultObj.ErrorMessage = new(consts.ERROR_TASK_NOT_FOUND)

		sendMessage(session, &proto.WorkerPacket{
			PacketInfo: packet.GetPacketInfo(),
			TaskResult: &resultObj,
		})
		currentTaskMap.Set(packet.GetPacketInfo().GetChallengeCode(), &resultObj)
		return
	}

	targetTask.InitTask(session, packet)
	workerPacket, err := targetTask.InvokeTask()
	resultObj.ResultFinished = true

	if err != nil {
		resultObj.ResultSuccessful = false
		resultObj.ErrorMessage = new(err.Error())
		workerPacket = targetTask.CreateDefaultMessage()
	} else {
		resultObj.ResultSuccessful = true
	}

	workerPacket.TaskResult = &resultObj
	currentTaskMap.Set(packet.GetPacketInfo().GetChallengeCode(), &resultObj)
	sendMessage(session, workerPacket)
}
