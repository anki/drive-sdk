// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"fmt"
	"golang.org/x/image/colornames"
	"math"

	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/track"
	"github.com/anki/goverdrive/viz"
)

const (
	dDofs = 0.10
	dCofs = 0.02
)

// MoverGamePhase does simple driving for a set of vehicles.
type MoverGamePhase struct {
	numVeh int
	curVeh int
}

func (gp *MoverGamePhase) InstructionText(rys *robo.System) string {
	return `SPACE BAR:              Select which vehicle to move
LEFT/RIGHT ARROW KEYS:  Change center   offset
UP/DOWN ARROW KEYS:     Change distance offset
RIGHT SHIFT KEY:        U-turn
`
}

func (gp *MoverGamePhase) Start(rsys *robo.System) {
	for i, _ := range rsys.Vehicles {
		rsys.Vehicles[i].Reposition(track.Pose{Point: track.Point{Dofs: 0, Cofs: 0}, DAngle: 0})
		rsys.Vehicles[i].SetCmdDriveCofs(0, 0)
		rsys.Vehicles[i].SetCmdDriveDspd(0, 0)
	}
	gp.numVeh = len(rsys.Vehicles)
	gp.curVeh = 0
}

func (gp *MoverGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *MoverGamePhase) VehRankings() []engine.VehRanking {
	rankings := make([]engine.VehRanking, gp.numVeh)
	for v := 0; v < gp.numVeh; v++ {
		rankings = append(rankings, engine.VehRanking{VehId: v, Rank: v, ScoreString: "0"})
	}
	return rankings
}

func (gp *MoverGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	veh := &rsys.Vehicles[gp.curVeh]

	if win.JustPressed(pixelgl.KeySpace) {
		// advance control to next vehicle
		gp.curVeh = ((gp.curVeh + 1) % gp.numVeh)
	}

	tpose := veh.CurTrackPose()
	if win.JustPressed(pixelgl.KeyRightShift) {
		tpose.DAngle = phys.NormalizeRadians(tpose.DAngle + math.Pi)
	}
	if win.JustPressed(pixelgl.KeyUp) {
		tpose.Dofs = rsys.Track.NormalizeDofs(tpose.Dofs + dDofs)
	}
	if win.JustPressed(pixelgl.KeyDown) {
		tpose.Dofs = rsys.Track.NormalizeDofs(tpose.Dofs - dDofs)
	}
	if win.JustPressed(pixelgl.KeyLeft) {
		tpose.Cofs += dCofs
	}
	if win.JustPressed(pixelgl.KeyRight) {
		tpose.Cofs -= dCofs
	}
	veh.Reposition(tpose)

	// circle in the controlled vehicle
	*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameCirc(gp.curVeh, phys.Point{X: 0, Y: 0}, 0.01, colornames.White, 0))

	// message board text
	for _, veh2 := range (*rsys).Vehicles {
		tpose := veh2.CurTrackPose()
		pose := rsys.Track.ToPose(tpose)
		vizObj.MBText += fmt.Sprintf("%s:  Dofs=%.3f  Cofs=%+.3f   X=%+.3f  Y=%+.3f  Th=%+.5f\n",
			veh2.Type(), tpose.Dofs, tpose.Cofs, pose.X, pose.Y, pose.Theta)
	}
	if len(rsys.Vehicles) > 1 {
		frVeh := &rsys.Vehicles[0]
		toVeh := &rsys.Vehicles[1]
		vizObj.MBText += fmt.Sprintf("%s-->%s: ", frVeh.Type(), toVeh.Type())
		vizObj.MBText += fmt.Sprintf("DofsDist=%.3f  ", rsys.Track.DofsDist(frVeh.CurTrackPose().Dofs, toVeh.CurTrackPose().Dofs))
		vizObj.MBText += fmt.Sprintf("DrDofsDist=%.3f  ", rsys.Track.DriveDofsDist(frVeh.CurTrackPose(), toVeh.CurTrackPose().Dofs))
		vizObj.MBText += fmt.Sprintf("DrDeltaDofs=%.3f  ", rsys.Track.DriveDeltaDofs(frVeh.CurTrackPose(), toVeh.CurTrackPose().Dofs))
		vizObj.MBText += fmt.Sprintf("DrDeltaCofs=%.3f  ", rsys.Track.DriveDeltaCofs(frVeh.CurTrackPose(), toVeh.CurTrackPose().Cofs))
		vizObj.MBText += fmt.Sprintf("DrDist=%.3f  ", rsys.Track.DriveDist(frVeh.CurTrackPose(), toVeh.CurTrackPose().Dofs))
		vizObj.MBText += fmt.Sprintf("DrDeltaDist=%.3f  ", rsys.Track.DriveDeltaDist(frVeh.CurTrackPose(), toVeh.CurTrackPose().Dofs))
		vizObj.MBText += "\n"
	}
	vizObj.MBText += "\n" + gp.InstructionText(rsys)

	return false, vizObj
}
