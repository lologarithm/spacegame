package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
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
	message_type   byte // 0: echo, 1: login_request 2: login success 3: login failure/logoff 4: physics update
	from_user      int32
	frame_length   int16
	content_length int16
}

func ParseFrame(raw_bytes []byte) *MessageFrame {
	if len(raw_bytes) >= 7 {
		mf := new(MessageFrame)
		mf.message_type = raw_bytes[0]
		var v int32
		binary.Read(bytes.NewBuffer(raw_bytes[1:5]), binary.LittleEndian, &v)
		mf.from_user = v
		var cl int16
		binary.Read(bytes.NewBuffer(raw_bytes[5:9]), binary.LittleEndian, &cl)
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
			fmt.Printf("Current Buffer: %v\n", client.buffer)
			if msg_frame != nil && int(msg_frame.frame_length+msg_frame.content_length) >= len(client.buffer) {
				if msg_frame.message_type == 0 {
					netmessage := &NetMessage{
						frame:     msg_frame,
						raw_bytes: client.buffer[0 : msg_frame.frame_length+msg_frame.content_length]}
					outgoing_msg <- *netmessage
				} else {
					msg_obj := client.parseMessage(msg_frame)
					toGameManager <- msg_obj
					switch msg_obj.(type) {
					case *LoginMessage:
						loginmsg, _ := msg_obj.(*LoginMessage)
						if !loginmsg.LoggingIn {
							disconnect_player <- *client
							break
						}
						fmt.Println("Disconnected Player: %v", client.client_address)
					}
				}
				client.buffer = client.buffer[msg_frame.frame_length+msg_frame.content_length:]
			}
		}
	}
}

// Accepts input of raw bytes from a NetMessage. Parses and returns a
// GameMessage that the GameManager can use.
func (client *Client) parseMessage(msg_frame *MessageFrame) GameMessage {
	fmt.Printf("Message: %v  Buffer: %v\n", msg_frame, client.buffer)
	content := client.buffer[msg_frame.frame_length : msg_frame.frame_length+msg_frame.content_length]
	gmv := &GameMessageValues{FromUser: msg_frame.from_user, Client: client}
	switch msg_frame.message_type {
	case 1:
		password := string(content)
		if password == "a" {
			msg := &LoginMessage{GameMessageValues: *gmv, LoggingIn: true}
			return msg
		}
	case 5:
		//5 USER CLEN [T1 PERC, T2 PERC]
		num_percents := len(content) / 4
		thrust_percents := make([]int16, num_percents)
		for i := 0; i < num_percents; i++ {
			c_pos := i * 4
			binary.Read(bytes.NewBuffer(content[c_pos:c_pos+4]), binary.LittleEndian, thrust_percents[i])
		}
		msg := &SetThrustMessage{GameMessageValues: *gmv, ThrustPercent: thrust_percents}
		return msg
	case 255:
		return &LoginMessage{GameMessageValues: *gmv, LoggingIn: false}
	}
	return nil
}
