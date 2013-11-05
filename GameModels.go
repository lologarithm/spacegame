package main

import (
	"bytes"
	"encoding/binary"
)

type User struct {
	id         int32
	characters []Character
}

type Character struct {
	EntityData // anonomous field gives character all entity fields
}

type EntityData struct {
	id       int32      // Uniquely id this entity in space
	speed    [2]float32 // speed in vector format
	rotation float32    // speed of rotation around the Z axis (negative is counter clockwise)
	position [2]float32 // coords x,y of entity
	mass     float32    // mass effects physics!
}

type EntityUpdate struct {
	update_type byte // 1 == login, 2 == logoff, 3 == physics update
	ent_obj     Entity
}

type CelestialBody struct {
	EntityData
	body_type string // 'star' 'planet' 'asteroid'
}

type Ship struct {
	EntityData
	hull string // something something darkside
}

func (ship *Ship) CreateUpdateMessage() (m Message) {
	m.frame = &MessageFrame{message_type: 3, frame_length: 9, content_length: 20}
	buf := new(bytes.Buffer)
	buf.Grow(49)
	buf.WriteByte(3)
	binary.Write(buf, binary.LittleEndian, int32(0))
	binary.Write(buf, binary.LittleEndian, int32(20))
	binary.Write(buf, binary.LittleEndian, ship.position[0])
	binary.Write(buf, binary.LittleEndian, ship.position[1])
	binary.Write(buf, binary.LittleEndian, ship.speed[0])
	binary.Write(buf, binary.LittleEndian, ship.speed[1])
	binary.Write(buf, binary.LittleEndian, ship.rotation)
	m.raw_bytes = buf.Bytes()
	return
}

type Entity interface {
	// Entity functions to here.
}
