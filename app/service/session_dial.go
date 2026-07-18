package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"staploy-worker/app/consts"
	"time"

	"github.com/coder/websocket"
)

type Config struct {
	Address    string `arg:"-a,env:STAPLOY_SERVER_ADDR,required" help:"server address"`
	Port       uint16 `arg:"-p,env:STAPLOY_SERVER_PORT,required" help:"server port"`
	BaseDir    string `arg:"-d,--base-dir,env:STAPLOY_DIR_BASE,required" help:"path to base binary directory"`
	ProfileDir string `arg:"--profile-dir,env:STAPLOY_DIR_PROFILE" help:"overrides path to profile directory"`
	CacheDir   string `arg:"--cache-dir,env:STAPLOY_DIR_CACHE" help:"overrides path to cache directory"`

	Enable_mTLS     bool   `arg:"--use-mtls,env:STAPLOY_MTLS_ENABLE" help:"use MTLS"`
	TLS_SERVER_CA   string `arg:"--mtls-server-ca,env:STAPLOY_MTLS_SERVER_CA" help:"path of mTLS server CA certificate (*.pem)"`
	TLS_CLIENT_CERT string `arg:"--mtls-cert,env:STAPLOY_MTLS_CERT" help:"path of mTLS client certificate (*.pem)"`
	TLS_CLIENT_KEY  string `arg:"--mtls-key,env:STAPLOY_MTLS_KEY" help:"path of mTLS client key (*.pem)"`

	OverrideName       string `arg:"--name-override,env:STAPLOY_WORKER_NAME" help:"override worker name rather then default machine name"`
	BufferSize         int64  `arg:"--buffer-size,env:STAPLOY_BUFFER_SIZE" default:"65535" help:"overrides buffer size in bytes"`
	RemoteShell        bool   `arg:"--remote-shell,env:STAPLOY_ENABLE_SHELL" help:"(Experimental) use remote shell"`
	DisableSymlinkDir  bool   `arg:"--disable-symlink-dir,env:STAPLOY_DISABLE_SYMDIR" help:"disable symlink version dir and directly create symlinks to files"`
	SkipHashValidCheck bool   `arg:"--skip-hash-verify,env:STAPLOY_SKIP_VERIFY" help:"skip hash verification when downloading package"`
	OverrideShellExec  string `arg:"--override-shell,env:STAPLOY_SHELL_PATH" help:"specify shell program rather than bash in $PATH"`
	MaxRetryOnDisconn  int8   `arg:"--max-retry,env:STAPLOY_MAX_RECONN" help:"maximum number of retry attempts to connect to server; -1 for infinite retry, 0 (Default) for no retry"`
	Verbose            bool   `arg:"-v,--verbose,env:STAPLOY_VERBOSE" help:"verbose output"`
}

func (c *Config) Version() string {
	return fmt.Sprintf("staploy-worker %s (%s)", consts.VERSION, GetWorkerCpuArch().String())
}

var WebSocketSession Session
var ArgsConfig *Config
var TlsConfig *tls.Config
var CurrentTryAttempt int8 = 0

func InitSession(a *Config, eventListener EventListener) {
	ArgsConfig = a

	if ArgsConfig.MaxRetryOnDisconn < -1 {
		log.Fatalf("max retry count must be greater than -1")
	}

	ctx := context.Background()
	customHTTPClient := &http.Client{}
	wsAddrPrefix := "ws"

	if a.Enable_mTLS {
		wsAddrPrefix = "wss"
		caCert, err := os.ReadFile(a.TLS_SERVER_CA)
		if err != nil {
			log.Fatalf("Failed to read server CA certification: %v", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		clientCert, err := tls.LoadX509KeyPair(a.TLS_CLIENT_CERT, a.TLS_CLIENT_KEY)
		if err != nil {
			log.Fatalf("Failed to load client pair key certification: %v", err)
		}

		TlsConfig = &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      caCertPool,
			MinVersion:   tls.VersionTLS13,
		}
		customHTTPClient.Transport = &http.Transport{TLSClientConfig: TlsConfig}
		log.Printf("Using mTLS certification %#v", TlsConfig.ClientAuth.String())
	}

	dialOpts := &websocket.DialOptions{
		HTTPClient: customHTTPClient,
	}

	var paths = fmt.Sprintf(consts.APIRouteSchema, "v1", consts.ConnTypeWorker)
	var addr = fmt.Sprintf("%s://%s:%d%s", wsAddrPrefix, ArgsConfig.Address, ArgsConfig.Port, paths)
	var isNotFirstTry = false

	for {
		if isNotFirstTry && ArgsConfig.MaxRetryOnDisconn != -1 && CurrentTryAttempt >= ArgsConfig.MaxRetryOnDisconn {
			if ArgsConfig.MaxRetryOnDisconn == 0 {
				log.Printf("Connection configuration specifies no retries (0). Terminating worker daemon...\n")
			} else {
				log.Printf("Reached maximum connection retry attempts (%d/%d). Terminating...\n",
					CurrentTryAttempt, ArgsConfig.MaxRetryOnDisconn)
			}
			break
		}

		isNotFirstTry = true
		c, _, err := websocket.Dial(ctx, addr, dialOpts)

		if err != nil {
			CurrentTryAttempt++

			if ArgsConfig.MaxRetryOnDisconn == -1 {
				log.Printf("Failed to connect to server: %s | Cause: %s [Infinite Retry Mode]\n", addr, err.Error())
			} else {
				log.Printf("Failed to connect to server: %s | Cause: %s (Attempt: %d/%d)\n",
					addr, err.Error(), CurrentTryAttempt, ArgsConfig.MaxRetryOnDisconn)
			}

			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Successfully established WebSocket connection to %s\n", addr)
		WebSocketSession = Session{
			context:  ctx,
			conn:     c,
			listener: eventListener,
		}

		err = readLoop(&WebSocketSession)
		if err != nil {
			log.Printf("Network stream read loop interrupted: %s\n", err.Error())
		}

		log.Println("Connection closed by remote server. Initiating cleanup...")
		if ArgsConfig.MaxRetryOnDisconn == 0 {
			log.Println("MaxRetryOnDisconn is 0. Terminating immediately without reconnection.")
			break
		}

		log.Println("Reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	log.Println("Staploy worker engine gracefully stopped. All kernel descriptors released.")
	os.Exit(0)
}

func IsDebug() bool {
	//goland:noinspection GoBoolExpressions
	return consts.VERSION == consts.VERSION_DEV
}

func readLoop(s *Session) error {
	CurrentTryAttempt = 0
	defer func() {
		go func() {
			_ = s.CloseNow()
		}()
	}()

	for {
		_, data, err := s.conn.Read(s.context)
		if err != nil {
			s.serverId = ""
			closureReason := websocket.CloseStatus(err)

			if closureReason == websocket.StatusNormalClosure {
				return nil
			}
			return err
		}

		if ArgsConfig.Verbose && IsDebug() {
			log.Printf("Received raw message: %s\n", data)
		}
		s.listener.OnIncomingData(s, data)
	}
}
