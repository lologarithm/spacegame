package server

import (
	"fmt"
	"time"
)

const (
	CreateGame  = byte(1)
	GameInLobby = byte(2)
	JoinGame    = byte(3)
	QuitGame    = byte(4)
	GameRunning = byte(5)
	GameEnded   = byte(6)
)

// Manages all connected users and games.
type ServerManager struct {
	// Player data
	Clients     map[uint32]*Client
	Games       map[uint32]*GameManager
	NextGameId  uint32
	FromNetwork chan GameMessage
	FromGames   chan GameMessage
	Exit        chan int
}

// Manages players and ships for a single game.
type GameManager struct {
	// Player data
	Clients           map[uint32]*Client
	IntoSimulator     chan EntityUpdate // Channel for sending to physics sim for this game
	OutSimulator      chan EntityUpdate // Channel for updates from physics sim
	IntoServerManager chan GameMessage
	FromNetwork       chan GameMessage // Messages from players.
	Exit              chan int
	Solar             *SolarManager
	Status            byte
}

// Manages ships/physics for a single solar system in a game.
type SolarManager struct {
	Characters map[uint32]*Character
	Ships      map[uint32]*Ship
	Simulator  *SolarSimulator
	lastUpdate time.Time
}

func (sm *ServerManager) Run() {

	for {
		select {
		case netMsg := <-sm.FromNetwork:
			sm.ProcessNetMsg(netMsg)
		case gMsg := <-sm.FromGames:
			sm.ProcessGameMsg(gMsg)
		case <-sm.Exit:
			for _, game := range sm.Games {
				game.Exit <- 1
			}
			return
		}
	}
	// 1. List of players connected.
	//   a. Players able to login/logout
	// 2. List of active games
	//   a. Ability to create new game
	//   b. Ability to join existing games.

	// Players can create a game creating a new GameManager to run it.
	// Each gamemanager will run its own game. It will have a channel for messages
	// going into that game.
	//
}

func (sm *ServerManager) ProcessNetMsg(msg GameMessage) {
	switch msg.(type) {
	case *LoginMessage:
		loginMsg, _ := msg.(*LoginMessage)
		if loginMsg.LoggingIn {
			sm.HandleLogin(loginMsg)
		} else {
			sm.HandleLogoff(loginMsg)
		}
	case *PlayerGameMessage:
		pgMsg, _ := msg.(*PlayerGameMessage)
		switch pgMsg.GameAction {
		case CreateGame:
			newGame := &GameManager{
				Clients:           make(map[uint32]*Client, 100),
				IntoSimulator:     make(chan EntityUpdate, 512),
				OutSimulator:      make(chan EntityUpdate, 512),
				FromNetwork:       make(chan GameMessage, 100),
				IntoServerManager: sm.FromGames,
			}
			pgMsg.Client.toGameManager = newGame.FromNetwork
			go newGame.RunGame()
			sm.Games[sm.NextGameId] = newGame
			sm.NextGameId += 1
			newGame.FromNetwork <- msg
			if sm.NextGameId > 100000 {
				sm.NextGameId = 1
			}
		case JoinGame:
			if game, ok := sm.Games[pgMsg.GameId]; ok {
				game.FromNetwork <- msg
			}
		case QuitGame:
			if game, ok := sm.Games[pgMsg.GameId]; ok {
				game.FromNetwork <- msg
			}
		}
	}
}

func (sm *ServerManager) ProcessGameMsg(msg GameMessage) {
	switch msg.(type) {
	case *GameStatusMessage:
		gStatusMsg, _ := msg.(*GameStatusMessage)
		switch gStatusMsg.GameStatus {
		case GameInLobby:
		case GameRunning:
		case GameEnded:
			delete(sm.Games, gStatusMsg.GameId)
		}
	}
}

func (sm *ServerManager) HandleLogin(msg *LoginMessage) {
	sm.Clients[msg.FromUser] = msg.Client
	sm.Clients[msg.FromUser].User = &User{Id: msg.FromUser}
	//gm.Users[msg.FromUser].User.ActiveCharacter = &Character{EntityData: EntityData{Id: uint32(len(gm.Users))}}
}

