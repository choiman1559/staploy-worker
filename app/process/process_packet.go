package process

import (
	"fmt"
	"staploy-worker/app/proto"
	"staploy-worker/app/service"

	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"
)

func PacketProcess(session *service.Session, data []byte) {
	packet := &proto.ServerPacket{}

	if err := protobuf.Unmarshal(data, packet); err != nil {
		fmt.Println("Error while decoding => proto.Unmarshal:", err)
		return
	}
	routePacket(session, packet)
}

func routePacket(session *service.Session, packet *proto.ServerPacket) {
	switch packet.GetPacketInfo().GetProcedure() {
	case proto.ProtocolProcedure_PROCEDURE_SERVER_HELLO:
		replyWorkerData(session, packet)

	case proto.ProtocolProcedure_PROCEDURE_REQUEST_TASK:
		processAppAction(session, packet)

	case proto.ProtocolProcedure_PROCEDURE_CHECK_TASK:
		replyCheckTask(session, packet)

	case proto.ProtocolProcedure_PROCEDURE_CANCEL_TASK:
		// TODO: implements cancel task logics

	default:
		fmt.Println("Unknown Procedure: ", packet.GetPacketInfo().GetProcedure())
		return
	}
}

func replyCheckTask(session *service.Session, packet *proto.ServerPacket) {
	resultObj, ok := currentTaskMap.Get(packet.GetPacketInfo().GetChallengeCode())
	workerPacket := proto.WorkerPacket{
		PacketInfo: packet.GetPacketInfo(),
	}

	if ok {
		workerPacket.TaskResult = resultObj
	} else {
		workerPacket.TaskResult = &proto.Result{
			TaskStarted:      false,
			ResultFinished:   false,
			ResultSuccessful: false,
		}
	}

	sendMessage(session, &workerPacket)
}

func replyWorkerData(session *service.Session, packet *proto.ServerPacket) {
	var requireDetailInfo = false

	switch packet.GetPacketInfo().GetActionProcedure() {
	case proto.ActionProcedure_PROCEDURE_NONE:
		fmt.Println("Connected to server, responding WorkerInfo...")
		requireDetailInfo = false
		session.RegisterServerId(packet.GetPacketInfo().GetExtraData())
	case proto.ActionProcedure_PROCEDURE_REQUEST_WORKER_INFO:
		requireDetailInfo = true
	case proto.ActionProcedure_PROCEDURE_ACK:
		fmt.Println("Server handshake done!")
		return
	}

	var responsePacket = proto.WorkerPacket{
		WorkerInfo: service.CreateDefaultWorkerInfo(requireDetailInfo),
		PacketInfo: packet.GetPacketInfo(),
		TaskResult: new(proto.Result{
			ResultFinished:   true,
			ResultSuccessful: true,
		}),
	}

	sendMessage(session, &responsePacket)
}

func sendMessage(session *service.Session, workerPacket *proto.WorkerPacket) {
	data, err := protobuf.Marshal(workerPacket)
	if err != nil {
		fmt.Println("Error while marshal responsePacket:", err)
		return
	}

	json, _ := protojson.Marshal(workerPacket)
	fmt.Println("Sending message: ", string(json))

	err = session.SendMessage(data)
	if err != nil {
		fmt.Println("Error while sending responsePacket:", err)
		return
	}
}
