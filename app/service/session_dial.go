package service

import (
	"context"
	"fmt"
	"log"
	"staploy-worker/app/consts"
	"time"

	"github.com/coder/websocket"
)

const VERSION = "0.1.0"

type Config struct {
	Address    string `arg:"-a,required" help:"server address"`
	Port       int    `arg:"-p,required" help:"server port"`
	BaseDir    string `arg:"-d,--base-dir,required" help:"path to base binary directory"`
	ProfileDir string `arg:"--profile-dir" help:"overrides path to profile directory"`
	CacheDir   string `arg:"--cache-dir" help:"overrides path to cache directory"`

	BufferSize  int64 `arg:"--buffer-size" default:"65535" help:"overrides buffer size in bytes"`
	RemoteShell bool  `arg:"--remote-shell" help:"(Experimental) use remote shell"`
}

func (c *Config) Version() string {
	return fmt.Sprintf("staploy-worker %s (%s)", VERSION, GetWorkerCpuArch().String())
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
		log.Printf("Failed to connect to server: %s\n", addr)
		panic(err)
	}

	WebSocketSession = Session{
		context:  ctx,
		conn:     c,
		listener: eventListener,
	}

	err = readLoop(&WebSocketSession)
	if err != nil {
		log.Printf("Failed to read data: %s\n", err)
	}

	defer func(s *Session) {
		err := s.CloseNow()
		if err != nil {
			log.Printf("Failed to close connection: %s\n", err)
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

		log.Printf("Received raw message: %s\n", data)
		s.listener.OnIncomingData(s, data)
	}
}
