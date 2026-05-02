package main

import (
	"fmt"
	"os"
	"staploy-worker/process"
	"staploy-worker/service"

	"github.com/alexflint/go-arg"
)

type EventFunc func(*service.Session, []byte)

func (f EventFunc) OnIncomingData(s *service.Session, d []byte) { f(s, d) }

func main() {
	var args service.Config
	arg.MustParse(&args)
	fmt.Printf("Connecting to %s on port: %d\n", args.Address, args.Port)

	stat, err := os.Stat(args.BinDir)
	if err != nil {
		fmt.Printf("Failed to stat binary directory: %s\n", err)
		panic(err)
	} else if !stat.IsDir() {
		fmt.Printf("%s is not a directory\n", args.BinDir)
		panic(args.BinDir)
	}

	service.InitSession(&args, EventFunc(func(s *service.Session, data []byte) {
		go process.PacketProcess(s, data)
	}))
}
