// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"fmt"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/gameutil/lapmetrics"
	"github.com/anki/goverdrive/gameutil/vehlights"
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/robo/track"
	"github.com/anki/goverdrive/viz"
)

const (
	minDspd = 0.3
	maxDspd = 1.5
)

// DriveGamePhase does simple driving for a set of vehicles.
type DriveGamePhase struct {
	numVeh     int
	curVeh     int
	lapMetrics lapmetrics.LapMetrics
	lapText    string
}

func (gp *DriveGamePhase) InstructionText(rys *robo.System) string {
	return `SPACE BAR:              Select which vehicle to control
LEFT/RIGHT ARROW KEYS:  Change horizontal offset
UP/DOWN ARROW KEYS:     Accelerate and decelerate
RIGHT SHIFT KEY:        U-turn
`
}

func (gp *DriveGamePhase) Start(rsys *robo.System) {
	gp.lapMetrics = *lapmetrics.New(rsys.Now(), &rsys.Vehicles, true, true)
	for i, _ := range rsys.Vehicles {
		lineupPoint := track.Point{
			Dofs: rsys.Track.NormalizeDofs(-0.2 * phys.Meters(i)),
			Cofs: 0}
		rsys.Vehicles[i].Reposition(track.Pose{Point: lineupPoint, DAngle: 0})
		rsys.Vehicles[i].SetCmdDriveDspd(0.4, 1.0)
	}
	gp.numVeh = len(rsys.Vehicles)
	gp.curVeh = 0
}

func (gp *DriveGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *DriveGamePhase) VehRankings() []engine.VehRanking {
	rankings := make([]engine.VehRanking, gp.numVeh)
	for v := 0; v < gp.numVeh; v++ {
		rankings[v] = engine.VehRanking{VehId: v, Rank: v, ScoreString: "0"}
	}
	return rankings
}

func (gp *DriveGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	veh := &rsys.Vehicles[gp.curVeh]

	if win.JustPressed(pixelgl.KeySpace) {
		// advance control to next vehicle
		gp.curVeh = ((gp.curVeh + 1) % gp.numVeh)
	}

	dspd := veh.CmdDriveDspd()
	if win.JustPressed(pixelgl.KeyUp) {
		frames := []light.Frame{light.Frame{Color: colornames.Lime, Tms: 200}}
		veh.Lights().SetAnimation(rsys.Now(), "guns", frames, 1)
		dspd += 0.1
		if dspd > maxDspd {
			dspd = maxDspd
		}
		veh.SetCmdDriveDspd(dspd, 0.4)
	}
	if win.JustPressed(pixelgl.KeyDown) {
		frames := []light.Frame{light.Frame{Color: colornames.Red, Tms: 200}}
		veh.Lights().SetAnimation(rsys.Now(), "tail", frames, 1)
		dspd -= 0.1
		if dspd < minDspd {
			dspd = minDspd
		}
		veh.SetCmdDriveDspd(dspd, 0.4)
	}
	if win.JustPressed(pixelgl.KeyRightShift) {
		veh.CmdUturn(robo.DefUturnRadius)
	}

	cofs := veh.CmdDriveCofs()
	dCofs := phys.Meters(0)
	if win.JustPressed(pixelgl.KeyLeft) {
		dCofs = +0.025
	}
	if win.JustPressed(pixelgl.KeyRight) {
		dCofs = -0.025
	}
	veh.SetCmdDriveCofs(cofs+dCofs, 0.1)

	// circle around the controlled vehicle
	*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameCirc(gp.curVeh, phys.Point{X: 0, Y: 0}, 0.05, colornames.White, 0.004))

	// speedometer light
	clr := vehlights.SpeedometerColor(vehlights.DefSpeedometerColors, veh.CurDriveDspd())
	veh.Lights().Set("top", clr)

	// lap counts
	gp.lapMetrics.Update(rsys.Now(), &rsys.Track, &rsys.Vehicles)
	for v := range rsys.Vehicles {
		newlaps := gp.lapMetrics.NewCompletedLapInfo(v)
		for _, li := range newlaps {
			gp.lapText = fmt.Sprintf("Veh %d lap completed: %s\n", v, li.String()) + gp.lapText
		}
	}

	// message board text
	vizObj.MBText = gp.InstructionText(rsys) + "\n" + gp.lapText

	return false, vizObj
}
