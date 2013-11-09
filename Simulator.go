package main

import "time"

const (
	UPDATES_PER_SECOND = 50.0
	UPDATE_SLEEP       = 1000 / UPDATES_PER_SECOND
)

type SolarSimulator struct {
	Entities      map[int32]Entity // Anything that can collide in space
	Characters    map[int32]Entity // Things that can go inside of ships
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
							ship.Velocity = update_ent.Velocity
							ship.Rotation = update_ent.Rotation
						}
					}
				}
			case <-time.After(timeout):
				wait_for_timeout = false
				break
			}
		}

		// Tick
		eu := &EntityUpdate{UpdateType: byte(4)}
		for _, entity := range ss.Entities {
			if ship, ok := entity.(Ship); ok {
				changed := false
				if ship.Velocity[0] != 0.0 {
					ship.Position[0] += ship.Velocity[0] / UPDATES_PER_SECOND
					changed = true
				}
				if ship.Velocity[1] != 0.0 {
					ship.Position[1] += ship.Velocity[1] / UPDATES_PER_SECOND
					changed = true
				}
				if ship.RotationVelocity != 0.0 {
					ship.Rotation += ship.RotationVelocity / UPDATES_PER_SECOND
					if ship.Rotation > 1.0 {
						ship.Rotation -= 1
					}
					changed = true
				}
				if changed {
					eu.EntityObj = Entity(ship)
					ss.output_update <- *eu
				}
			}
		}
		// Check for collisions?
	}
}
