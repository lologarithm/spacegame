package main

import "fmt"
import "runtime"

func main() {
	runtime.GOMAXPROCS(1)
	exit := make(chan int, 1)
	toGameManager := make(chan GameMessage, 1024)
	outToNetwork := make(chan NetMessage, 1024)
	fmt.Println("Starting!")
	go RunServer(exit, toGameManager, outToNetwork)
	go ManageRequests(exit, toGameManager)
	fmt.Println("Server started. Press a key to exit.")
	fmt.Scanln()
	fmt.Println("Goodbye!")
	exit <- 1
	return
}
