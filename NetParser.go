package main

import (
	"bytes"
	"encoding/binary"
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
	frame_length   int32
	content_length int32
}

func ParseFrame(raw_bytes []byte) *MessageFrame {
	if len(raw_bytes) >= 9 {
		mf := new(MessageFrame)
		mf.message_type = raw_bytes[0]
		var v int32
		binary.Read(bytes.NewBuffer(raw_bytes[1:5]), binary.LittleEndian, &v)
		mf.from_user = v
		binary.Read(bytes.NewBuffer(raw_bytes[5:9]), binary.LittleEndian, &v)
		mf.content_length = v
		mf.frame_length = 9
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
					}
				}
			}
		}
	}
}

// Accepts input of raw bytes from a NetMessage. Parses and returns a
// GameMessage that the GameManager can use.
func (client *Client) parseMessage(msg_frame *MessageFrame) GameMessage {
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
	case 255:
		return &LoginMessage{GameMessageValues: *gmv, LoggingIn: false}
	}
	return nil
}
