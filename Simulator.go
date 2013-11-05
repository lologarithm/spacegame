package main

import "time"

type SolarSimulator struct {
	entities      map[int32]Entity // Anything that can collide in space
	characters    map[int32]Entity // Things that can go inside of ships
	next_update   time.Time
	output_update chan EntityUpdate
}

func (ss *SolarSimulator) RunSimulation(input_update chan EntityUpdate) {
	// Wait
	for {
		update_timer := time.After(ss.next_update.Sub(time.Now()))
		ss.next_update = time.Now().Add(time.Millisecond * 20)
		for {
			select {
			case update_msg := <-input_update:
				if update_msg.update_type == 1 {
					if ship, ok := update_msg.ent_obj.(Ship); ok {
						ss.entities[ship.id] = Entity(ship)
					}
				} else if update_msg.update_type == 3 {
					if update_ent, ok := update_msg.ent_obj.(Ship); ok {
						if ship, ok := ss.entities[update_ent.id].(Ship); ok {
							ship.speed = update_ent.speed
							ship.rotation = update_ent.rotation
						}
					}
				}
			case <-update_timer:
				break
			}
		}

		// Tick

	}
}
