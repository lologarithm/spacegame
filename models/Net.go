package models

import (
	"encoding/binary"
	"fmt"
	"net"
)

type NetMessageType byte

const (
	Echo         NetMessageType = 0
	LoginRequest NetMessageType = 1
	LoginSuccess NetMessageType = 2
	LoginFail    NetMessageType = 3
	Physics      NetMessageType = 4
	SetThrust    NetMessageType = 5
	GameStatus   NetMessageType = 10
	Disconnect   NetMessageType = 255
)

type Client struct {
	buffer          []byte
	clientAddress   *net.UDPAddr
	fromNetwork     chan []byte      // Bytes from client to server
	fromGameManager chan GameMessage // GameMessages from GameManger to client
	toServerManager chan GameMessage // Messages to server manager to join a game
	toGameManager   chan GameMessage // Messages to the game the client is connected to.
	User            *User            // User attached to this network client
	Seq             uint16
	Quit            bool
}

type NetMessage struct {
	rawBytes    []byte
	frame       MessageFrame
	destination *Client
}

func (m *NetMessage) Content() []byte {
	return m.rawBytes[m.frame.frame_length : m.frame.frame_length+m.frame.content_length]
}

type MessageFrame struct {
	message_type   NetMessageType // byte 0
	sequence       uint16         // byte 1-2
	content_length uint16         // byte 3-4
	from_user      uint32         // Determined by net addr the request came on.
	frame_length   uint16         // This is only here in case of dynamic sized frames.
}

func (mf MessageFrame) String() string {
	return fmt.Sprintf("Type: %d, Seq: %d, CL: %d, FL: %d\n", mf.message_type, mf.sequence, mf.content_length, mf.frame_length)
}

func ParseFrame(rawBytes []byte) (mf MessageFrame, ok bool) {
	if len(rawBytes) < 5 {
		return
	}
	mf.message_type = NetMessageType(rawBytes[0])
	mf.sequence = binary.LittleEndian.Uint16(rawBytes[1:3])
	mf.content_length = binary.LittleEndian.Uint16(rawBytes[3:5])
	mf.frame_length = 5
	return mf, true
}
