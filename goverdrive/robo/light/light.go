// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package light supports lights for an individual vehicle.
//
// Note that light do NOT have any dependency on the visualization system code.
// The light package is designed to not depend on the vizualization.
package light

import (
	"fmt"
	"golang.org/x/image/colornames"
	"image/color"

	"github.com/anki/goverdrive/phys"
)

const (
	RepeatForever = -1
)

// Position is the position and size of a "point" light on a vehicle. Position
// is relative to vehicle's center:
//   X>0 means towards vehicle front
//   Y>0 means towards vehicle left side
//   R = radius
type Position struct {
	X phys.Meters
	Y phys.Meters
	R phys.Meters
}

// Group is a group of lights have the same color at any time, even though they
// may be located in different positions on the vehicle. For example, OverDrive
// Gen2 vehicles have two separate "gun" lights that share one hardware control
// signal.
type Group struct {
	defColor color.Color
	lights   []Position
}

// Spec is the spec for the full set of lights for the vehicle. The
// handle to each light group in the set is a string name.
type Spec map[string]Group

// Gen2Spec is matches the real lights on OverDrive (Gen2) vehicle hardware.
var Gen2Spec = Spec{
	"top": Group{defColor: colornames.Black,
		lights: []Position{
			Position{X: 0, Y: 0, R: 0.01},
		}},
	"guns": Group{defColor: colornames.Goldenrod,
		lights: []Position{
			Position{X: 0.035, Y: -0.015, R: 0.005},
			Position{X: 0.035, Y: +0.015, R: 0.005},
		}},
	"tail": Group{defColor: colornames.Darkred,
		lights: []Position{
			Position{X: -0.034, Y: -0.015, R: 0.006},
			Position{X: -0.034, Y: -0.009, R: 0.006},
			Position{X: -0.034, Y: +0.009, R: 0.006},
			Position{X: -0.034, Y: +0.015, R: 0.006},
		}},
}

// HexPodSpec is a hexagon pod with a center light, ie seven lights total. The
// pod lights are enumerated is counter-clockwise order, starting with the
// "forward" light.
//
// NOTE: Presently, this light configuration does not exist with real hardware.
var HexPodSpec = Spec{
	"center": Group{defColor: colornames.Black, lights: []Position{Position{X: +0.0000, Y: +0.0000, R: 0.006}}},
	"h0":     Group{defColor: colornames.Black, lights: []Position{Position{X: +0.0150, Y: +0.0000, R: 0.006}}},
	"h1":     Group{defColor: colornames.Black, lights: []Position{Position{X: +0.0076, Y: +0.0106, R: 0.006}}},
	"h2":     Group{defColor: colornames.Black, lights: []Position{Position{X: -0.0076, Y: +0.0106, R: 0.006}}},
	"h3":     Group{defColor: colornames.Black, lights: []Position{Position{X: -0.0150, Y: +0.0000, R: 0.007}}},
	"h4":     Group{defColor: colornames.Black, lights: []Position{Position{X: -0.0076, Y: -0.0106, R: 0.006}}},
	"h5":     Group{defColor: colornames.Black, lights: []Position{Position{X: +0.0076, Y: -0.0106, R: 0.006}}},
}

//////////////////////////////////////////////////////////////////////

// Frame is a single "frame" of an animation for a single light
type Frame struct {
	Color color.Color
	Tms   uint // duration, in milliseconds
}

// GroupFrame is a "frame" of an animation for a group of independent lights
type GroupFrame struct {
	Colors []color.Color
	Tms    uint // duration, in milliseconds
}

// animation has frames and internal state to play an animation on a single
// light
type animation struct {
	frames       []Frame
	curFrame     int
	frameEndTime phys.SimTime
	countLeft    int
}

// VehLights has the physical spec and state of the set of lights for one
// vehicle.
type VehLights struct {
	spec   Spec
	static map[string]color.Color // light name -> color
	anim   map[string]*animation  // light name -> animation
	cur    map[string]color.Color // light name -> color
}

// VehLightState has all of the information needed to visualize one point light.
type VizInfo struct {
	Position
	Color color.Color
}

