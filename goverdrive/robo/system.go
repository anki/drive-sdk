// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package robo is all of the native robotics, such as tracks, vehicles, and
// simulation modeling.
package robo

import (
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo/track"
)

const (
	// TODO(gwenz): It's unclear what range of values works for simDeltaT, and
	// whether there is any need to change it.
	// Note that this value interacts with roboTicksPerGameTick in gameloop.go.
	// They may need to be changed together.
	simDeltaT = 1e7 * phys.SimNanosecond
)

// System is the complete robotics system being simulated.
type System struct {
	dt       phys.SimTime // length of time for one sim tick
	now      phys.SimTime
	Track    track.Track
	Vehicles []Vehicle
	Collider VehicleCollider
	sim      Simulator
}

func NewSystem(trk *track.Track, vehs *[]Vehicle, sim Simulator, collider VehicleCollider) *System {
	return &System{
		dt:       simDeltaT,
		now:      0,
		Track:    *trk,
		Vehicles: *vehs,
		Collider: collider,
		sim:      sim,
	}
}

func (s *System) SimDeltaT() phys.SimTime {
	return s.dt
}

func (s *System) Now() phys.SimTime {
	return s.now
}

// Tick runs the robotics system (vehicle simulation, collision detection, etc)
// for one small "tick" of time, ie a duration of s.dt. Only the game loop
// should call Tick.
func (s *System) Tick() {
	s.now += s.dt
	s.sim.Tick(s.dt, &s.Track, &s.Vehicles)
	for _, v := range s.Vehicles {
		v.Lights().Update(s.now)
	}
	s.Collider.update(s.now, &s.Track, &s.Vehicles)
	// TODO: Update/apply external forces?
}
