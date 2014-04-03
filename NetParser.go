package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

// TODO: Track 'reliable' messages. Decide which need to be resent.

type Client struct {
	buffer            []byte
	client_address    *net.UDPAddr
	incoming_bytes    chan []byte      // Bytes from client to server
	outgoing_messages chan GameMessage // GameMessages from GameManger to client
	User              *User            // User attached to this network client
	Seq               uint16
	quit              bool
}

// Accepts raw bytes from a socket and turns them into NetMessage objects and then
// later into GameMessages. These are passed into the GameManager. This function also
// accepts outgoing messages from the GameManager to the client.
func (client *Client) ProcessBytes(toGameManager chan GameMessage, outgoing_msg chan NetMessage, disconnect_player chan Client) {
	client.quit = false
	for !client.quit {
		select {
		case dem_bytes, ok := <-client.incoming_bytes:
			if !ok {
				break
			} else {
				client.buffer = append(client.buffer, dem_bytes...)
				msg_frame := ParseFrame(client.buffer)
				if msg_frame != nil && int(msg_frame.frame_length+msg_frame.content_length) <= len(client.buffer) {
					fmt.Println("Handling message: %v\n", msg_frame)
					if msg_frame.message_type == ECHO {
						netmessage := &NetMessage{
							frame:       msg_frame,
							raw_bytes:   client.buffer[0 : msg_frame.frame_length+msg_frame.content_length],
							destination: client}
						outgoing_msg <- *netmessage
					} else {
						msg_obj := client.parseMessage(msg_frame)
						toGameManager <- msg_obj
						switch msg_obj.(type) {
						case *LoginMessage:
							loginmsg, _ := msg_obj.(*LoginMessage)
							if !loginmsg.LoggingIn {
								disconnect_player <- *client
								fmt.Printf("Disconnected Player: %v\n", client.client_address)
								break
							} else {
								m := loginmsg.CreateLoginMessageBytes(client.Seq)
								m.destination = client
								outgoing_msg <- *m
								client.Seq += 1
							}
						}
					}
					client.buffer = client.buffer[msg_frame.frame_length+msg_frame.content_length:]
				}
			}
		case msg_out, ok := <-client.outgoing_messages:
			if ok {
				switch cast_msg := msg_out.(type) {
				case PhysicsUpdateMessage:
					ship_msg := CreateShipUpdateMessage(cast_msg.Ships, client.Seq)
					outgoing_msg <- ship_msg
					client.Seq += 1
				}
			} else {
				break
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
		user_pass := strings.Split(string(content), ":")
		fmt.Printf("UserVa: '%v\n", user_pass)

		//fmt.Printf("User: '%v''  Pass: '%v'\n", user_pass[0], user_pass[1])
		// TODO: Check password? Lookup user? Maybe this should go to the game manager
		if user_pass[1] == "a" {
			client.User = &User{Id: 0}
			msg := &LoginMessage{GameMessageValues: *gmv, LoggingIn: true}
			return msg
		} else {
			fmt.Printf("Not handling incorrect password yet.")
			msg := &LoginMessage{GameMessageValues: *gmv, LoggingIn: false}
			return msg
		}
	case SETTHRUST:
		//5 USER CLEN [T1 PERC, T2 PERC]
		num_percents := len(content) / 2
		thrust_percents := make([]uint8, num_percents)
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

func (m *NetMessage) CreateMessageBytes(content []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Grow(5 + len(content))
	buf.WriteByte(byte(m.frame.message_type))
	binary.Write(buf, binary.LittleEndian, m.frame.sequence)
	binary.Write(buf, binary.LittleEndian, m.frame.content_length)
	binary.Write(buf, binary.LittleEndian, content)
	return buf.Bytes()
}

func (lm *LoginMessage) CreateLoginMessageBytes(seq uint16) *NetMessage {
	mt := LOGINSUCCESS
	if !lm.LoggingIn {
		mt = LOGINFAIL
	}
	m := &NetMessage{}
	m.frame = &MessageFrame{message_type: mt, content_length: 0, sequence: seq}
	m.raw_bytes = m.CreateMessageBytes([]byte{})
	return m
}

func CreateShipUpdateMessage(ships []*Ship, seq uint16) (m NetMessage) {
	content_length := 20 * len(ships)
	m.frame = &MessageFrame{message_type: 4, content_length: uint16(content_length)}
	buf := new(bytes.Buffer)
	buf.Grow(9 + content_length)
	buf.WriteByte(4)
	binary.Write(buf, binary.LittleEndian, seq)                    // Write seq
	binary.Write(buf, binary.LittleEndian, uint16(content_length)) // Write 2 byte content len
	for _, ship := range ships {
		buf.Write(ship.UpdateBytes())
	}
	m.raw_bytes = buf.Bytes()
	return
}
