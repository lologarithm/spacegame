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
}

// Update message linked to an Entity.
type EntityUpdate struct {
	UpdateType byte   // 2 == login, 3 == logoff, 4 == physics update
	EntityObj  Entity // Passed by value through channels
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
	Hull *Hull // Ship hull information
} // TODO: Create "ShipMass" function to override RigidBody Mass.

type Hull struct {
	Name      string     // Ship hull name
	Thrusters []Thruster // List of thrusters and current thrust state
	COM       Vect2      // Center of Mass
}

type Thruster struct {
	Max            float32 // Max thrust (N)
	Current        float32 // Current thrust (N)
	AngularPercent float32 // Percent of thrust that applies to torque. Can be negative
	LinearPercent  float32 // Percent of thrust that applies to force
	LinearVector   Vect2   // Unit Vector to apply thrust in.
}

func (ship *Ship) UpdateBytes() []byte {
	buf := new(bytes.Buffer)
	buf.Grow(36)
	binary.Write(buf, binary.LittleEndian, ship.Position[0])
	binary.Write(buf, binary.LittleEndian, ship.Position[1])
	binary.Write(buf, binary.LittleEndian, ship.Velocity[0])
	binary.Write(buf, binary.LittleEndian, ship.Velocity[1])
	binary.Write(buf, binary.LittleEndian, ship.Force[0])
	binary.Write(buf, binary.LittleEndian, ship.Force[1])
	binary.Write(buf, binary.LittleEndian, ship.Angle)
	binary.Write(buf, binary.LittleEndian, ship.AngularVelocity)
	binary.Write(buf, binary.LittleEndian, ship.Torque)
	return buf.Bytes()
}

func (ship *Ship) CreateTestShip() {
	ship.Id = 1
	thrusters := []Thruster{
		Thruster{Max: 100.0, AngularPercent: 0.0, LinearPercent: 1.0, LinearVector: Vect2{0, 1}},
		Thruster{Max: 50.0, AngularPercent: 1.0, LinearPercent: 0.0, LinearVector: Vect2{0, 1}},
		Thruster{Max: 50.0, AngularPercent: -1.0, LinearPercent: 0.0, LinearVector: Vect2{0, 1}}}
	ship.Hull = &Hull{Name: "TestShip", Thrusters: thrusters}
	ship.Mass = 1000.0
	ship.Position = Vect2{0, 0}
}

type Entity interface {
	// Entity functions to here.
}
