package app

import (
	"log"
	"staploy-worker/app/files"
	"staploy-worker/app/process"
	"staploy-worker/app/service"

	"github.com/alexflint/go-arg"
)

type EventFunc func(*service.Session, []byte)

func (f EventFunc) OnIncomingData(s *service.Session, d []byte) { f(s, d) }

func StartApplication() {
	var args service.Config
	arg.MustParse(&args)

	if args.RemoteShell {
		log.Printf("WARNING: Starting application with REMOTE_SHELL_EXECUTION.")
		log.Printf("			Use this option with VERY caution.")
	}
	log.Printf("Connecting to %s on port: %d\n", args.Address, args.Port)

	files.SetConfig(files.IoConfig{
		BaseDir:    args.BaseDir,
		CacheDir:   args.CacheDir,
		ProfileDir: args.ProfileDir,
		BufferSize: args.BufferSize,
	})

	fil, err := files.GetBaseDir()
	if err != nil {
		log.Printf("Failed to stat base directory: %s\n", err)
		panic(err)
	}

	stat, err := fil.Stat()
	if err != nil || !stat.IsDir() {
		log.Printf("%s is not a directory\n", args.BaseDir)
		panic(args.BaseDir)
	}

	service.InitSession(&args, EventFunc(func(s *service.Session, data []byte) {
		go process.PacketProcess(s, data)
	}))
}
