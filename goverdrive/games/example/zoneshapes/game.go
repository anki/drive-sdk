// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"fmt"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/robo/track"
	"github.com/anki/goverdrive/viz"
)

var (
	shoulderColor = colornames.Goldenrod
	greenColor    = colornames.Green
	redColor      = colornames.Red
	purpleColor   = colornames.Purple
)

// ZoneShapesGamePhase defines zones that trigger display of various GameShape
// objects. It is designed to demonstrate:
//  - TrackRegion
//  - GameShape
type ZoneShapesGamePhase struct {
	trShoulder1 *viz.TrackRegion
	trShoulder2 *viz.TrackRegion
	trGreen     *viz.TrackRegion
	trRed       *viz.TrackRegion
	trPurple    *viz.TrackRegion
}

func (gp *ZoneShapesGamePhase) InstructionText(rys *robo.System) string {
	return `LEFT/RIGHT ARROW KEYS:  Change center offset
UP/DOWN ARROW KEYS:     Accelerate and decelerate
RIGHT SHIFT KEY:        U-turn
`
}

func (gp *ZoneShapesGamePhase) Start(rsys *robo.System) {
	// Define the (static) track regions at the start of the game

	// The "shoulders" are narrow, full-track-length regions at the sides of the
	// track, much like the shoulder of an actual road.
	shoulderWidth := phys.Meters(0.02)
	gp.trShoulder1 = &viz.TrackRegion{
		Region: *track.NewRegion(&rsys.Track, track.Point{Dofs: 0, Cofs: -rsys.Track.Width() / 2}, rsys.Track.CenLen(), shoulderWidth),
		Color:  shoulderColor,
	}
	gp.trShoulder2 = &viz.TrackRegion{
		Region: *track.NewRegion(&rsys.Track, track.Point{Dofs: 0, Cofs: +rsys.Track.Width()/2 - shoulderWidth}, rsys.Track.CenLen(), shoulderWidth),
		Color:  shoulderColor,
	}

	// The colored regions are placed at a few random places, including some
	// overlap. Their start point and size should be fine for any normal modular
	// track, but could fail hard for unusually small tracks.
	gp.trGreen = &viz.TrackRegion{
		Region: *track.NewRegion(&rsys.Track, track.Point{Dofs: 0.3, Cofs: -rsys.Track.Width() / 2}, 1.0, rsys.Track.Width()/2),
		Color:  greenColor,
	}
	gp.trRed = &viz.TrackRegion{
		Region: *track.NewRegion(&rsys.Track, track.Point{Dofs: 0.4, Cofs: -rsys.Track.Width() / 2}, 0.5, rsys.Track.Width()),
		Color:  redColor,
	}
	dofs := rsys.Track.NormalizeDofs(-1.1) // regions starts from "behind" the finish line
	gp.trPurple = &viz.TrackRegion{
		Region: *track.NewRegion(&rsys.Track, track.Point{Dofs: dofs, Cofs: 0.01}, 1.0, rsys.Track.Width()/4),
		Color:  purpleColor,
	}

	// Lineup the vehicle and start driving
	numVeh := len(rsys.Vehicles)
	if numVeh != 1 {
		panic(fmt.Sprintf("ZoneShapesGamePhase is a one-vehicle game; actual numVeh=%d", numVeh))
	}
	rsys.Vehicles[0].SetCmdDriveDspd(0.4, 0.4)
	rsys.Vehicles[0].Reposition(track.Pose{Point: track.Point{Dofs: 0, Cofs: 0}, DAngle: 0})
}

func (gp *ZoneShapesGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *ZoneShapesGamePhase) VehRankings() []engine.VehRanking {
	// This example game has no meaningful ranking concept.
	// there should only be one vehicle; see Start()
	rankings := []engine.VehRanking{
		engine.VehRanking{VehId: 0, Rank: 1, ScoreString: "0"},
	}
	return rankings
}

func (gp *ZoneShapesGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	veh := &rsys.Vehicles[0] // more concise handle to the game's only vehicle

	// Adjust driving speed arrow Up/Down arrow keys are pressed
	dspd := veh.CmdDriveDspd()
	if win.JustPressed(pixelgl.KeyUp) {
		frames := []light.Frame{light.Frame{Color: colornames.Lime, Tms: 200}}
		veh.Lights().SetAnimation(rsys.Now(), "guns", frames, 1)
		dspd += 0.1
		if dspd > 1.5 {
			dspd = 1.5
		}
		veh.SetCmdDriveDspd(dspd, 0.4)
	}
	if win.JustPressed(pixelgl.KeyDown) {
		frames := []light.Frame{light.Frame{Color: colornames.Red, Tms: 200}}
		veh.Lights().SetAnimation(rsys.Now(), "tail", frames, 1)
		dspd -= 0.1
		if dspd < 0.2 {
			dspd = 0.2
		}
		veh.SetCmdDriveDspd(dspd, 0.4)
	}
	if win.JustPressed(pixelgl.KeyRightShift) {
		veh.CmdUturn(robo.DefUturnRadius)
	}

	// Adjust center offset when Left/Right arrow keys are pressed
	cofs := veh.CmdDriveCofs()
	dCofs := phys.Meters(0)
	if win.JustPressed(pixelgl.KeyLeft) {
		dCofs = +0.02
	}
	if win.JustPressed(pixelgl.KeyRight) {
		dCofs = -0.02
	}
	veh.SetCmdDriveCofs(cofs+dCofs, 0.1)

	// Draw the track regions
	*vizObj.Regions = append(*vizObj.Regions, gp.trShoulder1)
	*vizObj.Regions = append(*vizObj.Regions, gp.trShoulder2)
	*vizObj.Regions = append(*vizObj.Regions, gp.trGreen)
	*vizObj.Regions = append(*vizObj.Regions, gp.trRed)
	*vizObj.Regions = append(*vizObj.Regions, gp.trPurple)

	// Track regions trigger game shapes that are anchored to the vehicle
	// Reminder: all lengths are in units of phys.Meters
	if gp.trShoulder1.ContainsPoint(veh.CurTrackPose().Point) {
		*vizObj.Shapes = append(*vizObj.Shapes, viz.NewTrackGameLine(0, track.Point{Dofs: -0.05, Cofs: -0.05}, track.Point{Dofs: 0.05, Cofs: -0.05}, shoulderColor, 0.005))
	}
	if gp.trShoulder2.ContainsPoint(veh.CurTrackPose().Point) {
		*vizObj.Shapes = append(*vizObj.Shapes, viz.NewTrackGameLine(0, track.Point{Dofs: -0.05, Cofs: +0.05}, track.Point{Dofs: 0.05, Cofs: +0.05}, shoulderColor, 0.005))
	}
	if gp.trGreen.ContainsPoint(veh.CurTrackPose().Point) {
		*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameLine(0, phys.Point{X: 0.05, Y: 0}, phys.Point{X: 0.10, Y: 0}, greenColor, 0.01))
	}
	if gp.trRed.ContainsPoint(veh.CurTrackPose().Point) {
		*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameCirc(0, phys.Point{X: -0.1, Y: 0}, 0.03, redColor, 0))
	}
	if gp.trPurple.ContainsPoint(veh.CurTrackPose().Point) {
		*vizObj.Shapes = append(*vizObj.Shapes, viz.NewTrackGameCirc(0, track.Point{Dofs: +0.1, Cofs: 0}, 0.03, purpleColor, 0))
	}

	// Message board text
	vizObj.MBText = gp.InstructionText(rsys)

	return false, vizObj
}
