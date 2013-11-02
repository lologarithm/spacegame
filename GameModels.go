package main

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
	id       int32
	rotation float32
	speed    [2]float32
}

type CelestialBody struct {
	EntityData
	body_type string // 'star' 'planet' 'asteroid'
}

type Ship struct {
	EntityData
	hull string // something something darkside
}

type Entity interface {
	// Entity functions to here.
}
