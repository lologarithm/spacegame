package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

func ManageRequests(exit chan int, incoming_requests chan Message, outgoing_player chan Message) {
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
				if msg.frame.message_type == 1 {
					outgoing_player <- HandleLogin(&msg, gm, sm, into_simulator)
				}
				if msg.frame.message_type == 255 {
					HandleLogoff(&msg, gm, sm)
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
		player_message := new(Message)
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

func HandleLogin(msg *Message, gm *GameManager, sm *SolarManager, into_simulator chan EntityUpdate) Message {
	gm.users[msg.frame.from_user] = msg.destination
	gm.users[msg.frame.from_user].user = &User{Id: msg.frame.from_user}
	ship_id := int32(len(sm.ships))
	ship := &Ship{Hull: "A", EntityData: EntityData{Id: ship_id}}
	sm.ships[ship_id] = ship
	eu := &EntityUpdate{UpdateType: msg.frame.message_type, EntityObj: *ship}
	into_simulator <- *eu
	success := true
	fmt.Println("Logged in: ", gm.users[msg.frame.from_user])
	m := CreateLoginMessage(gm.users[msg.frame.from_user].user, success)
	m.destination = msg.destination
	return *m
}

func HandleLogoff(msg *Message, gm *GameManager, sm *SolarManager) {
	delete(gm.users, msg.frame.from_user)
	// Ship removal?
}

func CreateLoginMessage(user *User, success bool) *Message {
	mt := byte(2)
	if !success {
		mt = byte(3)
	}
	m := &Message{}
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

func (sm *SolarManager) CreateShipUpdateMessage() (m Message) {
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

type GameMessage struct {
}
