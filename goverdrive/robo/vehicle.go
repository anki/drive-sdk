// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package robo

import (
	"fmt"
	"image/color"
	"math"

	cn "golang.org/x/image/colornames"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/robo/track"
)

const (
	// DefUturnRadius is the "typical" turn radius for a non-truck vehicle
	DefUturnRadius phys.Meters = 0.05
)

//////////////////////////////////////////////////////////////////////
// Vehicle Types
//////////////////////////////////////////////////////////////////////

// VehType is a two-letter abbreviation for a vehicle "type" (aka "model" in
// some contexts). Using two-letter names is simple, concise, and lines up
// nicely for table-driven code.
//
// Examples:
//   "gs" = Groundshock
//   "sk" = Skull
//   "th" = Thermo
type VehType string // two letters, lowercase (eg "gs", "sk")

// VehTypeInfo stores name, physical properties, etc for a vehicle.
type VehTypeInfo struct {
	FullName string // eg "Groundshock"
	Color    color.Color
	Width    phys.Meters
	Length   phys.Meters
	Mass     phys.Grams
}

// TODO: Better to put vehicle info into JSON file(s)?
var vehTypeInfoTable = map[VehType]VehTypeInfo{
	"gs": VehTypeInfo{FullName: "Groundshock" /**/, Color: cn.Royalblue /*******/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"sk": VehTypeInfo{FullName: "Skull" /********/, Color: cn.Darkslategray /***/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"nk": VehTypeInfo{FullName: "Nuke" /*********/, Color: cn.Limegreen /*******/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"th": VehTypeInfo{FullName: "Thermo" /*******/, Color: cn.Orangered /*******/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"gu": VehTypeInfo{FullName: "Guardian" /*****/, Color: cn.Skyblue /*********/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"bb": VehTypeInfo{FullName: "BigBang" /******/, Color: cn.Seagreen /********/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"fw": VehTypeInfo{FullName: "Freewheel" /****/, Color: cn.Lime /************/, Width: 0.044, Length: 0.24, Mass: 40.0},
	"xr": VehTypeInfo{FullName: "X52" /**********/, Color: cn.Red /*************/, Width: 0.044, Length: 0.24, Mass: 40.0},
	"xi": VehTypeInfo{FullName: "X52Ice" /*******/, Color: cn.White /***********/, Width: 0.044, Length: 0.24, Mass: 40.0},
	"dy": VehTypeInfo{FullName: "Dynamo" /*******/, Color: cn.Darkgray /********/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"mm": VehTypeInfo{FullName: "Mammoth" /******/, Color: cn.Lightsteelblue /**/, Width: 0.044, Length: 0.08, Mass: 40.0},
	"np": VehTypeInfo{FullName: "NukePhantom" /**/, Color: cn.Ghostwhite /******/, Width: 0.044, Length: 0.08, Mass: 40.0},
}

//////////////////////////////////////////////////////////////////////
// Vehicle
//////////////////////////////////////////////////////////////////////

// Vehicle models a vehicle's robotics state and qualities, for simulation. It
// does NOT capture any game-specific state or behavior, such as weapons
// game-imposed min/max speeds, lap counts, etc.
//
// Abbreviations:
//   cmd = commanded (eventual)
//   des = desired at this moment
//   cur = current at this moment
type Vehicle struct {
	trackLen phys.Meters
	vtype    VehType
	lights   light.VehLights

	odom    phys.Meters // odometer = total distance driven (runs continuously)
	curPose track.Pose
	curVel  track.Vel

	cmdDspd phys.MetersPerSec  // commanded target distance speed (eventual)
	cmdDacl phys.MetersPerSec2 // commanded distance accel
	desDspd phys.MetersPerSec  // desired distance speed at this moment

	cmdCofs phys.Meters       // commanded target center offset (eventual)
	cmdCspd phys.MetersPerSec // commanded center speed (for lane change)
	desCofs phys.Meters       // desired center offset at this moment

	// TODO: Include fields to model [temporary] external accel? (eg centrifugal; hills; collision)
	// TODO: Or, is this handled in a different part of the robotics system?
}

// NewVehicle creates a new vehicle of the desired type. The vehicle is idle at
// the origin.
func NewVehicle(vt VehType, lspec light.Spec, trackLen phys.Meters) *Vehicle {
	_, ok := vehTypeInfoTable[vt]
	if !ok {
		helpstr := ""
		for k, v := range vehTypeInfoTable {
			helpstr += fmt.Sprintf("  %s  %s\n", k, v.FullName)
		}
		panic(fmt.Sprintf("VehType=%v is invalid. Valid vehicle types:\n%s", vt, helpstr))
	}

	return &Vehicle{
		trackLen: trackLen,
		vtype:    vt,
		lights:   *light.NewVehLights(lspec),
		odom:     0,
		curPose:  track.Pose{Point: track.Point{Dofs: 0, Cofs: 0}, DAngle: 0},
		curVel:   track.Vel{D: 0, C: 0},
		cmdDspd:  0,
		cmdDacl:  0.1,
		desDspd:  0,
		cmdCofs:  0,
		cmdCspd:  0.1,
		desCofs:  0,
	}
}

// Type is the vehicle's type.
func (v *Vehicle) Type() VehType {
	return v.vtype
}

// Width is the physical width of the vehicle.
func (v *Vehicle) Width() phys.Meters {
	return vehTypeInfoTable[v.vtype].Width
}

// Length is the physical length of the vehicle.
func (v *Vehicle) Length() phys.Meters {
	return vehTypeInfoTable[v.vtype].Length
}

// Color is the vehicle's shell color
func (v *Vehicle) Color() color.Color {
	return vehTypeInfoTable[v.vtype].Color
}

// Lights returns a handle to the vehicle's lights.
func (v *Vehicle) Lights() *light.VehLights {
	return &v.lights
}

// Odom is the current odometer reading, ie total Meters driven since the
// vehicle was created.
func (v *Vehicle) Odom() phys.Meters {
	return v.odom
}

// CurTrackPose is the vehicle's current pose on the track.
func (v *Vehicle) CurTrackPose() track.Pose {
	return v.curPose
}

// CurTrackVel is the vehicle's current track velocity (V and H).
func (v *Vehicle) CurTrackVel() track.Vel {
	return v.curVel
}

// IsFacingTrackwise returns true if the vehicle is facing in the natural
// forward direction of the track. If the vehicle is not stopped, this is also
// the direcion the vehicle is driving.
// NOTE: The "trackwise" concept is akin to the "clockwise" concept.
func (v *Vehicle) IsFacingTrackwise() bool {
	absDAngle := math.Abs(float64(v.curPose.DAngle))
	if absDAngle <= (math.Pi / 2) {
		return true
	}
	if (absDAngle > (math.Pi / 2)) && (absDAngle <= math.Pi) {
		return false
	}

	panic(fmt.Sprintf("Vehicle.curPose.Dangle=%v is invalid!", v.curPose.DAngle))
	return false
}

// CmdDriveDspd is the commanded distance speed, in the vehicle's driving
// direction. It is always >= 0.
func (v *Vehicle) CmdDriveDspd() phys.MetersPerSec {
	return v.cmdDspd
}

// CurDriveDspd is the current distance speed, in the vehicle's driving
// direction. It is always >= 0.
func (v *Vehicle) CurDriveDspd() phys.MetersPerSec {
	return phys.MetersPerSec(math.Abs(float64(v.curVel.D)))
}

// CmdDriveCofs is the commanded center offset, in the vehicle's driving
// direction. That is, CmdDriveCofs>0 means left-of-center in the vehicle's
// driving direction.
func (v *Vehicle) CmdDriveCofs() phys.Meters {
	if v.IsFacingTrackwise() {
		return v.cmdCofs
	} else {
		return -v.cmdCofs
	}
}

// CurDriveDofs is the current distance offset (along road center) from the
// finish line, in the vehicle's driving direction.
func (v *Vehicle) CurDriveDofs() phys.Meters {
	if v.IsFacingTrackwise() {
		return v.curPose.Dofs
	} else {
		return v.trackLen - v.curPose.Dofs
	}
}

// CurDriveDofsRem is the distance offset to the finish line (ie road center
// distance remaining), in the vehicle's driving direction.
func (v *Vehicle) CurDriveDofsRem() phys.Meters {
	if v.IsFacingTrackwise() {
		return v.trackLen - v.curPose.Dofs
	} else {
		return v.curPose.Dofs
	}
}

// CurDriveCofs is the current center offset, in the vehicle's driving
// direction. That is, CurDriveCofs>0 means left-of-center in the vehicle's
// driving direction.
func (v *Vehicle) CurDriveCofs() phys.Meters {
	if v.IsFacingTrackwise() {
		return v.curPose.Cofs
	} else {
		return -v.curPose.Cofs
	}
}

// CmdTrackCofs is the commanded center offset, in Track coordinate space, not
// the vehicle's driving direction.
func (v *Vehicle) CmdTrackCofs() phys.Meters {
	return v.cmdCofs
}

// CurTrackCofs is the current center offset, in Track coordinate space, not the
// vehicle's driving direction.
func (v *Vehicle) CurTrackCofs() phys.Meters {
	return v.curPose.Cofs
}

// Reposition manually changes position and driving direction of a vehicle, as
// if a physical vehicle was picked up and move. This changes the "commanded"
// Cofs, but does NOT change the commanded driving distance speed.
func (v *Vehicle) Reposition(p track.Pose) {
	v.curPose = p
	v.desCofs = p.Cofs
	v.cmdCofs = p.Cofs
}

// SetCmdDriveDspd commands a new distance speed and acceleration, in the
// vehicle's current driving direction.
func (v *Vehicle) SetCmdDriveDspd(vs phys.MetersPerSec, va phys.MetersPerSec2) {
	v.cmdDspd = vs
	v.cmdDacl = va
}

// SetCmdDriveCofs commands a new center offset and speed, in the vehicle's
// current driving direction.
func (v *Vehicle) SetCmdDriveCofs(cofs phys.Meters, speed phys.MetersPerSec) {
	if v.IsFacingTrackwise() {
		v.cmdCofs = cofs
	} else {
		v.cmdCofs = -cofs
	}
	v.cmdCspd = speed
}

// SetCmdTrackCofs commands a new center offset and speed. The center offset is
// absolute, in Track coordinate space.
func (v *Vehicle) SetCmdTrackCofs(cofs phys.Meters, speed phys.MetersPerSec) {
	v.cmdCofs = cofs
	v.cmdCspd = speed
}

// CmdUturn commands a 180-degree uturn, toward the road center.
func (v *Vehicle) CmdUturn(radius phys.Meters) {
	// XXX(gwenz): For now, uturn is instantaneous
	tp := v.CurTrackPose()
	if tp.Cofs < 0 {
		tp.Cofs += (radius * 2)
	} else {
		tp.Cofs -= (radius * 2)
	}
	if v.IsFacingTrackwise() {
		tp.DAngle = math.Pi // counter-trackwise
	} else {
		tp.DAngle = 0 // trackwise
	}
	v.Reposition(tp)
}
