package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// TODO: Track 'reliable' messages. Decide which need to be resent.

type Client struct {
	buffer            []byte
	client_address    *net.UDPAddr
	incoming_bytes    chan []byte      // Bytes from client to server
	outgoing_messages chan GameMessage // GameMessages from GameManger to client
	User              *User            // User attached to this network client
	quit              bool
}

// Accepts raw bytes from a socket and turns them into NetMessage objects and then
// later into GameMessages
func (client *Client) ProcessBytes(toGameManager chan GameMessage, outgoing_msg chan NetMessage, disconnect_player chan Client) {
	client.quit = false
	for !client.quit {
		if dem_bytes, ok := <-client.incoming_bytes; !ok {
			break
		} else {
			client.buffer = append(client.buffer, dem_bytes...)
			msg_frame := ParseFrame(client.buffer)
			if msg_frame != nil && int(msg_frame.frame_length+msg_frame.content_length) >= len(client.buffer) {
				if msg_frame.message_type == ECHO {
					netmessage := &NetMessage{
						frame:       msg_frame,
						raw_bytes:   client.buffer[0 : msg_frame.frame_length+msg_frame.content_length],
						destination: client}
					outgoing_msg <- *netmessage
				} else {
					msg_obj := client.parseMessage(msg_frame)
					toGameManager <- msg_obj
					// Only handling to do in this method is the checking for disconnect message.
					switch msg_obj.(type) {
					case *LoginMessage:
						loginmsg, _ := msg_obj.(*LoginMessage)
						if !loginmsg.LoggingIn {
							disconnect_player <- *client
							fmt.Printf("Disconnected Player: %v\n", client.client_address)
							break
						}
					}
				}
				client.buffer = client.buffer[msg_frame.frame_length+msg_frame.content_length:]
			}
		}
	}
}

// Accepts input of raw bytes from a NetMessage. Parses and returns a
// GameMessage that the GameManager can use. Might want to separate each
// message type parser into the object?
func (client *Client) parseMessage(msg_frame *MessageFrame) GameMessage {
	content := client.buffer[msg_frame.frame_length : msg_frame.frame_length+msg_frame.content_length]
	gmv := &GameMessageValues{FromUser: msg_frame.from_user, Client: client}
	switch msg_frame.message_type {
	case LOGINREQUEST:
		password := string(content)
		// TODO: Check password? Lookup user? Maybe this should go to the game manager
		if password == "a" {
			msg := &LoginMessage{GameMessageValues: *gmv, LoggingIn: true}
			return msg
		}
	case SETTHRUST:
		//5 USER CLEN [T1 PERC, T2 PERC]
		num_percents := len(content) / 2
		thrust_percents := make([]int16, num_percents)
		for i := 0; i < num_percents; i++ {
			c_pos := i * 2
			binary.Read(bytes.NewBuffer(content[c_pos:c_pos+2]), binary.LittleEndian, thrust_percents[i])
		}
		msg := &SetThrustMessage{GameMessageValues: *gmv, ThrustPercent: thrust_percents}
		return msg
	case DISCONNECT:
		return &LoginMessage{GameMessageValues: *gmv, LoggingIn: false}
	}
	return nil
}

func (lm *LoginMessage) CreateLoginMessageBytes(seq uint16) *NetMessage {
	mt := LOGINSUCCESS
	if !success {
		mt = LOGINFAIL
	}
	m := &NetMessage{}
	m.frame = &MessageFrame{message_type: mt, content_length: 1}
	buf := new(bytes.Buffer)
	buf.Grow(10)
	buf.WriteByte(byte(mt))
	binary.Write(buf, binary.LittleEndian, uint16(seq)) // Write seq
	binary.Write(buf, binary.LittleEndian, uint16(1))   // Write 2 byte content len
	buf.WriteByte(mt == LOGINSUCCESS)                   // Content, 1=="success"
	m.raw_bytes = buf.Bytes()
	return m
}

func (sm *SolarManager) CreateShipUpdateMessage() (m NetMessage) {
	content_length := 20 * len(sm.ships)
	m.frame = &MessageFrame{message_type: 4, frame_length: 9, content_length: int16(content_length)}
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
