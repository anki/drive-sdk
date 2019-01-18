// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package robo

import (
	_ "fmt"
	"math"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo/track"
)

// Simulator simulates all vehicle robotic movement and interaction, such as:
//  - Steady-state driving
//  - Speed changes
//  - Horziontal offset changes
//  - Collisions
type Simulator interface {
	// Tick updates the state of all vehicles on the track, including speed, pose,
	// etc.
	Tick(dt phys.SimTime, trk *track.Track, vehs *[]Vehicle)
}

//////////////////////////////////////////////////////////////////////

// IdealSimulator simulates ideal motion with no limits to motor ability, no
// center offset drift, etc. It is intended to be very simple.
type IdealSimulator struct {
	// TODO: Any fields needed?
}

func NewIdealSimulator() *IdealSimulator {
	return &IdealSimulator{}
}

func (sim *IdealSimulator) Tick(dt phys.SimTime, trk *track.Track, vehs *[]Vehicle) {
	for v, _ := range *vehs {
		var veh *Vehicle = &(*vehs)[v]
		rpi, _ := trk.RpiAndRpDofs(veh.CurTrackPose().Dofs)
		rp := trk.Rp(rpi)

		// To reduce clutter, use type float64 for all intermediate values
		fdt := float64(dt) * 1e-9
		desDspd := float64(veh.desDspd)
		cmdDspd := float64(veh.cmdDspd)

		// Calc new dofs speed (ie apply constant [de/a]cceleration)
		dspdDelta := fdt * float64(veh.cmdDacl)
		if math.Abs(desDspd-cmdDspd) <= dspdDelta {
			desDspd = cmdDspd
		} else if desDspd < cmdDspd {
			desDspd += dspdDelta
		} else { // desDspd > cmdDspd
			desDspd -= dspdDelta
		}
		curDspd := desDspd // ideal sim model means (cur==des) always

		// Calc new dofs
		// Formula = standard calculus for rigid body movement under constant acceleration
		deltaFwd := (curDspd * fdt) + ((float64(veh.cmdDacl) / 2) * fdt * fdt)
		deltaDofs := deltaFwd
		if rp.CurveRadius(0) != 0 {
			// remember that Dofs is measured along road center
			deltaDofs *= float64(rp.CurveRadius(0)) / float64(rp.CurveRadius(veh.CurTrackPose().Cofs))
		}

		// Calc new hofs
		if veh.cmdCofs < -trk.MaxCofs() {
			veh.cmdCofs = -trk.MaxCofs()
		} else if veh.cmdCofs > trk.MaxCofs() {
			veh.cmdCofs = trk.MaxCofs()
		}
		desCofs := float64(veh.desCofs)
		cmdCofs := float64(veh.cmdCofs)
		curCspd := math.Abs(float64(veh.cmdCspd))
		curHvel := curCspd
		maxDeltaCofs := fdt * curCspd // max possible (for this tick)
		absDeltaCofs := float64(0)    // actual
		if desCofs < cmdCofs {
			curHvel = curCspd
			if (desCofs + maxDeltaCofs) > cmdCofs {
				absDeltaCofs = cmdCofs - desCofs
				desCofs = cmdCofs
			} else {
				absDeltaCofs = maxDeltaCofs
				desCofs += maxDeltaCofs
			}
		} else if desCofs > cmdCofs {
			curHvel = -curCspd
			if (desCofs - maxDeltaCofs) < cmdCofs {
				absDeltaCofs = desCofs - cmdCofs
				desCofs = cmdCofs
			} else {
				absDeltaCofs = maxDeltaCofs
				desCofs -= maxDeltaCofs
			}
		} else {
			curHvel = 0
		}
		//fmt.Printf("  desCofs=%v, curCspd=%v, absDeltaCofs=%v\n", desCofs, curCspd, absDeltaCofs)

		// Update the vehicle's state
		veh.desDspd = phys.MetersPerSec(desDspd)
		veh.desCofs = phys.Meters(desCofs)
		if veh.IsFacingTrackwise() {
			veh.curVel.D = phys.MetersPerSec(curDspd)
			veh.curPose.Dofs += phys.Meters(deltaDofs)
		} else {
			veh.curVel.D = -phys.MetersPerSec(curDspd)
			veh.curPose.Dofs -= phys.Meters(deltaDofs)
		}
		veh.curPose.Dofs = trk.NormalizeDofs(veh.curPose.Dofs)
		veh.curPose.Cofs = phys.Meters(desCofs) // ideal sim model means (cur==des) always
		veh.curVel.C = phys.MetersPerSec(curHvel)

		// Update pose angle based on new V and H speeds
		if curDspd > 0 {
			//fmt.Printf("veh.curVel.C=%v, veh.curVel.D=%v\n", veh.curVel.C, veh.curVel.D)
			angle := math.Atan2(float64(veh.curVel.C), float64(veh.curVel.D))
			veh.curPose.DAngle = phys.Radians(angle)
		}
		//fmt.Printf("  veh.curVel.D=%v\n", veh.curVel.D)

		// Update Odometer
		pathLen := math.Sqrt((deltaFwd * deltaFwd) + (absDeltaCofs * absDeltaCofs))
		veh.odom += phys.Meters(pathLen)
	}
}
