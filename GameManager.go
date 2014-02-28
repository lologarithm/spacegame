package main

import (
	"fmt"
	"time"
)

// GameMessages come in. EntityUpdate objects goto physics. GameMessages go out to use goroutines to parse.
func ManageRequests(exit chan int, incoming_requests chan GameMessage, outgoingNetwork chan NetMessage) {
	sm := &SolarManager{ships: make(map[uint32]*Ship, 50), last_update: time.Now()}
	gm := &GameManager{Users: make(map[uint32]*Client, 100)}
	into_simulator := make(chan EntityUpdate, 512)
	//out_simulator := make(chan EntityUpdate, 512)
	//simulator := &SolarSimulator{output_update: out_simulator, Entities: map[uint32]Entity{}, Characters: map[uint32]Entity{}, last_update: time.Now()}
	//go simulator.RunSimulation(into_simulator)
	update_time := int64(0)
	update_count := 0
	for {
		timeout := sm.last_update.Add(time.Millisecond * 50).Sub(time.Now())
		wait_for_timeout := true
		for wait_for_timeout {
			select {
			case msg := <-incoming_requests:
				switch msg.(type) {
				case *LoginMessage:
					login_msg, _ := msg.(*LoginMessage)
					if login_msg.LoggingIn {
						HandleLogin(login_msg, gm, sm, into_simulator)
					} else {
						HandleLogoff(login_msg, gm, sm)
					}
				case *SetThrustMessage:
					HandleThrust(&msg, gm, sm)
				default:
					fmt.Println("GameManager.go:ManageRequests(): UNKNOWN MESSAGE TYPE")
				}
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			case <-exit:
				return
			}
		}
		sm.last_update = time.Now()
		for _, user := range gm.Users {
			temp_ships := []*Ship{}
			// TODO: Cache currently visible ships and recheck them every so often instead of re-creating?
			for _, ship := range sm.ships {
				// Check if player can detect ship?
				temp_ships = append(temp_ships, ship)
			}
			user.outgoing_messages <- &PhysicsUpdateMessage{Ships: temp_ships}
		}
		if len(gm.Users) > 0 {
			last_update_time := time.Now().Sub(sm.last_update).Nanoseconds()
			update_time += last_update_time
			update_count += 1

			if update_count%100 == 0 {
				fmt.Printf("  **   Last player messaging time: %d microseconds\n", last_update_time/1000)
				fmt.Printf("  **Average player messaging time: %d microseconds\n", (update_time / int64(update_count*1000)))
			}
		}
	}
}

func HandleLogin(msg *LoginMessage, gm *GameManager, sm *SolarManager, into_simulator chan EntityUpdate) {
	gm.Users[msg.FromUser] = msg.Client
	gm.Users[msg.FromUser].User = &User{Id: msg.FromUser}
	gm.Users[msg.FromUser].User.ActiveCharacter = &Character{EntityData: EntityData{Id: 0}}

	ship_id := uint32(len(sm.ships))
	ship := CreateShip(ship_id, "TestShip")

	sm.ships[ship_id] = ship
	eu := &EntityUpdate{UpdateType: 1, EntityObj: *ship}
	into_simulator <- *eu
}

func HandleLogoff(msg *LoginMessage, gm *GameManager, sm *SolarManager) {
	delete(gm.Users, msg.FromUser)
	// Ship removal?
}

func HandleThrust(msg *GameMessage, gm *GameManager, sm *SolarManager) {
	// TODO: Create ship designs that have angle of thruster
	fmt.Println("SETTING SOME THRUSTER CRAP")
}

// TODO: Check if ID already exists (logged off etc) and return that instead of creating.
func CreateShip(ship_id uint32, hull string) *Ship {
	return (&Ship{}).CreateTestShip(ship_id, hull)
}

type GameManager struct {
	// Player data
	Users      map[uint32]*Client
	Characters map[uint32]*Character
}

type SolarManager struct {
	characters  map[uint32]*Character
	ships       map[uint32]*Ship
	last_update time.Time
}
