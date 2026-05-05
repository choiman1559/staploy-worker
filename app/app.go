package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"staploy-worker/app/files"
	"staploy-worker/app/process"
	"staploy-worker/app/service"
	"syscall"

	"github.com/alexflint/go-arg"
)

type EventFunc func(*service.Session, []byte)

func (f EventFunc) OnIncomingData(s *service.Session, d []byte) { f(s, d) }

func StartApplication() {
	var args service.Config
	arg.MustParse(&args)

	files.SetConfig(files.IoConfig{
		BaseDir:    args.BaseDir,
		CacheDir:   args.CacheDir,
		ProfileDir: args.ProfileDir,
		BufferSize: args.BufferSize,
	})

	if args.RemoteShell {
		log.Printf("WARNING: Starting application with REMOTE_SHELL_EXECUTION.")
		log.Printf("		Use this option with VERY caution.")
	}

	fil, err := files.GetBaseDir()
	if err != nil {
		log.Printf("Failed to stat base directory: %s\n", err)
		log.Fatal(err)
	}

	stat, err := fil.Stat()
	if err != nil || !stat.IsDir() {
		log.Fatalf("%s is not a directory\n", args.BaseDir)
	}

	log.Printf("Starting application. Worker ID: %s", service.GetWorkerUniqueId())
	log.Printf("Connecting to %s on port: %d\n", args.Address, args.Port)

	locked, err := service.GetWorkerUUIDLock().TryLock()
	if !locked || err != nil {
		log.Fatal("FATAL: You are already running a worker on same directory.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		service.InitSession(&args, EventFunc(func(s *service.Session, data []byte) {
			go process.PacketProcess(s, data)
		}))
	}()

	<-ctx.Done()
	err = service.GetWorkerUUIDLock().Unlock()
	if err != nil {
		return
	}
}
