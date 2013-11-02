package main

import "fmt"
import "time"

func ManageRequests(incoming_requests chan Message) {
	for {
		select {
		case msg := <-incoming_requests:
			fmt.Println("MESSAGE:", msg)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

type GameManager struct {
	// Player data
	users map[int32]User
}

type SolarManager struct {
	characters map[int32]Character
	ships      map[int32]Ship
}
