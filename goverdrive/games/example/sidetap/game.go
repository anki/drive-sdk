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

const (
	minDspd = 0.3
	maxDspd = 1.2

	gameAccel = 0
	gameDecel = 1
	gameCofsL = 2
	gameCofsR = 3
	gameUturn = 4
)

type buttonMapType map[int]pixelgl.Button

var buttonMap [2]buttonMapType

// BumperCarsGamePhase does simple driving for a set of vehicles.
type BumperCarsGamePhase struct {
	numVeh     int
	hitPointsL [2]int
	hitPointsR [2]int
}

func (gp *BumperCarsGamePhase) InstructionText(rys *robo.System) string {
	return `Requires exactly two vehicles.
To win: Knock out all of your opponent's side lights, using collisions.

ACTION                       VEHICLE 0       VEHILCE 1
Change center offset Left    A Key           Arrow Key Left
Change center offset Right   D Key           Arrow Key Right
Accelerate                   W Key           Arrow Key Up
Brake                        S Key           Arrow Key Down
U-turn                       Left Shift Key  Right Shift Key
`
}

func (gp *BumperCarsGamePhase) Start(rsys *robo.System) {
	buttonMap = [2]buttonMapType{
		buttonMapType{ // player 0
			gameAccel: pixelgl.KeyW,
			gameDecel: pixelgl.KeyS,
			gameCofsL: pixelgl.KeyA,
			gameCofsR: pixelgl.KeyD,
			gameUturn: pixelgl.KeyLeftShift,
		},
		buttonMapType{ // player 1
			gameAccel: pixelgl.KeyUp,
			gameDecel: pixelgl.KeyDown,
			gameCofsL: pixelgl.KeyLeft,
			gameCofsR: pixelgl.KeyRight,
			gameUturn: pixelgl.KeyRightShift,
		},
	}

	gp.numVeh = len(rsys.Vehicles)
	if gp.numVeh != 2 {
		panic("BumperCarsGamePhase requires exactly 2 vehicles")
	}
	for i, _ := range rsys.Vehicles {
		lineupPoint := track.Point{
			Dofs: rsys.Track.NormalizeDofs(-0.2 * phys.Meters(i)),
			Cofs: 0}
		rsys.Vehicles[i].Reposition(track.Pose{Point: lineupPoint, DAngle: 0})
		rsys.Vehicles[i].SetCmdDriveDspd(0.4, 1.0)
		gp.hitPointsL[i] = 2
		gp.hitPointsR[i] = 2
	}
}

func (gp *BumperCarsGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *BumperCarsGamePhase) VehRankings() []engine.VehRanking {
	var score [2]int
	for v := 0; v < gp.numVeh; v++ {
		score[v] = gp.hitPointsL[v] + gp.hitPointsR[v]
	}
	rankings := make([]engine.VehRanking, gp.numVeh)
	for v := 0; v < gp.numVeh; v++ {
		rank := 1
		if score[(v+1)%2] > score[v] {
			rank = 2
		}
		rankings[v] = engine.VehRanking{
			VehId:       v,
			Rank:        rank,
			ScoreString: fmt.Sprintf("%d Hit Points", score[v])}
	}
	return rankings
}

