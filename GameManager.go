package main

import "fmt"
import "time"

func ManageRequests(exit chan int, incoming_requests chan Message, outgoing_player chan Message) {
	sm := &SolarManager{}
	gm := &GameManager{users: make(map[int32]User, 100)}
	into_simulator := make(chan EntityUpdate, 200)
	out_simulator := make(chan EntityUpdate, 200)
	simulator := &SolarSimulator{output_update: out_simulator}
	go simulator.RunSimulation(into_simulator)
	for {
		update_timer := time.After(sm.next_update.Sub(time.Now()))
		sm.next_update = time.Now().Add(time.Millisecond * 20)
		for {
			select {
			case msg := <-incoming_requests:
				fmt.Println("MESSAGE:", msg)
			case <-update_timer:
				break
			}
		}
		update_messages := make([]Message, len(sm.ships))
		index := 0
		for _, ship := range sm.ships {
			update_messages[index] = ship.CreateUpdateMessage()
		}
		for _, user := range gm.users {
			dest := &Client{user: user}
			for _, msg := range update_messages {
				msg.destination = dest
				outgoing_player <- msg
			}
		}
	}
}

type GameManager struct {
	// Player data
	users map[int32]User
}

type SolarManager struct {
	characters  map[int32]Character
	ships       map[int32]Ship
	next_update time.Time
}
