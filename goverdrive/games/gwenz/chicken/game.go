// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"fmt"
	"golang.org/x/image/colornames"
	"math"

	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/gameutil/shapes/persist"
	"github.com/anki/goverdrive/gameutil/vehlights"
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/robo/track"
	"github.com/anki/goverdrive/viz"
)

//////////////////////////////////////////////////////////////////////

type fsmState int // Finite State Machine for the game

func (s fsmState) String() string {
	nameMap := map[fsmState]string{
		stRecover:      "stRecover",
		stStartPrepare: "stStartPrepare",
		stPrepare:      "stPrepare",
		stCharge:       "stCharge",
	}
	return nameMap[s]
}

//////////////////////////////////////////////////////////////////////

const (
	kTie         = -1
	winningScore = 5

	// FSM state names
	stRecover      fsmState = iota
	stStartPrepare fsmState = iota
	stPrepare      fsmState = iota
	stCharge       fsmState = iota

	kRecoverTime        = 3 * phys.SimSecond
	kStartChargeDistMin = -0.20
	kStartChargeDistMax = -0.10
	kRoundDoneDistMin   = -0.09 // XXX: must be > kStartChargeDistMax
	kRoundDoneDistMax   = -0.00

	kCofsMiss    = 0.05  // off-center
	kCofsCollide = 0.001 // XXX: tiny bit off-center, to help with collision detection
	kCspd        = 0.1

	kDspdRecover = 0.4
	kDaclRecover = 1.0
	kDspdPrepare = 0.8
	kDaclPrepare = 0.2
	kDspdCharge  = 1.5
	kDaclCharge  = 0.15
)

type ChickenGamePhase struct {
	numVeh     int
	state      fsmState
	tStateBeg  phys.SimTime
	score      [2]int
	tSwerve    [2]phys.SimTime
	didCollide bool
	persister  *persist.Manager
}

func (gp *ChickenGamePhase) InstructionText(rys *robo.System) string {
	s := `********** CHICKEN **********
CONTROL
  Player 1 Swerve = Left  Shift Key
  Player 2 Swerve = Right Shift Key

SCORING
     Collision => Player who swerved FIRST gets a point
  No Collision => Player who swerved LAST  gets a point
`
	return s + fmt.Sprintf("  Winner is first to %d points\n", winningScore)
}

func (gp *ChickenGamePhase) Start(rsys *robo.System) {
	gp.numVeh = len(rsys.Vehicles)
	if gp.numVeh != 2 {
		panic("Chicken requires exactly two vehicles")
	}

	// Vehicle "lineup"
	dofs0 := rsys.Track.NormalizeDofs(+0.1)
	dofs1 := rsys.Track.NormalizeDofs(-0.1)
	rsys.Vehicles[0].Reposition(track.Pose{Point: track.Point{Dofs: dofs0, Cofs: +kCofsMiss}, DAngle: 0})
	rsys.Vehicles[1].Reposition(track.Pose{Point: track.Point{Dofs: dofs1, Cofs: -kCofsMiss}, DAngle: math.Pi / 2})

	gp.state = stRecover
	gp.tStateBeg = rsys.Now()
	gp.persister = persist.New()
}

func (gp *ChickenGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *ChickenGamePhase) VehRankings() []engine.VehRanking {
	rankings := make([]engine.VehRanking, gp.numVeh)
	for v := 0; v < gp.numVeh; v++ {
		rank := 1
		if gp.score[v] < gp.score[(v+1)%2] {
			rank = 2
		}
		rankings[v] = engine.VehRanking{
			VehId:       v,
			Rank:        rank,
			ScoreString: fmt.Sprintf("%v", gp.score[v]),
		}
	}
	return rankings
}

