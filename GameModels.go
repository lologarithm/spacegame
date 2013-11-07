package main

import (
	"bytes"
	"encoding/binary"
)

type User struct {
	Id              int32
	Name            string
	Characters      []Character
	ActiveCharacter *Character
}

type Character struct {
	EntityData // anonomous field gives character all entity fields
	Health     int32
}

type EntityData struct {
	Id               int32      // Uniquely id this entity in space
	Position         [2]float32 // coords x,y of entity
	Velocity         [2]float32 // speed in vector format
	Rotation         float32    // Current heading
	RotationVelocity float32    // speed of rotation around the Z axis (negative is counter clockwise)
	Mass             float32    // mass effects physics!
}

type EntityUpdate struct {
	UpdateType byte // 2 == login, 3 == logoff, 4 == physics update
	EntityObj  Entity
}

type CelestialBody struct {
	EntityData
	BodyType string // 'star' 'planet' 'asteroid'
}

type Ship struct {
	EntityData
	Hull string // something something darkside
}

func (ship *Ship) UpdateBytes() []byte {
	buf := new(bytes.Buffer)
	buf.Grow(24)
	binary.Write(buf, binary.LittleEndian, ship.Position[0])
	binary.Write(buf, binary.LittleEndian, ship.Position[1])
	binary.Write(buf, binary.LittleEndian, ship.Velocity[0])
	binary.Write(buf, binary.LittleEndian, ship.Velocity[1])
	binary.Write(buf, binary.LittleEndian, ship.Rotation)
	binary.Write(buf, binary.LittleEndian, ship.RotationVelocity)
	return buf.Bytes()
}

type Entity interface {
	// Entity functions to here.
}