func (sm *ServerManager) HandleLogoff(msg *LoginMessage) {
	delete(sm.Clients, msg.FromUser)
	// Ship removal?
}

// GameMessages come in. EntityUpdate objects goto physics. GameMessages go out to use goroutines to parse.
func (gameManager *GameManager) RunGame() {
	solarManager := &SolarManager{Ships: make(map[uint32]*Ship, 50), lastUpdate: time.Now()}

	simulator := &SolarSimulator{
		outSimulator: gameManager.OutSimulator, intoSimulator: gameManager.IntoSimulator,
		Entities: map[uint32]Entity{}, Characters: map[uint32]Entity{}, lastUpdate: time.Now(),
	}
	solarManager.Simulator = simulator
	go simulator.RunSimulation()
	update_time := int64(0)
	update_count := 0
	wait_for_timeout := true
	// TODO: Ticker should be allocated once and just reset instead of creating new time.After
	for {
		timeout := solarManager.lastUpdate.Add(time.Millisecond * 50).Sub(time.Now())
		wait_for_timeout = true
		for wait_for_timeout {
			select {
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			case msg := <-gameManager.FromNetwork:
				fmt.Printf("GameManager: Received message: %T\n", msg)
				switch msg.(type) {
				case *SetThrustMessage:
					gameManager.HandleThrust(&msg)
				default:
					fmt.Println("GameManager.go:ManageRequests(): UNKNOWN MESSAGE TYPE: %T", msg)
				}
			case msg := <-gameManager.OutSimulator:
				fmt.Printf("Physics data from server.\n")
				HandlePhysicsUpdate(&msg, solarManager)
			case <-gameManager.Exit:
				fmt.Println("EXITING MANAGER")
				return
			}
		}
		fmt.Printf("Sending client update!\n")
		solarManager.lastUpdate = time.Now()
		for _, user := range gameManager.Clients {
			temp_ships := []*Ship{}
			// TODO: Cache currently visible ships and recheck them every so often instead of re-creating?
			// TODO: Actually calculate vision/sensor range
			for _, ship := range solarManager.Ships {
				// Check if player can detect ship?
				temp_ships = append(temp_ships, ship)
			}
			user.fromGameManager <- PhysicsUpdateMessage{Ships: temp_ships}
		}
		if len(gameManager.Clients) > 0 {
			last_update_time := time.Now().Sub(solarManager.lastUpdate).Nanoseconds()
			update_time += last_update_time
			update_count += 1

			if update_count%100 == 0 {
				fmt.Printf("  **   Last player messaging time: %d microseconds\n", last_update_time/1000)
				fmt.Printf("  **Average player messaging time: %d microseconds\n", (update_time / int64(update_count*1000)))
			}
		}
	}
}

func (gm *GameManager) HandleGameStart() {

	for index, client := range gm.Clients {
		client.User.ActiveCharacter = &Character{}
		ship_id := uint32(index)
		ship := CreateShip(ship_id, "TestShip")

		gm.Solar.Ships[ship_id] = ship
		eu := &EntityUpdate{UpdateType: 1, EntityObj: *ship}
		gm.IntoSimulator <- *eu

		client.User.ActiveCharacter.CurrentShip = ship
	}
}

func (gm *GameManager) HandleThrust(msg *GameMessage) {
	st_msg, ok := (*msg).(*SetThrustMessage)
	if !ok {
		// TODO: Error handling? This should never happen..
		return
	}
	current_char := gm.Clients[st_msg.FromUser].User.ActiveCharacter
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
			sm.Ships[ship.Id].Angle = ship.Angle
			sm.Ships[ship.Id].Position = ship.Position
			sm.Ships[ship.Id].Velocity = ship.Velocity
			sm.Ships[ship.Id].AngularVelocity = ship.AngularVelocity
		}
	}
}

// TODO: Check if ID already exists (logged off etc) and return that instead of creating.
func CreateShip(ship_id uint32, hull string) *Ship {
	return (&Ship{}).CreateTestShip(ship_id, hull)
}