func (gp *BumperCarsGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	isDone := false

	// Update hit point indicator lights
	for v := 0; v < gp.numVeh; v++ {
		if gp.hitPointsL[v] >= 2 {
			rsys.Vehicles[v].Lights().Set("h1", colornames.Yellow)
			rsys.Vehicles[v].Lights().Set("h2", colornames.Yellow)
		} else if gp.hitPointsL[v] == 1 {
			rsys.Vehicles[v].Lights().Set("h1", colornames.Black)
			rsys.Vehicles[v].Lights().Set("h2", colornames.Yellow)
		} else {
			rsys.Vehicles[v].Lights().Set("h1", colornames.Black)
			rsys.Vehicles[v].Lights().Set("h2", colornames.Black)
		}
		if gp.hitPointsR[v] >= 2 {
			rsys.Vehicles[v].Lights().Set("h4", colornames.Yellow)
			rsys.Vehicles[v].Lights().Set("h5", colornames.Yellow)
		} else if gp.hitPointsR[v] == 1 {
			rsys.Vehicles[v].Lights().Set("h4", colornames.Yellow)
			rsys.Vehicles[v].Lights().Set("h5", colornames.Black)
		} else {
			rsys.Vehicles[v].Lights().Set("h4", colornames.Black)
			rsys.Vehicles[v].Lights().Set("h5", colornames.Black)
		}
		if (gp.hitPointsL[v] + gp.hitPointsR[v]) == 0 {
			// you lost!
			isDone = true
		}
	}

	// Process keyboard inputs
	for v := 0; v < gp.numVeh; v++ {
		veh := &rsys.Vehicles[v]

		dspd := veh.CmdDriveDspd()
		if win.JustPressed(buttonMap[v][gameAccel]) {
			frames := []light.Frame{light.Frame{Color: colornames.Lime, Tms: 200}}
			veh.Lights().SetAnimation(rsys.Now(), "h0", frames, 1)
			dspd += 0.1
			if dspd > maxDspd {
				dspd = maxDspd
			}
			veh.SetCmdDriveDspd(dspd, 0.8)
		}
		if win.JustPressed(buttonMap[v][gameDecel]) {
			frames := []light.Frame{light.Frame{Color: colornames.Red, Tms: 200}}
			veh.Lights().SetAnimation(rsys.Now(), "h3", frames, 1)
			dspd -= 0.1
			if dspd < minDspd {
				dspd = minDspd
			}
			veh.SetCmdDriveDspd(dspd, 0.8)
		}
		if win.JustPressed(buttonMap[v][gameUturn]) {
			veh.CmdUturn(robo.DefUturnRadius)
		}

		cofs := veh.CmdDriveCofs()
		dCofs := phys.Meters(0)
		if win.JustPressed(buttonMap[v][gameCofsL]) {
			dCofs = +0.025
		}
		if win.JustPressed(buttonMap[v][gameCofsR]) {
			dCofs = -0.025
		}
		veh.SetCmdDriveCofs(cofs+dCofs, 0.1)
	}

	// Collision effects
	for _, ce := range rsys.Collider.CurCollisions() {
		// display impact points of current collisions, using red dot on the vehicle
		for i := 0; i < 2; i++ {
			*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameCirc(ce.VehInfo[i].Id, ce.VehInfo[i].POI, 0.01, colornames.Red, 0))
		}
	}
	for _, ce := range rsys.Collider.NewCollisions() {
		// change cofs/dspd when collisions happen
		for i := 0; i < 2; i++ {
			cvi := ce.VehInfo[i]
			ovi := ce.VehInfo[(i+1)%2] // other vehicle
			cv := &rsys.Vehicles[cvi.Id]
			ov := &rsys.Vehicles[ovi.Id]

			if cvi.IsRightSideCollision() {
				cv.SetCmdDriveCofs(cv.CmdDriveCofs()+0.025, 0.1)
				if gp.hitPointsR[cvi.Id] > 0 {
					gp.hitPointsR[cvi.Id]--
				}
			}

			if cvi.IsLeftSideCollision() {
				cv.SetCmdDriveCofs(cv.CmdDriveCofs()-0.025, 0.1)
				if gp.hitPointsL[cvi.Id] > 0 {
					gp.hitPointsL[cvi.Id]--
				}
			}

			if cvi.IsRearCollision() {
				dspd := ov.CurDriveDspd() + 0.2
				if dspd > maxDspd {
					dspd = maxDspd
				}
				cv.SetCmdDriveDspd(dspd, 2.0)
			}

			if cvi.IsFrontCollision() {
				cv.SetCmdDriveDspd(minDspd, 2.0)
			}
		}
	}

	// message board text
	rankings := gp.VehRankings()
	for v, veh2 := range (*rsys).Vehicles {
		vizObj.MBText += fmt.Sprintf("Veh %d    %s    %s\n", v, veh2.Type(), rankings[v].ScoreString)
	}

	return isDone, vizObj
}
