package main

import (
	"fmt"
	"time"
)

type ServerManager struct {
	// Player data
	Users       map[uint32]*Client
	Games       map[uint32]*GameManager
	FromNetwork chan GameMessage
	Exit        chan int
}

// TODO: Create a 'server manager' that is the main director. It will let players
// join into a game
func (sm *ServerManager) Run() {
	// 1. List of players connected.
	// 2. List of active games
	// 3. Ability to create new game
	// 4. Ability to join existing games.

	// Players can create a game creating a new GameManager to run it.
	// Each gamemanager will run its own game. It will have a channel for messages
	// going into that game.
	//gameManager := &GameManager{Users: make(map[uint32]*Client, 100), IntoSimulator: make(chan EntityUpdate, 512), OutSimulator: make(chan EntityUpdate, 512)}
}

// GameMessages come in. EntityUpdate objects goto physics. GameMessages go out to use goroutines to parse.
func (gameManager *GameManager) ManageRequests() {
	solarManager := &SolarManager{ships: make(map[uint32]*Ship, 50), last_update: time.Now()}
	simulator := &SolarSimulator{outSimulator: gameManager.OutSimulator, intoSimulator: gameManager.IntoSimulator, Entities: map[uint32]Entity{}, Characters: map[uint32]Entity{}, lastUpdate: time.Now()}
	go simulator.RunSimulation()
	update_time := int64(0)
	update_count := 0
	wait_for_timeout := true
	// TODO: Ticker should be allocated once and just reset instead of creating new time.After
	for {
		timeout := solarManager.last_update.Add(time.Millisecond * 50).Sub(time.Now())
		wait_for_timeout = true
		for wait_for_timeout {
			select {
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			case msg := <-gameManager.FromNetwork:
				fmt.Printf("GameManager: Received message: %T\n", msg)
				switch msg.(type) {
				case *LoginMessage:
					l_msg, _ := msg.(*LoginMessage)
					if l_msg.LoggingIn {
						HandleLogin(l_msg, gameManager, solarManager)
					} else {
						HandleLogoff(l_msg, gameManager, solarManager)
					}
				case *SetThrustMessage:
					HandleThrust(&msg, gameManager, solarManager)
				default:
					fmt.Println("GameManager.go:ManageRequests(): UNKNOWN MESSAGE TYPE: %T", msg)
				}
			case msg := <-gameManager.OutSimulator:
				fmt.Printf("Physics data from server.\n")
				HandlePhysicsUpdate(&msg, solarManager)
			case <-exit:
				fmt.Println("EXITING MANAGER")
				return
			}
		}
		fmt.Printf("Sending client update!\n")
		solarManager.last_update = time.Now()
		for _, user := range gameManager.Users {
			temp_ships := []*Ship{}
			// TODO: Cache currently visible ships and recheck them every so often instead of re-creating?
			// TODO: Actually calculate vision/sensor range
			for _, ship := range solarManager.ships {
				// Check if player can detect ship?
				temp_ships = append(temp_ships, ship)
			}
			user.fromGameManager <- PhysicsUpdateMessage{Ships: temp_ships}
		}
		if len(gameManager.Users) > 0 {
			last_update_time := time.Now().Sub(solarManager.last_update).Nanoseconds()
			update_time += last_update_time
			update_count += 1

			if update_count%100 == 0 {
				fmt.Printf("  **   Last player messaging time: %d microseconds\n", last_update_time/1000)
				fmt.Printf("  **Average player messaging time: %d microseconds\n", (update_time / int64(update_count*1000)))
			}
		}
	}
}

func HandleLogin(msg *LoginMessage, gm *GameManager, sm *SolarManager) {
	gm.Users[msg.FromUser] = msg.Client
	gm.Users[msg.FromUser].User = &User{Id: msg.FromUser}
	gm.Users[msg.FromUser].User.ActiveCharacter = &Character{EntityData: EntityData{Id: uint32(len(gm.Users))}}

	ship_id := uint32(len(sm.ships))
	ship := CreateShip(ship_id, "TestShip")

	sm.ships[ship_id] = ship
	eu := &EntityUpdate{UpdateType: 1, EntityObj: *ship}
	gm.IntoSimulator <- *eu

	gm.Users[msg.FromUser].User.ActiveCharacter.CurrentShip = ship
}

func HandleLogoff(msg *LoginMessage, gm *GameManager, sm *SolarManager) {
	delete(gm.Users, msg.FromUser)
	// Ship removal?
}

func HandleThrust(msg *GameMessage, gm *GameManager, sm *SolarManager) {
	st_msg, ok := (*msg).(*SetThrustMessage)
	if !ok {
		// TODO: Error handling?
		return
	}
	current_char := gm.Users[st_msg.FromUser].User.ActiveCharacter
	if current_char.CurrentShip == nil {
		// TODO: Check if player is 'pilot'
		// TODO: Warn player he isnt flying a ship?
		return
	}

	final_force := Vect2{0, 0}
	final_torque := float32(0.0)
	var t Thruster
	for ind, thm := range st_msg.ThrustPercent {
		if len(current_char.CurrentShip.Hull.Thrusters) <= ind {
			// Handle error case of too many thrusters being set.
			break
		}
		t = current_char.CurrentShip.Hull.Thrusters[ind]
		t.Current = t.Max * (float32(thm) / 100.0)

		// Multipy linearvector by force to get total thrust vector.
		final_force = final_force.Add(MultVect2(t.LinearVector, t.Current))
		final_torque += t.AngularPercent * t.Current
	}

	current_char.CurrentShip.Force = final_force
	current_char.CurrentShip.Torque = final_torque
	eu := &EntityUpdate{UpdateType: UpdateForces, EntityObj: *current_char.CurrentShip}
	gm.IntoSimulator <- *eu
}

func HandlePhysicsUpdate(msg *EntityUpdate, sm *SolarManager) {
	switch msg.EntityObj.(type) {
	case Ship:
		ship, _ := msg.EntityObj.(Ship)
		if msg.UpdateType == UpdatePosition {
			sm.ships[ship.Id].Angle = ship.Angle
			sm.ships[ship.Id].Position = ship.Position
			sm.ships[ship.Id].Velocity = ship.Velocity
			sm.ships[ship.Id].AngularVelocity = ship.AngularVelocity
		}
	}
}

// TODO: Check if ID already exists (logged off etc) and return that instead of creating.
func CreateShip(ship_id uint32, hull string) *Ship {
	return (&Ship{}).CreateTestShip(ship_id, hull)
}

type GameManager struct {
	// Player data
	Users         map[uint32]*Client
	Characters    map[uint32]*Character
	IntoSimulator chan EntityUpdate
	OutSimulator  chan EntityUpdate
	FromNetwork   chan GameMessage
	Exit          chan int
}

type SolarManager struct {
	characters  map[uint32]*Character
	ships       map[uint32]*Ship
	last_update time.Time
}
