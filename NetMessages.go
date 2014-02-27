package main

import (
	"bytes"
	"encoding/binary"
)

type NetMessageType byte

const (
	ECHO         NetMessageType = 0
	LOGINREQUEST NetMessageType = 1
	LOGINSUCCESS NetMessageType = 2
	LOGINFAIL    NetMessageType = 3
	PHYSICS      NetMessageType = 4
	SETTHRUST    NetMessageType = 5
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
	message_type   NetMessageType // byte 0
	sequence       uint16         // byte 1-2
	content_length uint16         // byte 3-4
	from_user      uint32         // Determined by net addr the request came on.
	frame_length   uint16         // This is only here in case of dynamic sized frames.
}

func ParseFrame(raw_bytes []byte) *MessageFrame {
	if len(raw_bytes) >= 5 {
		mf := new(MessageFrame)
		mf.message_type = NetMessageType(raw_bytes[0])
		var v uint16
		binary.Read(bytes.NewBuffer(raw_bytes[1:2]), binary.LittleEndian, &v)
		mf.sequence = v
		var cl uint16
		binary.Read(bytes.NewBuffer(raw_bytes[3:4]), binary.LittleEndian, &cl)
		mf.content_length = cl
		mf.frame_length = 5
		return mf
	}

	return nil
}
