package service

import (
	"context"
	"fmt"
	"log"
	"staploy-worker/app/consts"

	"github.com/coder/websocket"
)

type Config struct {
	Address    string `arg:"-a,required" help:"server address"`
	Port       int    `arg:"-p,required" help:"server port"`
	BaseDir    string `arg:"-d,--base-dir,required" help:"path to base binary directory"`
	ProfileDir string `arg:"--profile-dir" help:"overrides path to profile directory"`
	CacheDir   string `arg:"--cache-dir" help:"overrides path to cache directory"`

	BufferSize         int64  `arg:"--buffer-size" default:"65535" help:"overrides buffer size in bytes"`
	RemoteShell        bool   `arg:"--remote-shell" help:"(Experimental) use remote shell"`
	DisableSymlinkDir  bool   `arg:"--disable-symlink-dir" help:"disable symlink version dir and directly create symlinks to files"`
	SkipHashValidCheck bool   `arg:"--skip-hash-verify" help:"skip hash verification when downloading package"`
	OverrideShellExec  string `arg:"--override-shell" help:"specify shell program rather than bash in $PATH"`
	Verbose            bool   `arg:"-v,--verbose" help:"verbose output"`
}

func (c *Config) Version() string {
	return fmt.Sprintf("staploy-worker %s (%s)", consts.VERSION, GetWorkerCpuArch().String())
}

var WebSocketSession Session
var ArgsConfig *Config

func InitSession(a *Config, eventListener EventListener) {
	ArgsConfig = a
	ctx := context.Background()

	var paths = fmt.Sprintf(consts.APIRouteSchema, "v1", consts.ConnTypeWorker)
	var addr = fmt.Sprintf("ws://%s:%d%s", ArgsConfig.Address, ArgsConfig.Port, paths)

	c, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %s\n", addr)
	}

	WebSocketSession = Session{
		context:  ctx,
		conn:     c,
		listener: eventListener,
	}

	err = readLoop(&WebSocketSession)
	if err != nil {
		log.Fatalf("Failed to read data: %s\n", err)
	}

	defer func(s *Session) {
		err := s.CloseNow()
		if err != nil {
			log.Fatalf("Failed to close connection: %s\n", err)
		}
	}(&WebSocketSession)
}

func IsDebug() bool {
	//goland:noinspection GoBoolExpressions
	return consts.VERSION == consts.VERSION_DEV
}

func readLoop(s *Session) error {
	for {
		_, data, err := s.conn.Read(s.context)
		if err != nil {
			s.serverId = ""
			closureReason := websocket.CloseStatus(err)

			if closureReason == websocket.StatusNormalClosure {
				log.Fatalf("Connection closed")
			}
			return err
		}

		if ArgsConfig.Verbose && IsDebug() {
			log.Printf("Received raw message: %s\n", data)
		}
		s.listener.OnIncomingData(s, data)
	}
}
