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
		for {
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
		for _, entity := range ss.Entities {
			if ship, ok := entity.(Ship); ok {
				ship.Position[0] += ship.Velocity[0] / UPDATES_PER_SECOND
			}
		}
	}
}
