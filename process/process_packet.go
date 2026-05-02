package process

import (
	"fmt"
	"staploy-worker/proto"
	"staploy-worker/service"

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
	case proto.ProtocolProcedure_PROCEDURE_CHECK_TASK:
	case proto.ProtocolProcedure_PROCEDURE_CANCEL_TASK:
	default:
		fmt.Println("Unknown Procedure: ", packet.GetPacketInfo().GetProcedure())
		return
	}
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

	data, err := protobuf.Marshal(&responsePacket)
	if err != nil {
		fmt.Println("Error while marshal responsePacket:", err)
		return
	}

	json, _ := protojson.Marshal(&responsePacket)
	fmt.Println("Sending message: ", string(json))

	err = session.SendMessage(data)
	if err != nil {
		fmt.Println("Error while sending responsePacket:", err)
		return
	}
}
