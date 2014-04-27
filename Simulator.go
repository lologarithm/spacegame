package main

import (
	"fmt"
	"math"
	"time"
)

const (
	SimUpdatesPerSecond = 50.0
	SimUpdateSleep      = 1000.0 / SimUpdatesPerSecond
	FullCircle          = math.Pi * 2
	AddShip             = byte(1)
	UpdateForces        = byte(3)
	UpdatePosition      = byte(4)
	UpdateCollision     = byte(5)
)

type SolarSimulator struct {
	Entities      map[uint32]Entity // Anything that can collide in space
	Characters    map[uint32]Entity // Things that can go inside of ships
	lastUpdate    time.Time
	outSimulator  chan EntityUpdate
	intoSimulator chan EntityUpdate
}

func (ss *SolarSimulator) RunSimulation() {
	sendUpdate := 0
	// Wait
	for {
		timeout := ss.lastUpdate.Add(time.Millisecond * SimUpdateSleep).Sub(time.Now())
		wait_for_timeout := true
		for wait_for_timeout {
			select {
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			case update_msg := <-ss.intoSimulator:
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
			}
		}
		ss.lastUpdate = time.Now()
		ss.Tick(sendUpdate%5 == 0)
		// Only set physics updates every X ticks.
		sendUpdate++
	}
}

func (ss *SolarSimulator) Tick(sendUpdate bool) {
	eu := EntityUpdate{UpdateType: UpdatePosition}
	changed := false
	for _, entity := range ss.Entities {
		if ship, ok := entity.(*Ship); ok {
			changed = false

			ship.Velocity = ship.Velocity.Add(MultVect2(ship.Force, ship.InvMass/SimUpdatesPerSecond))
			ship.AngularVelocity += (ship.Torque * ship.InvInertia) / SimUpdatesPerSecond

			if ship.Velocity.X != 0.0 {
				ship.Position.X += ship.Velocity.X / SimUpdatesPerSecond
				changed = true
			}
			if ship.Velocity.Y != 0.0 {
				ship.Position.Y += ship.Velocity.Y / SimUpdatesPerSecond
				changed = true
			}
			if ship.AngularVelocity != 0.0 {
				ship.Angle += ship.AngularVelocity / SimUpdatesPerSecond
				for ship.Angle > FullCircle {
					ship.Angle -= FullCircle
				}
				for ship.Angle < -FullCircle {
					ship.Angle += FullCircle
				}
				changed = true
			}
			if changed && sendUpdate {
				eu.EntityObj = Entity(*ship)
				ss.outSimulator <- eu
			}
		} else {
			fmt.Println("Non-ship entities not supported in Tick yet.")
		}
	}
	// Check for collisions?
}