// NewVehLights creates a usable VehLights object based on a spec.
func NewVehLights(spec Spec) *VehLights {
	// static lights start with default colors
	static := make(map[string]color.Color)
	cur := make(map[string]color.Color)
	for k, v := range spec {
		static[k] = v.defColor
		cur[k] = v.defColor
	}
	return &VehLights{
		spec:   spec,
		static: static,
		anim:   make(map[string]*animation),
		cur:    cur,
	}
}

func (vl *VehLights) validateName(name string) {
	if _, ok := vl.spec[name]; !ok {
		panic(fmt.Sprintf("VehLights.Set(%v) failed, light name not recognized", name))
	}
}

// Set sets the static color of a light group. This cancels any ongoing
// animation for the light.
func (vl *VehLights) Set(name string, color color.Color) {
	vl.validateName(name)
	vl.anim[name] = nil
	vl.static[name] = color
}

// SetAnimation starts animation of one or more "frames" for a single light. The
// animation is repeated a programmable number of times, or indefinitely if
// repeatCount=RepeatForever.
func (vl *VehLights) SetAnimation(now phys.SimTime, name string, frames []Frame, repeatCount int) {
	vl.validateName(name)
	if len(frames) == 0 {
		panic("SetAnimation with len(frames)=0 is invalid")
	}
	vl.anim[name] = startAnimation(now, frames, repeatCount)
}

// SetGroupAnimation starts animation of one or more "frames" for a group of
// independent lights. The animation is repeated a programmable number of times,
// or indefinitely if repeatCount=RepeatForever.
func (vl *VehLights) SetGroupAnimation(now phys.SimTime, names []string, gframes []GroupFrame, repeatCount int) {
	if len(gframes) == 0 {
		panic("SetGroupAnimation with len(gframes)=0 is invalid")
	}
	for l, name := range names {
		vl.validateName(name)
		frames := make([]Frame, len(gframes))
		for i := range gframes {
			frames[i].Color = gframes[i].Colors[l]
			frames[i].Tms = gframes[i].Tms
		}
		vl.anim[name] = startAnimation(now, frames, repeatCount)
	}
}

// IsAnimating returns true if a named light has an ongoing animation.
func (vl *VehLights) IsAnimating(name string) bool {
	vl.validateName(name)
	return vl.anim[name] != nil
}

// Update updates all light animations and determines the current color of each
// light.
func (vl *VehLights) Update(now phys.SimTime) {
	// update light animation and choose final color value for each light
	for name, _ := range vl.spec {
		vl.cur[name] = vl.static[name]
		// animations take precedence over static color value
		if anim := vl.anim[name]; anim != nil {
			vl.cur[name] = anim.updateAnimation(now)
			if anim.isDone() {
				vl.anim[name] = nil
				vl.cur[name] = vl.static[name]
			}
		}
	}
}

// VizInfo returns the info to visuzlize for each individual point light of the
// vehicle.
func (vl *VehLights) VizInfo() []*VizInfo {
	vizinfo := make([]*VizInfo, 0)
	for name, color := range vl.cur {
		for _, lp := range vl.spec[name].lights {
			vizinfo = append(vizinfo, &VizInfo{Position: Position{X: lp.X, Y: lp.Y, R: lp.R}, Color: color})
		}
	}
	return vizinfo
}

func startAnimation(now phys.SimTime, frames []Frame, repeatCount int) *animation {
	return &animation{
		frames:       frames,
		curFrame:     0,
		frameEndTime: now + (phys.SimTime(frames[0].Tms) * phys.SimMillisecond),
		countLeft:    repeatCount,
	}
}

func (a *animation) isDone() bool {
	return a.countLeft == 0
}

// updateAnimation advances the animation based on the current sim time, and
// returns the updated light color.
func (a *animation) updateAnimation(now phys.SimTime) color.Color {
	if a.isDone() {
		return color.RGBA{0, 0, 0, 0}
	}

	for now >= a.frameEndTime {
		// frame is done => next frame
		a.curFrame++
		if a.curFrame >= len(a.frames) {
			a.curFrame = 0
			if a.countLeft > 0 { // <0 means repeat forever
				a.countLeft--
			}
		}
		a.frameEndTime += (phys.SimTime(a.frames[a.curFrame].Tms) * phys.SimMillisecond)
	}
	return a.frames[a.curFrame].Color
}
