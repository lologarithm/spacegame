package main

import (
	"fmt"
	"math"
	"time"
)

const (
	UpdatesPerSecond = 50.0
	UpdateSleep      = 1000.0 / UpdatesPerSecond
	FullCircle       = math.Pi * 2
	AddShip          = byte(1)
	UpdateForces     = byte(3)
	UpdatePosition   = byte(4)
	UpdateCollision  = byte(5)
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
		timeout := ss.last_update.Add(time.Millisecond * UpdateSleep).Sub(time.Now())
		wait_for_timeout := true
		for wait_for_timeout {
			select {
			case update_msg := <-input_update:
				if update_msg.UpdateType == AddShip {
					if ship, ok := update_msg.EntityObj.(Ship); ok {
						ss.Entities[ship.Id] = Entity(&ship)
					}
				} else if update_msg.UpdateType == UpdateForces {
					if update_ent, ok := update_msg.EntityObj.(Ship); ok {
						if ship, ok := ss.Entities[update_ent.Id].(*Ship); ok {
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
		ss.last_update = time.Now()
		ss.Tick()
	}
}

func (ss *SolarSimulator) Tick() {
	eu := EntityUpdate{UpdateType: UpdatePosition}
	for _, entity := range ss.Entities {
		if ship, ok := entity.(*Ship); ok {
			changed := false

			ship.Velocity = ship.Velocity.Add(MultVect2(ship.Force, ship.InvMass/UpdatesPerSecond))
			ship.AngularVelocity += (ship.Torque * ship.InvInertia) / UpdatesPerSecond

			if ship.Velocity.X != 0.0 {
				ship.Position.X += ship.Velocity.X / UpdatesPerSecond
				changed = true
			}
			if ship.Velocity.Y != 0.0 {
				ship.Position.Y += ship.Velocity.Y / UpdatesPerSecond
				changed = true
			}
			if ship.AngularVelocity != 0.0 {
				ship.Angle += ship.AngularVelocity / UpdatesPerSecond
				for ship.Angle > FullCircle {
					ship.Angle -= FullCircle
				}
				for ship.Angle < -FullCircle {
					ship.Angle += FullCircle
				}
				changed = true
			}
			if changed {
				eu.EntityObj = Entity(*ship)
				ss.output_update <- eu
			}
		} else {
			fmt.Println("That was not a ship")
		}
	}
	// Check for collisions?
}
