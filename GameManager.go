package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// GameMessages come in. EntityUpdate objects goto physics. NetMessages go out to network.
func ManageRequests(exit chan int, incoming_requests chan GameMessage, outgoing_player chan NetMessage) {
	sm := &SolarManager{ships: make(map[int32]*Ship, 50), last_update: time.Now()}
	gm := &GameManager{users: make(map[int32]*Client, 100)}
	into_simulator := make(chan EntityUpdate, 512)
	out_simulator := make(chan EntityUpdate, 512)
	simulator := &SolarSimulator{output_update: out_simulator, Entities: map[int32]Entity{}, Characters: map[int32]Entity{}}
	go simulator.RunSimulation(into_simulator)
	update_time := int64(0)
	update_count := 0
	for {
		timeout := sm.last_update.Add(time.Millisecond * 20).Sub(time.Now())
		wait_for_timeout := true
		for wait_for_timeout {
			select {
			case msg := <-incoming_requests:
				switch msg.(type) {
				case LoginMessage:
					login_msg, _ := msg.(LoginMessage)
					if login_msg.LoggingIn {
						outgoing_player <- HandleLogin(&login_msg, gm, sm, into_simulator)
					} else {
						HandleLogoff(&login_msg, gm, sm)
					}
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
		for _, user := range gm.users {
			*player_message = base_message
			player_message.destination = user
			outgoing_player <- *player_message
		}
		update_time += time.Now().Sub(sm.last_update).Nanoseconds()
		update_count += 1

		if update_count%100 == 0 {
			fmt.Printf("  **Average player messaging time: %d microseconds\n", (update_time / int64(update_count*1000)))
		}
	}
}

func HandleLogin(msg *LoginMessage, gm *GameManager, sm *SolarManager, into_simulator chan EntityUpdate) NetMessage {
	gm.users[msg.FromUser] = msg.Client
	gm.users[msg.FromUser].user = &User{Id: msg.FromUser}
	ship_id := int32(len(sm.ships))
	ship := &Ship{Hull: "A", EntityData: EntityData{Id: ship_id}}
	sm.ships[ship_id] = ship
	eu := &EntityUpdate{UpdateType: 1, EntityObj: *ship}
	into_simulator <- *eu
	success := true
	fmt.Println("Logged in: ", gm.users[msg.FromUser])
	m := CreateLoginMessage(gm.users[msg.FromUser].user, success)
	m.destination = msg.Client
	return *m
}

func HandleLogoff(msg *LoginMessage, gm *GameManager, sm *SolarManager) {
	delete(gm.users, msg.FromUser)
	// Ship removal?
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
	binary.Write(buf, binary.LittleEndian, int32(user.Id))
	binary.Write(buf, binary.LittleEndian, int32(1))
	buf.WriteByte(1)
	m.raw_bytes = buf.Bytes()
	return m
}

type GameManager struct {
	// Player data
	users map[int32]*Client
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
