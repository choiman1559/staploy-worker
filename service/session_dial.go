package service

import (
	"context"
	"fmt"
	"time"

	"staploy-worker/consts"

	"github.com/coder/websocket"
)

type Config struct {
	Address    string `arg:"-a,required" help:"server address"`
	Port       int    `arg:"-p,required" help:"server port"`
	BinDir     string `arg:"--bin-dir,required" help:"path to binary directory"`
	ProfileDir string `arg:"--profile-dir" help:"path to profile directory"`
	CacheDir   string `arg:"--cache-dir" help:"path to cache directory"`
}

var WebSocketSession Session
var ArgsConfig *Config

func InitSession(a *Config, eventListener EventListener) {
	ArgsConfig = a
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var paths = fmt.Sprintf(consts.APIRouteSchema, "v1", consts.ConnTypeWorker)
	var addr = fmt.Sprintf("ws://%s:%d%s", ArgsConfig.Address, ArgsConfig.Port, paths)

	c, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		fmt.Printf("Failed to connect to server: %s\n", addr)
		panic(err)
	}

	WebSocketSession = Session{
		context:  ctx,
		conn:     c,
		listener: eventListener,
	}

	err = readLoop(&WebSocketSession)
	if err != nil {
		fmt.Printf("Failed to read data: %s\n", err)
	}

	defer func(s *Session) {
		err := s.CloseNow()
		if err != nil {
			fmt.Printf("Failed to close connection: %s\n", err)
			panic(err)
		}
	}(&WebSocketSession)
}

func readLoop(s *Session) error {
	for {
		_, data, err := s.conn.Read(s.context)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				fmt.Println("Connection closed")
				return nil
			}
			return err
		}

		fmt.Printf("Received raw message: %s\n", data)
		s.listener.OnIncomingData(s, data)
	}
}
