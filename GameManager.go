package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// GameMessages come in. EntityUpdate objects goto physics. NetMessages go out to network.
func ManageRequests(exit chan int, incoming_requests chan GameMessage, outgoingNetwork chan NetMessage) {
	sm := &SolarManager{ships: make(map[int32]*Ship, 50), last_update: time.Now()}
	gm := &GameManager{Users: make(map[int32]*Client, 100)}
	into_simulator := make(chan EntityUpdate, 512)
	out_simulator := make(chan EntityUpdate, 512)
	simulator := &SolarSimulator{output_update: out_simulator, Entities: map[int32]Entity{}, Characters: map[int32]Entity{}}
	go simulator.RunSimulation(into_simulator)
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
					fmt.Println("GameManager got login.")
					login_msg, _ := msg.(*LoginMessage)
					if login_msg.LoggingIn {
						outgoingNetwork <- HandleLogin(login_msg, gm, sm, into_simulator)
					} else {
						HandleLogoff(login_msg, gm, sm)
					}
				case *SetThrustMessage:
					HandleThrust(&msg, gm, sm)
				default:
					fmt.Println("UNKNOWN MESSAGE TYPE")
				}
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			case <-exit:
				return
			}
		}
		sm.last_update = time.Now()
		base_message := sm.CreateShipUpdateMessage()
		player_message := new(NetMessage)
		for _, user := range gm.Users {
			*player_message = base_message
			player_message.destination = user
			outgoingNetwork <- *player_message
		}
		update_time += time.Now().Sub(sm.last_update).Nanoseconds()
		update_count += 1

		if update_count%100 == 0 {
			fmt.Printf("  **Average player messaging time: %d microseconds\n", (update_time / int64(update_count*1000)))
		}
	}
}

func HandleLogin(msg *LoginMessage, gm *GameManager, sm *SolarManager, into_simulator chan EntityUpdate) NetMessage {
	gm.Users[msg.FromUser] = msg.Client
	gm.Users[msg.FromUser].User = &User{Id: msg.FromUser}
	gm.Users[msg.FromUser].User.ActiveCharacter = &Character{EntityData: EntityData{Id: 0}}
	ship_id := int32(len(sm.ships))
	ship := CreateShip(ship_id, "A")
	sm.ships[ship_id] = ship
	eu := &EntityUpdate{UpdateType: 1, EntityObj: *ship}
	into_simulator <- *eu
	success := true
	fmt.Println("Logged in: ", gm.Users[msg.FromUser])
	m := CreateLoginMessage(gm.Users[msg.FromUser].User, success)
	m.destination = msg.Client
	return *m
}

func HandleLogoff(msg *LoginMessage, gm *GameManager, sm *SolarManager) {
	delete(gm.Users, msg.FromUser)
	// Ship removal?
}

func HandleThrust(msg *GameMessage, gm *GameManager, sm *SolarManager) {
	// TODO: Create ship designs that have angle of thruster

}

// TODO: Check if ID already exists (logged off etc) and return that instead of creating.
func CreateShip(ship_id int32, hull string) *Ship {
	return &Ship{Hull: "A", EntityData: EntityData{Id: ship_id},
		RigidBody: RigidBody{Mass: 2000, Position: Vect2{0, 0}, Velocity: Vect2{0, 0}, Force: Vect2{0, 0}}}
}

func CreateLoginMessage(user *User, success bool) *NetMessage {
	mt := byte(2)
	if !success {
		mt = byte(3)
	}
	m := &NetMessage{}
	m.frame = &MessageFrame{message_type: mt, frame_length: 9, content_length: 1}
	buf := new(bytes.Buffer)
	buf.Grow(10)
	buf.WriteByte(mt)
	binary.Write(buf, binary.LittleEndian, int32(user.Id)) // Write 4byte user id
	binary.Write(buf, binary.LittleEndian, int32(1))       // Write 4 byte content len
	buf.WriteByte(1)                                       // Content, 1=="success"
	m.raw_bytes = buf.Bytes()
	return m
}

type GameManager struct {
	// Player data
	Users      map[int32]*Client
	Characters map[int32]*Character
}

type SolarManager struct {
	characters  map[int32]*Character
	ships       map[int32]*Ship
	last_update time.Time
}

func (sm *SolarManager) CreateShipUpdateMessage() (m NetMessage) {
	content_length := 20 * len(sm.ships)
	m.frame = &MessageFrame{message_type: 4, frame_length: 9, content_length: int32(content_length)}
	buf := new(bytes.Buffer)
	buf.Grow(9 + content_length)
	buf.WriteByte(4)
	binary.Write(buf, binary.LittleEndian, int32(0))
	binary.Write(buf, binary.LittleEndian, int32(content_length))
	for _, ship := range sm.ships {
		buf.Write(ship.UpdateBytes())
	}
	m.raw_bytes = buf.Bytes()
	return
}

type GameMessage interface {
}

type GameMessageValues struct {
	FromUser int32
	Client   *Client
}

type LoginMessage struct {
	GameMessageValues
	LoggingIn bool
}

type SetThrustMessage struct {
	GameMessageValues
	ThrustPercent []int32
}
