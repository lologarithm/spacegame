package main

import "time"

type SolarSimulator struct {
	entities    map[int32]*Entity // Anything that can collide in space
	characters  map[int32]*Entity // Things that can go inside of ships
	next_update time.Time
}

func (ss *SolarSimulator) runSimulation(update chan EntityUpdate, login chan Entity, quit chan int) {
	//
	for {
		//time.After(time.Millisecond)
	}
}
