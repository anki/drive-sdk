// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"fmt"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/gameutil/follow"
	"github.com/anki/goverdrive/gameutil/vehlights"
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/robo/track"
	_ "github.com/anki/goverdrive/viz"
)

const (
	minDspd       = 0.3
	maxDspd       = 1.5
	followDacl    = 3.0
	followCspd    = 0.1
	numVeh        = 4
	numFormations = 9
)

// FourmationGamePhase drives a set of four vehicles in a formation.
type FourmationGamePhase struct {
	followers    []*follow.Follower
	curFormation int
}

type formation struct {
	dDofs1 phys.Meters
	dCofs1 phys.Meters
	dDofs2 phys.Meters
	dCofs2 phys.Meters
	dDofs3 phys.Meters
	dCofs3 phys.Meters
}

func (gp *FourmationGamePhase) changeFormation(f int) {
	if f > numFormations {
		panic(fmt.Sprintf("Formation number %d is invalid", f))
	}
	formationList := [numFormations]formation{
		formation{dDofs1: -0.10, dCofs1: +0.10, dDofs2: -0.10, dCofs2: -0.10, dDofs3: -0.20, dCofs3: +0.00}, // diamond
		formation{dDofs1: -0.20, dCofs1: +0.10, dDofs2: -0.20, dCofs2: -0.10, dDofs3: -0.10, dCofs3: +0.00}, // Y
		formation{dDofs1: -0.60, dCofs1: +0.00, dDofs2: -0.40, dCofs2: +0.00, dDofs3: -0.20, dCofs3: +0.00}, // vert line
		formation{dDofs1: -0.00, dCofs1: +0.10, dDofs2: -0.00, dCofs2: +0.20, dDofs3: -0.00, dCofs3: -0.10}, // horz line
		formation{dDofs1: -0.00, dCofs1: +0.10, dDofs2: -0.13, dCofs2: +0.10, dDofs3: -0.13, dCofs3: +0.00}, // square
		formation{dDofs1: -0.00, dCofs1: +0.10, dDofs2: -0.13, dCofs2: +0.10, dDofs3: +0.00, dCofs3: -0.10}, // L
		formation{dDofs1: -0.13, dCofs1: +0.00, dDofs2: -0.13, dCofs2: -0.10, dDofs3: +0.13, dCofs3: -0.00}, // L rotated
		formation{dDofs1: -0.00, dCofs1: +0.10, dDofs2: -0.13, dCofs2: +0.00, dDofs3: -0.13, dCofs3: -0.10}, // Z
		formation{dDofs1: -0.13, dCofs1: +0.10, dDofs2: -0.13, dCofs2: +0.00, dDofs3: -0.00, dCofs3: -0.10}, // Z rotated
	}
	fmtn := formationList[f]
	gp.followers[0].SetTargetDeltaDofs(fmtn.dDofs1)
	gp.followers[0].SetTargetDeltaCofs(fmtn.dCofs1)
	gp.followers[1].SetTargetDeltaDofs(fmtn.dDofs2)
	gp.followers[1].SetTargetDeltaCofs(fmtn.dCofs2)
	gp.followers[2].SetTargetDeltaDofs(fmtn.dDofs3)
	gp.followers[2].SetTargetDeltaCofs(fmtn.dCofs3)
}

func (gp *FourmationGamePhase) InstructionText(rys *robo.System) string {
	return `SPACE BAR:              Next formation
LEFT/RIGHT ARROW KEYS:  Change horizontal offset
UP/DOWN ARROW KEYS:     Accelerate and decelerate
RIGHT SHIFT KEY:        U-turn
`
}

func (gp *FourmationGamePhase) Start(rsys *robo.System) {
	if len(rsys.Vehicles) != numVeh {
		panic(fmt.Sprintf("Exactly %d vehicles are required", numVeh))
	}

	for i, _ := range rsys.Vehicles {
		lineupPoint := track.Point{
			Dofs: rsys.Track.NormalizeDofs(-0.2 * phys.Meters(i)),
			Cofs: 0}
		rsys.Vehicles[i].Reposition(track.Pose{Point: lineupPoint, DAngle: 0})
		rsys.Vehicles[i].SetCmdDriveDspd(0.4, 1.0)
	}

	gp.followers = make([]*follow.Follower, numVeh-1)
	for v := 0; v < (numVeh - 1); v++ {
		vFollow := v + 1
		deltaDofs := phys.Meters(-0.2 - (float32(v) * 0.2))
		deltaCofs := phys.Meters(0)
		gp.followers[v] = follow.New(0, vFollow, deltaDofs, deltaCofs, followDacl, followCspd, rsys.Track.CenLen(), rsys.Now(), 0)
	}
	gp.curFormation = 0
	gp.changeFormation(gp.curFormation)
}

func (gp *FourmationGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *FourmationGamePhase) VehRankings() []engine.VehRanking {
	rankings := make([]engine.VehRanking, numVeh)
	for v := 0; v < numVeh; v++ {
		rankings[v] = engine.VehRanking{VehId: v, Rank: v, ScoreString: "0"}
	}
	return rankings
}

func (gp *FourmationGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	veh := &rsys.Vehicles[0]

	if win.JustPressed(pixelgl.KeySpace) {
		gp.curFormation = (gp.curFormation + 1) % numFormations
		gp.changeFormation(gp.curFormation)
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

	// followers
	for v := 1; v < numVeh; v++ {
		gp.followers[v-1].Update(rsys)
	}

	// speedometer light
	clr := vehlights.SpeedometerColor(vehlights.DefSpeedometerColors, veh.CurDriveDspd())
	veh.Lights().Set("top", clr)

	// message board text
	vizObj.MBText += fmt.Sprintf("Formation %d\n", gp.curFormation)
	vizObj.MBText += gp.InstructionText(rsys) + "\n"

	return false, vizObj
}
