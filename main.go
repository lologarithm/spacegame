package main

import (
	"fmt"
	"github.com/lologarithm/spacegame/server"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(1)
	exit := make(chan int, 1)
	toServerManager := make(chan server.GameMessage, 1024)
	outToNetwork := make(chan server.NetMessage, 1024)
	fmt.Println("Starting!")

	manager := server.NewServerManager()
	manager.FromNetwork = toServerManager
	manager.Exit = exit

	// Launch server manager
	go manager.Run()
	// Launch network manager
	go server.RunServer(exit, toServerManager, outToNetwork)
	fmt.Println("Server started. Press a key to exit.")
	fmt.Scanln()
	fmt.Println("Goodbye!")
	exit <- 1
	return
}
