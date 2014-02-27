package main

import (
	"fmt"
	"math"
	"time"
)

const (
	UPDATES_PER_SECOND = 50.0
	UPDATE_SLEEP       = 1000 / UPDATES_PER_SECOND
	FULL_CIRCLE        = math.Pi * 2
)

type SolarSimulator struct {
	Entities      map[uint32]Entity // Anything that can collide in space
	Characters    map[uint32]Entity // Things that can go inside of ships
	last_update   time.Time
	output_update chan EntityUpdate
}

func (ss *SolarSimulator) RunSimulation(input_update chan EntityUpdate) {
	// Wait
	for {
		timeout := ss.last_update.Add(time.Millisecond * UPDATE_SLEEP).Sub(time.Now())
		wait_for_timeout := true
		for wait_for_timeout {
			select {
			case update_msg := <-input_update:
				if update_msg.UpdateType == 1 {
					if ship, ok := update_msg.EntityObj.(Ship); ok {
						ss.Entities[ship.Id] = Entity(ship)
					}
				} else if update_msg.UpdateType == 3 {
					if update_ent, ok := update_msg.EntityObj.(Ship); ok {
						if ship, ok := ss.Entities[update_ent.Id].(Ship); ok {
							ship.Force = update_ent.Force
							ship.Torque = update_ent.Torque
						}
					}
				}
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			}
		}
		ss.Tick()
	}
}

func (ss *SolarSimulator) Tick() {
	eu := &EntityUpdate{UpdateType: byte(4)}
	for _, entity := range ss.Entities {
		if ship, ok := entity.(Ship); ok {
			changed := false

			ship.Velocity.Add(MultVect2(&ship.Force, ship.InvMass/UPDATES_PER_SECOND))
			ship.AngularVelocity += (ship.Torque * ship.InvInertia) / UPDATES_PER_SECOND

			if ship.Velocity[0] != 0.0 {
				ship.Position[0] += ship.Velocity[0] / UPDATES_PER_SECOND
				changed = true
			}
			if ship.Velocity[1] != 0.0 {
				ship.Position[1] += ship.Velocity[1] / UPDATES_PER_SECOND
				changed = true
			}
			if ship.AngularVelocity != 0.0 {
				ship.Angle += ship.AngularVelocity / UPDATES_PER_SECOND
				for ship.Angle > FULL_CIRCLE {
					ship.Angle -= FULL_CIRCLE
				}
				for ship.Angle < -FULL_CIRCLE {
					ship.Angle += FULL_CIRCLE
				}
				changed = true
			}
			if changed {
				eu.EntityObj = Entity(ship)
				ss.output_update <- *eu
			}
		} else {
			fmt.Println("That was not a ship")
		}
	}
	// Check for collisions?
}
