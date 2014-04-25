package main

import (
	"bytes"
	"encoding/binary"
)

// Users are 'accounts' that can login.
type User struct {
	Id              uint32
	Name            string
	Characters      []Character
	ActiveCharacter *Character
}

// Character is the in-game representation of a User.
type Character struct {
	EntityData  // anonomous field gives character all entity fields
	Health      int32
	CurrentShip *Ship
}

// Unique data for all in game Entities.
type EntityData struct {
	Id uint32 // Uniquely id this entity in space
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

// TODO: Move serial/deserial methdos to their own file. No need for them here.
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

func (ship *Ship) FromBytes(serial_ship []byte) {
	vals := []float32{0, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.Read(bytes.NewBuffer(serial_ship), binary.LittleEndian, &vals)
	ship.Position[0] = vals[0]
	ship.Position[1] = vals[1]
	ship.Velocity[0] = vals[2]
	ship.Velocity[1] = vals[3]
	ship.Force[0] = vals[4]
	ship.Force[1] = vals[5]
	ship.Angle = vals[6]
	ship.AngularVelocity = vals[7]
	ship.Torque = vals[8]
}

func (ship *Ship) CreateTestShip(id uint32, hull string) *Ship {
	ship.Id = id
	thrusters := []Thruster{
		Thruster{Max: 100.0, AngularPercent: 0.0, LinearPercent: 1.0, LinearVector: Vect2{0, 1}},
		Thruster{Max: 50.0, AngularPercent: 1.0, LinearPercent: 0.0, LinearVector: Vect2{0, 1}},
		Thruster{Max: 50.0, AngularPercent: -1.0, LinearPercent: 0.0, LinearVector: Vect2{0, 1}}}
	ship.Hull = &Hull{Name: hull, Thrusters: thrusters}
	ship.Mass = 1000.0
	ship.Position = Vect2{0, 0}
	ship.Velocity = Vect2{0, 0}
	ship.Force = Vect2{0, 0}
	return ship
}

type Entity interface {
	// Entity functions to here.
}
