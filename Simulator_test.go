package main

import (
	"fmt"
	"math"
	"testing"
)

func TestTick(t *testing.T) {
	// 1. Create simple scene
	physics_updates := make(chan EntityUpdate, 200)
	ss := &SolarSimulator{Entities: map[int32]Entity{}, Characters: map[int32]Entity{}, output_update: physics_updates}
	ship1 := &Ship{EntityData: EntityData{Id: 1}}
	ship1.Velocity = Vect2{1, 1}
	ship1.Position = Vect2{0, 0}
	ship1.Force = Vect2{0, 0}
	ss.Entities[1] = Entity(*ship1)
	// 2. Make sure single tick correctly ticks.
	for i := float32(1); i < 50.0; i += 1 {
		ss.Tick()
		if ship, ok := ss.Entities[1].(Ship); ok {
			if !FloatCompare(ship.Position.X(), i/50.0) {
				fmt.Printf("Incorrect X position after physics update. Expected: %f Actual: %f\n", i/50.0, ship.Position.X())
				t.FailNow()
			}
			fmt.Printf("(%d)Position: (%f, %f)\n", i, ship.Position.X(), ship.Position.Y())
		} else {
			fmt.Println("Error casting: ", ok)
		}
	}
}

func FloatCompare(a float32, b float32) bool {
	if math.Abs(float64(a)-float64(b)) < 0.00001 {
		return true
	}
	return false
}
