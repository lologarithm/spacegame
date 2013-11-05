package main

import "fmt"
import "runtime"

func main() {
	runtime.GOMAXPROCS(1)
	exit := make(chan int, 1)
	incoming_requests := make(chan Message, 200)
	outgoing_player := make(chan Message, 1024)
	fmt.Println("Starting!")
	go RunServer(exit, incoming_requests, outgoing_player)
	go ManageRequests(exit, incoming_requests, outgoing_player)
	fmt.Println("Server started. Press a key to exit.")
	fmt.Scanln()
	fmt.Println("Goodbye!")
	exit <- 1
	return
}
