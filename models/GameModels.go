package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	binary.Write(buf, binary.LittleEndian, ship.Position.X)
	binary.Write(buf, binary.LittleEndian, ship.Position.Y)
	binary.Write(buf, binary.LittleEndian, ship.Velocity.X)
	binary.Write(buf, binary.LittleEndian, ship.Velocity.Y)
	binary.Write(buf, binary.LittleEndian, ship.Force.X)
	binary.Write(buf, binary.LittleEndian, ship.Force.Y)
	binary.Write(buf, binary.LittleEndian, ship.Angle)
	binary.Write(buf, binary.LittleEndian, ship.AngularVelocity)
	binary.Write(buf, binary.LittleEndian, ship.Torque)
	return buf.Bytes()
}

func (ship *Ship) FromBytes(serial_ship []byte) {
	vals := []float32{0, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.Read(bytes.NewBuffer(serial_ship), binary.LittleEndian, &vals)
	ship.Position.X = vals[0]
	ship.Position.Y = vals[1]
	ship.Velocity.X = vals[2]
	ship.Velocity.Y = vals[3]
	ship.Force.X = vals[4]
	ship.Force.Y = vals[5]
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
	ship.InvMass = 1.0 / ship.Mass
	ship.Inertia = 1000.0
	ship.InvInertia = 1.0 / 1000.0
	ship.Position = Vect2{0, 0}
	ship.Velocity = Vect2{0, 0}
	ship.Force = Vect2{0, 0}
	return ship
}

func (ship *Ship) String() string {
	return fmt.Sprintf("ID: %d, Pos: %v, Vel: %v, Force: %v", ship.Id, ship.Position, ship.Velocity, ship.Force)
}

type Entity interface {
	// Entity functions to here.
}
