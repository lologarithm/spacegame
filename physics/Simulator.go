package physics

import (
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
	Entities      map[uint32]*RigidBody // Anything that can collide in space
	Characters    map[uint32]*RigidBody // Things that can go inside of ships
	lastUpdate    time.Time
	OutSimulator  chan PhysicsEntityUpdate
	IntoSimulator chan PhysicsEntityUpdate
}

func (ss *SolarSimulator) RunSimulation() {
	ss.lastUpdate = time.Now()
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
			case update_msg := <-ss.IntoSimulator:
				if update_msg.UpdateType == AddShip {
					ss.Entities[update_msg.Body.Id] = &update_msg.Body
				} else if update_msg.UpdateType == UpdateForces {
					if body, ok := ss.Entities[update_msg.Body.Id]; ok {
						body.Force = update_msg.Body.Force
						body.Torque = update_msg.Body.Torque
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
	eu := PhysicsEntityUpdate{UpdateType: UpdatePosition}
	changed := false
	for _, rigid := range ss.Entities {
		changed = false

		rigid.Velocity = rigid.Velocity.Add(MultVect2(rigid.Force, rigid.InvMass/SimUpdatesPerSecond))
		rigid.AngularVelocity += (rigid.Torque * rigid.InvInertia) / SimUpdatesPerSecond

		if rigid.Velocity.X != 0.0 {
			rigid.Position.X += rigid.Velocity.X / SimUpdatesPerSecond
			changed = true
		}
		if rigid.Velocity.Y != 0.0 {
			rigid.Position.Y += rigid.Velocity.Y / SimUpdatesPerSecond
			changed = true
		}
		if rigid.AngularVelocity != 0.0 {
			rigid.Angle += rigid.AngularVelocity / SimUpdatesPerSecond
			for rigid.Angle > FullCircle {
				rigid.Angle -= FullCircle
			}
			for rigid.Angle < -FullCircle {
				rigid.Angle += FullCircle
			}
			changed = true
		}
		if changed && sendUpdate {
			eu.Body = *rigid
			ss.OutSimulator <- eu
		}
	}
	// Check for collisions?
}
