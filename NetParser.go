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
	buffer          []byte
	clientAddress   *net.UDPAddr
	fromClient      chan []byte      // Bytes from client to server
	fromGameManager chan GameMessage // GameMessages from GameManger to client
	User            *User            // User attached to this network client
	Seq             uint16
	quit            bool
}

// Accepts raw bytes from a socket and turns them into NetMessage objects and then
// later into GameMessages. These are passed into the GameManager. This function also
// accepts outgoing messages from the GameManager to the client.
func (client *Client) ProcessBytes(toGameManager chan GameMessage, toClient chan NetMessage, disconnect_player chan Client) {
	client.quit = false
	for !client.quit {
		select {
		case new_bytes, ok := <-client.fromClient:
			if !ok {
				break
			} else {
				client.buffer = append(client.buffer, new_bytes...)
				msgFrame, ok := ParseFrame(client.buffer)
				// Only try to parse if we have collected enough bytes.
				if ok && int(msgFrame.frame_length+msgFrame.content_length) <= len(client.buffer) {
					if msgFrame.message_type == Echo {
						netMsg := &NetMessage{
							frame:       msgFrame,
							rawBytes:    client.buffer[0 : msgFrame.frame_length+msgFrame.content_length],
							destination: client}
						toClient <- *netMsg
					} else {
						gameMsg := client.parseNetMessage(msgFrame)
						toGameManager <- gameMsg

						switch gameMsg.(type) {
						case *LoginMessage:
							loginMsg, _ := gameMsg.(*LoginMessage)
							if !loginMsg.LoggingIn {
								disconnect_player <- *client
								fmt.Printf("Disconnected Player: %v\n", client.clientAddress)
								break
							} else {
								m := loginMsg.CreateLoginMessageBytes(client.Seq)
								m.destination = client
								toClient <- *m
								client.Seq += 1
							}
						}
					}
					// Remove the used bytes from the buffer.
					client.buffer = client.buffer[msgFrame.frame_length+msgFrame.content_length:]
				}
			}
		case outgoingMsg, ok := <-client.fromGameManager:
			if ok {
				fmt.Printf("Message from game manager: %T", outgoingMsg)
				switch cast_msg := outgoingMsg.(type) {
				case PhysicsUpdateMessage:
					ship_msg := CreateShipUpdateMessage(cast_msg.Ships, client)
					toClient <- ship_msg
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
func (client *Client) parseNetMessage(msgFrame MessageFrame) GameMessage {
	content := client.buffer[msgFrame.frame_length : msgFrame.frame_length+msgFrame.content_length]
	gmv := &GameMessageValues{FromUser: msgFrame.from_user, Client: client}
	switch msgFrame.message_type {
	case LoginRequest:
		user_pass := strings.Split(string(content), ":")
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
	case SetThrust:
		num_percents := len(content)
		thrust_percents := make([]uint8, num_percents)
		for i := 0; i < num_percents; i++ {
			thrust_percents[i] = uint8(content[i])
		}
		msg := &SetThrustMessage{GameMessageValues: *gmv, ThrustPercent: thrust_percents}
		return msg
	case Disconnect:
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
	mt := LoginSuccess
	if !lm.LoggingIn {
		mt = LoginFail
	}
	m := &NetMessage{}
	m.frame = MessageFrame{message_type: mt, content_length: 0, sequence: seq}
	m.rawBytes = m.CreateMessageBytes([]byte{})
	return m
}

func CreateShipUpdateMessage(ships []*Ship, cl *Client) (m NetMessage) {
	content_length := uint16(36 * len(ships)) // TODO: Fix the 36 here to be a value from ship telling you how long serialized size is.
	m.frame = MessageFrame{message_type: Physics, content_length: content_length}
	buf := new(bytes.Buffer)
	buf.Grow(int(content_length))
	for _, ship := range ships {
		buf.Write(ship.UpdateBytes())
	}
	m.rawBytes = m.CreateMessageBytes(buf.Bytes())
	m.destination = cl
	return
}
