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
	EntityData  // anonomous field gives character all entity fields
	Health      int32
	CurrentShip *Ship
}

type EntityData struct {
	Id int32 // Uniquely id this entity in space
} // TODO: Mass a function to calculate total mass?

type RigidBody struct {
	Position Vect2 // coords x,y of entity  (meters)
	Velocity Vect2 // speed in vector format (m/s)
	Force    Vect2 // Force

	Angle           float32 // Current heading (radians)
	AngularVelocity float32 // speed of rotation around the Z axis (radians/sec)
	Torque          float32 // Torque

	Mass       float32 // mass of the ship, (kg)
	InvMass    float32
	Inertia    float32
	InvInertia float32
}

type EntityUpdate struct {
	UpdateType byte // 2 == login, 3 == logoff, 4 == physics update
	EntityObj  Entity
}

type CelestialBody struct {
	EntityData
	BodyType string // 'star' 'planet' 'asteroid'
}

// Object to describe a ship.
// TODO: define all customizable bits, subsystems, power, etc
type Ship struct {
	EntityData
	RigidBody
	Hull          string    // String to identify ship type
	ThrusterPower []float32 // List of thrusters and % power
}

func (ship *Ship) UpdateBytes() []byte {
	buf := new(bytes.Buffer)
	buf.Grow(24)
	binary.Write(buf, binary.LittleEndian, ship.Position[0])
	binary.Write(buf, binary.LittleEndian, ship.Position[1])
	binary.Write(buf, binary.LittleEndian, ship.Velocity[0])
	binary.Write(buf, binary.LittleEndian, ship.Velocity[1])
	binary.Write(buf, binary.LittleEndian, ship.Angle)
	binary.Write(buf, binary.LittleEndian, ship.AngularVelocity)
	return buf.Bytes()
}

type Entity interface {
	// Entity functions to here.
}
