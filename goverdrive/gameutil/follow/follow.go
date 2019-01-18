// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package follow

import (
	"fmt"
	"math"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
)

// Follower issues commands to a "follower" vehicle, to maintain a positional
// relationship relative to a "leader" vehicle. Two or more Followers can be
// used to create a multi-vehicle formation.
type Follower struct {
	vLeader         int
	vFollow         int
	targetDeltaDofs phys.Meters
	targetDeltaCofs phys.Meters
	dacl            phys.MetersPerSec2
	cspd            phys.MetersPerSec
	trackLen        phys.Meters
	adjustPeriod    phys.SimTime // how often to adjust speed/position of follow vehicle
	nextUpdateTime  phys.SimTime
}

const (
	maxDofsDistNear     = 0.010
	maxCofsDistNear     = 0.002
	majorCatchupFactor  = 1.25
	majorFallbackFactor = 0.75
	minorCatchupFactor  = 1.05
	minorFallbackFactor = 0.95
)

// New returns a pointer to a new Follow object.
func New(vLeader, vFollow int,
	targetDeltaDofs, targetDeltaCofs phys.Meters,
	dacl phys.MetersPerSec2,
	cspd phys.MetersPerSec,
	trackLen phys.Meters,
	now, adjustPeriod phys.SimTime) *Follower {

	c := Follower{
		vLeader:        vLeader,
		vFollow:        vFollow,
		dacl:           dacl,
		cspd:           cspd,
		trackLen:       trackLen,
		adjustPeriod:   adjustPeriod,
		nextUpdateTime: now,
	}
	c.SetTargetDeltaDofs(targetDeltaDofs)
	c.SetTargetDeltaCofs(targetDeltaCofs)

	return &c
}

func (c *Follower) TargetDeltaDofs() phys.Meters {
	return c.targetDeltaDofs
}

func (c *Follower) TargetDeltaCofs() phys.Meters {
	return c.targetDeltaCofs
}

func (c *Follower) SetTargetDeltaDofs(target phys.Meters) {
	if math.Abs(float64(target)) >= float64(c.trackLen/2) {
		panic(fmt.Sprintf("target=%v is too large; must be less than (trackLen/2)=%v", target, c.trackLen/2))
	}
	c.targetDeltaDofs = target
}

func (c *Follower) SetTargetDeltaCofs(target phys.Meters) {
	if math.Abs(float64(target)) >= float64(c.trackLen/2) {
		panic(fmt.Sprintf("target=%v is too large; must be less than (trackLen/2)=%v", target, c.trackLen/2))
	}
	c.targetDeltaCofs = target
}

// Update issues new vehicle commands, if needed, to maintain the desired
// leader-follower positional relationship. true is returned if the follower is
// sufficiently near its relative target position.
func (c *Follower) Update(rsys *robo.System) bool {
	// l=leader, f=follower (for brevity)
	lVeh := &rsys.Vehicles[c.vLeader]
	fVeh := &rsys.Vehicles[c.vFollow]

	deltaDist := rsys.Track.DriveDeltaDist(lVeh.CurTrackPose(), fVeh.CurTrackPose().Dofs)
	deltaDofsErrAmt := deltaDist - c.targetDeltaDofs
	deltaCofsErrAmt := rsys.Track.DriveDeltaCofs(lVeh.CurTrackPose(), fVeh.CmdTrackCofs()) - c.targetDeltaCofs

	if rsys.Now() >= c.nextUpdateTime {
		c.nextUpdateTime = rsys.Now() + c.adjustPeriod

		// Driving direction
		if fVeh.IsFacingTrackwise() != lVeh.IsFacingTrackwise() {
			fVeh.CmdUturn(robo.DefUturnRadius)
			return false
		}

		// Dofs Speed
		lRp := rsys.Track.Rp(rsys.Track.RpiAt(lVeh.CurTrackPose().Dofs))
		fDspd := lVeh.CurDriveDspd()
		if !lRp.IsStraight() {
			fDspd *= phys.MetersPerSec(lRp.CurveRadius(fVeh.CurTrackPose().Cofs) / lRp.CurveRadius(lVeh.CurTrackPose().Cofs))
		}
		if deltaDofsErrAmt > (+maxDofsDistNear) {
			// ahead of desired position => fall back
			fDspd *= majorFallbackFactor
		} else if deltaDofsErrAmt < (-maxDofsDistNear) {
			// behind desired position => catch up
			fDspd *= majorCatchupFactor
		} else if deltaDofsErrAmt > 0 {
			fDspd *= minorFallbackFactor
		} else if deltaDofsErrAmt < 0 {
			fDspd *= minorCatchupFactor
		}
		fVeh.SetCmdDriveDspd(fDspd, c.dacl)
		//fmt.Printf("deltaDist=%v  deltaDofsErrAmt=%v  fDspd=%v\n", deltaDist, deltaDofsErrAmt, fDspd)

		// Cofs
		if math.Abs(float64(deltaCofsErrAmt)) > maxCofsDistNear {
			fVeh.SetCmdDriveCofs(lVeh.CurDriveCofs()+c.targetDeltaCofs, c.cspd)
		}
	}

	// return value = "follower is near target position"
	return (math.Abs(float64(deltaDofsErrAmt)) <= maxDofsDistNear) &&
		/**/ (math.Abs(float64(deltaCofsErrAmt)) <= maxCofsDistNear)
}
