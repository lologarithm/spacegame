package main

import (
	"fmt"
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
	for i := float32(1); i < 50.0; i += 1 {
		ss.Tick()
		if ship, ok := ss.Entities[1].(Ship); ok {
			if float32(ship.Position.X()) != float32(i/50.0) {
				fmt.Printf("Incorrect X position after physics update. Expected: %f,Actual: %f\n", i/50.0, ship.Position.X())
				t.FailNow()
			}
			//fmt.Printf("(%d)Position: (%f, %f)\n", i, ship.Position.X(), ship.Position.Y())
		} else {
			fmt.Println("Error casting: ", ok)
		}
	}
	// 2. Make sure single tick correctly ticks.
}
