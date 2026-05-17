package service

import (
	"context"

	"github.com/coder/websocket"
)

type EventListener interface {
	OnIncomingData(session *Session, data []byte)
}

type Session struct {
	context  context.Context
	conn     *websocket.Conn
	serverId string
	listener EventListener
}

func (session *Session) CloseNow() error {
	if session.IsAlive() {
		return session.conn.CloseNow()
	}
	session.serverId = ""
	return nil
}

func (session *Session) SendMessage(data []byte) error {
	return session.conn.Write(context.Background(), websocket.MessageBinary, data)
}

func (session *Session) IsAlive() bool {
	if session.conn != nil {
		err := session.conn.Ping(session.context)
		if err == nil {
			return true
		}
	}
	return false
}

func (session *Session) RegisterServerId(uuid string) {
	session.serverId = uuid
}

func (session *Session) IsConnected() bool {
	return session.ServerId() != ""
}

func (session *Session) ServerId() string {
	return session.serverId
}