func (gp *ChickenGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	done := false

	// Display the score and status
	for v, r := range gp.VehRankings() {
		vizObj.MBText += fmt.Sprintf("%s   Score %s   Speed %.3f\n",
			rsys.Vehicles[v].Type(), r.ScoreString, rsys.Vehicles[v].CurDriveDspd())
	}
	vizObj.MBText += "\n" + gp.state.String() + "\n"

	// Gather shared FSM inputs
	isTrackwise := [2]bool{
		rsys.Vehicles[0].IsFacingTrackwise(),
		rsys.Vehicles[1].IsFacingTrackwise()}
	now := rsys.Now()
	tState := now - gp.tStateBeg // time in current state

	// Compute next state
	nextState := gp.state
	switch gp.state {
	case stStartPrepare:
		for v := range rsys.Vehicles {
			rsys.Vehicles[v].SetCmdDriveDspd(kDspdPrepare, kDaclPrepare)
			rsys.Vehicles[v].SetCmdDriveCofs(kCofsMiss, kCspd)
			rsys.Vehicles[v].Lights().Set("top", colornames.Black)
		}
		nextState = stPrepare

	case stPrepare:
		// Prepare = drive slowly off-center, until just after cars pass each other
		// (without colliding). This gives a full lap to acclerate for the "chicken
		// charge".
		deltaDofs := rsys.Track.DriveDeltaDofs(rsys.Vehicles[0].CurTrackPose(), rsys.Vehicles[1].CurTrackPose().Dofs)
		vizObj.MBText += fmt.Sprintf("deltaDofs = %v\n", deltaDofs)
		justPassed := (deltaDofs >= kStartChargeDistMin) && (deltaDofs <= kStartChargeDistMax)
		if /**/ phys.MetersAreNear(rsys.Vehicles[0].CurDriveCofs(), kCofsMiss, 0.01) &&
			/***/ phys.MetersAreNear(rsys.Vehicles[1].CurDriveCofs(), kCofsMiss, 0.01) &&
			isTrackwise[0] && !isTrackwise[1] && justPassed {
			for v := 0; v < gp.numVeh; v++ {
				gp.tSwerve[v] = 0
				rsys.Vehicles[v].SetCmdDriveDspd(kDspdCharge, kDaclCharge)
				rsys.Vehicles[v].SetCmdDriveCofs(kCofsCollide, kCspd)
			}
			gp.didCollide = false
			nextState = stCharge
		}

	case stCharge:
		// top light = speedometer
		for v := 0; v < gp.numVeh; v++ {
			clr := vehlights.SpeedometerColor(vehlights.DefSpeedometerColors, rsys.Vehicles[v].CurDriveDspd())
			rsys.Vehicles[v].Lights().Set("top", clr)
		}

		// Swerve when button is pressed, and remember who swerved first
		if (gp.tSwerve[0] == 0) && win.JustPressed(pixelgl.KeyLeftShift) {
			gp.tSwerve[0] = now
			rsys.Vehicles[0].SetCmdDriveCofs(kCofsMiss, kCspd)
			rsys.Vehicles[0].Lights().Set("top", colornames.Black)
		}
		if (gp.tSwerve[1] == 0) && win.JustPressed(pixelgl.KeyRightShift) {
			gp.tSwerve[1] = now
			rsys.Vehicles[1].SetCmdDriveCofs(kCofsMiss, kCspd)
			rsys.Vehicles[1].Lights().Set("top", colornames.Black)
		}

		// Monitor collisions
		for _, ce := range rsys.Collider.NewCollisions() {
			gp.didCollide = true
			// display impact points of current collisions, using red dot on the vehicle
			for i := 0; i < 2; i++ {
				gp.persister.Add(rsys.Now(), 1000, viz.NewCartesGameCirc(ce.VehInfo[i].Id, ce.VehInfo[i].POI, 0.01, colornames.Red, 0))
			}
		}

		deltaDofs := rsys.Track.DriveDeltaDofs(rsys.Vehicles[0].CurTrackPose(), rsys.Vehicles[1].CurTrackPose().Dofs)
		roundDone := (deltaDofs <= kRoundDoneDistMax) && (deltaDofs >= kRoundDoneDistMin)
		if roundDone {
			// determine the winner, update the score, etc
			// XXX: Assume simultaneous swerve is very unlikely

			for i := range gp.tSwerve {
				if gp.tSwerve[i] == 0 { // never swerved
					gp.tSwerve[i] = now
				}
			}
			winner := kTie
			if (gp.tSwerve[0] == now) && (gp.tSwerve[1] == now) {
				// neither car swerved => tie, even though there was a collision
				winner = kTie
			} else if gp.didCollide {
				// if there was a collision, the person who hesitated is the loser..
				// it's THEIR fault for waiting too long to avoid disaster!
				if gp.tSwerve[0] < gp.tSwerve[1] {
					winner = 0
				} else {
					winner = 1
				}
			} else {
				// no collision => the scaredy-cat who swerved first is the loser
				if gp.tSwerve[0] < gp.tSwerve[1] {
					winner = 1
				} else {
					winner = 0
				}
			}

			// update score, set vehicle lights, etc
			for v := 0; v < gp.numVeh; v++ {
				clr := colornames.Black
				if winner == v {
					clr = colornames.Limegreen
					gp.score[v]++
					if gp.score[v] >= winningScore {
						done = true
					}
				} else if winner == kTie {
					clr = colornames.Yellow
				} else {
					clr = colornames.Red
				}
				frames := []light.Frame{
					light.Frame{Color: clr, Tms: 200},
					light.Frame{Color: colornames.Black, Tms: 200},
				}
				rsys.Vehicles[v].Lights().SetAnimation(rsys.Now(), "top", frames, light.RepeatForever)
			}
			nextState = stRecover
		}

	case stRecover:
		// Make sure vehicles are driving in the opposite direction from one another
		if !isTrackwise[0] {
			rsys.Vehicles[0].CmdUturn(robo.DefUturnRadius)
		}
		if isTrackwise[1] {
			rsys.Vehicles[1].CmdUturn(robo.DefUturnRadius)
		}

		for v := 0; v < gp.numVeh; v++ {
			rsys.Vehicles[v].SetCmdDriveDspd(kDspdRecover, kDaclRecover)
			rsys.Vehicles[v].SetCmdDriveCofs(kCofsMiss, kCspd)
		}

		if (tState > kRecoverTime) && isTrackwise[0] && !isTrackwise[1] {
			nextState = stStartPrepare
		}
	}

	// Assign next state
	if nextState != gp.state {
		gp.tStateBeg = now
		gp.state = nextState
	}

	// Update persistent game shapes
	*vizObj.Shapes = *gp.persister.Update(rsys.Now())

	return done, vizObj
}
