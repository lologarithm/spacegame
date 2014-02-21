package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type NetMessageType byte

// 0: echo, 1: login_request 2: login success 3: login failure/logoff 4: physics update
const (
	ECHO         NetMessageType = 0
	LOGINREQUEST NetMessageType = 1
	LOGINSUCCESS NetMessageType = 2
	LOGINFAIL    NetMessageType = 3
	SETTHRUST    NetMessageType = 4
	DISCONNECT   NetMessageType = 255
)

type NetMessage struct {
	raw_bytes   []byte
	frame       *MessageFrame
	destination *Client
}

func (m *NetMessage) Content() []byte {
	return m.raw_bytes[m.frame.frame_length : m.frame.frame_length+m.frame.content_length]
}

type MessageFrame struct {
	message_type   NetMessageType
	from_user      int32
	frame_length   int16
	content_length int16
}

func ParseFrame(raw_bytes []byte) *MessageFrame {
	if len(raw_bytes) >= 7 {
		mf := new(MessageFrame)
		mf.message_type = NetMessageType(raw_bytes[0])
		var v int32
		binary.Read(bytes.NewBuffer(raw_bytes[1:5]), binary.LittleEndian, &v)
		mf.from_user = v
		var cl int16
		binary.Read(bytes.NewBuffer(raw_bytes[5:7]), binary.LittleEndian, &cl)
		mf.content_length = cl
		mf.frame_length = 7
		return mf
	}

	return nil
}

type Client struct {
	buffer         []byte
	client_address *net.UDPAddr
	incoming_bytes chan []byte
	User           *User
	quit           bool
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
						frame:     msg_frame,
						raw_bytes: client.buffer[0 : msg_frame.frame_length+msg_frame.content_length]}
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
