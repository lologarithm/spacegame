package main

import "time"

type SolarSimulator struct {
	entities    map[int32]*Entity
	next_update time.Time
}

func (ss *SolarSimulator) runSimulation(update chan EntityUpdate, login chan Entity, quit chan int) {
	for {

	}
}
