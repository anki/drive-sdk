// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"fmt"
	cn "golang.org/x/image/colornames"

	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/gameutil/follow"
	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/track"
	"github.com/anki/goverdrive/viz"
)

const (
	vLeader = 0
	vFollow = 1
	vPlayer = 2

	// form = formation
	formDspd      = 0.80 // nominal speed
	formDspdDelta = 0.10 // adjustment to maintain formation
	formDacl      = 2.0
	formCspd      = 0.05
	formDofsDelta = 0.10
	formDofsBuf   = 0.04
	formCofsDelta = 0.02
	formMaxDofs   = 0.40

	playerSlowDspd = 0.4
	playerFastDspd = 1.2
	playerDacl     = 0.5
	playerCspd     = 0.2

	newAttemptDist = 0.4
	goalRadius     = 0.04
)

type fsmState int

const (
	stAttempt = iota
	stGoalHit = iota
	stSuccess = iota
	stFail    = iota
)

// ConnectGamePhase does simple driving for a set of vehicles.
type ConnectGamePhase struct {
	numVeh        int
	score         int
	playerDesDspd phys.MetersPerSec
	playerDesCofs phys.Meters
	state         fsmState
	follower      *follow.Follower
}

func (gp *ConnectGamePhase) InstructionText(rys *robo.System) string {
	return `LEADER CAR:    Q,E Keys
FOLLOW CAR:    S,W Keys
PLAYER CAR:    Arrow keys, Right Shift Key
`
}

func (gp *ConnectGamePhase) Start(rsys *robo.System) {
	gp.numVeh = len(rsys.Vehicles)
	if gp.numVeh != 3 {
		panic("Game requires exactly 3 vehicles!")
	}

	// initial formation
	rsys.Vehicles[vLeader].Reposition(track.Pose{Point: track.Point{Dofs: 0.1, Cofs: 0}, DAngle: 0})
	rsys.Vehicles[vFollow].Reposition(track.Pose{Point: track.Point{Dofs: 0.0, Cofs: 0}, DAngle: 0})
	rsys.Vehicles[vLeader].SetCmdDriveDspd(formDspd, formDacl)
	rsys.Vehicles[vFollow].SetCmdDriveDspd(formDspd, formDacl)

	gp.playerDesDspd = playerSlowDspd
	gp.playerDesCofs = rsys.Track.Width() / 2
	rsys.Vehicles[vPlayer].Reposition(track.Pose{Point: track.Point{Dofs: 0.0, Cofs: gp.playerDesCofs}, DAngle: 0})
	rsys.Vehicles[vPlayer].SetCmdDriveDspd(gp.playerDesDspd, playerDacl)

	gp.score = 0
	gp.state = stAttempt
	gp.follower = follow.New(vLeader, vFollow, -0.4, 0.0, formDacl, formCspd, rsys.Track.CenLen(), rsys.Now(), 0)
}

func (gp *ConnectGamePhase) Stop(rsys *robo.System) {
	// no-op
}

func (gp *ConnectGamePhase) VehRankings() []engine.VehRanking {
	rankings := make([]engine.VehRanking, gp.numVeh)
	for v := 0; v < gp.numVeh; v++ {
		rankings = append(rankings, engine.VehRanking{VehId: v, Rank: v, ScoreString: "0"})
	}
	return rankings
}

func (gp *ConnectGamePhase) Update(rsys *robo.System, win *pixelgl.Window) (bool, engine.GamePhaseVizObjects) {
	vizObj := engine.EmptyGamePhaseVizObjects()
	// concise pointers to (not copies of!!) game vehicles
	lVeh := &rsys.Vehicles[vLeader]
	fVeh := &rsys.Vehicles[vFollow]
	pVeh := &rsys.Vehicles[vPlayer]

	ltpose := lVeh.CurTrackPose()
	ftpose := fVeh.CurTrackPose()
	ptpose := pVeh.CurTrackPose()

	// Adjust position of the leader car
	cofs := lVeh.CmdDriveCofs()
	dCofs := phys.Meters(0)
	if win.JustPressed(pixelgl.KeyQ) {
		dCofs = +0.025
	}
	if win.JustPressed(pixelgl.KeyE) {
		dCofs = -0.025
	}
	lVeh.SetCmdDriveCofs(cofs+dCofs, 0.1)

	// Adjust desired position of the Follow car
	followDofs := gp.follower.TargetDeltaDofs()
	if win.JustPressed(pixelgl.KeyW) {
		followDofs += formDofsDelta
	}
	if win.JustPressed(pixelgl.KeyS) {
		followDofs -= formDofsDelta
	}
	if followDofs <= (-rsys.Track.CenLen() / 2) {
		followDofs = (-rsys.Track.CenLen() / 2) + 0.001
	}
	if followDofs > formMaxDofs {
		followDofs = formMaxDofs
	}
	vizObj.MBText += fmt.Sprintf("followDofs %.3f\n", followDofs)
	gp.follower.SetTargetDeltaDofs(followDofs)
	gp.follower.Update(rsys)

	// Player controls
	// can't issue new player command until previous command is finished
	lcolor := cn.Black
	if phys.MetersPerSecAreNear(pVeh.CurDriveDspd(), gp.playerDesDspd, 0.02) &&
		phys.MetersAreNear(pVeh.CurDriveCofs(), gp.playerDesCofs, 0.002) {
		// new player command ok
		if win.JustPressed(pixelgl.KeyRightShift) {
			pVeh.CmdUturn(robo.DefUturnRadius)
		}

		// speed
		if win.JustPressed(pixelgl.KeyUp) {
			gp.playerDesDspd = playerFastDspd
		}
		if win.JustPressed(pixelgl.KeyDown) {
			gp.playerDesDspd = playerSlowDspd
		}
		pVeh.SetCmdDriveDspd(gp.playerDesDspd, playerDacl)

		// center offset
		if win.JustPressed(pixelgl.KeyLeft) {
			gp.playerDesCofs = rsys.Track.Width() / 2
		}
		if win.JustPressed(pixelgl.KeyRight) {
			gp.playerDesCofs = -(rsys.Track.Width() / 2)
		}
		pVeh.SetCmdDriveCofs(gp.playerDesCofs, playerCspd)
	} else {
		lcolor = cn.Violet
	}

	// Game state and scoring
	isFar := (rsys.Track.DofsDist(ltpose.Dofs, ptpose.Dofs) > newAttemptDist) && (rsys.Track.DofsDist(ftpose.Dofs, ptpose.Dofs) > newAttemptDist)
	didCollide := false
	for _, ce := range rsys.Collider.NewCollisions() {
		if (ce.VehInfo[0].Id == vPlayer) || (ce.VehInfo[1].Id == vPlayer) {
			didCollide = true
		}
	}

	switch gp.state {
	case stAttempt:
		deltaDofs := rsys.Track.DriveDeltaDofs(ltpose, ftpose.Dofs)
		tposeGoal := track.Pose{
			Point: track.Point{
				Dofs: ltpose.Dofs + (deltaDofs / 2),
				Cofs: (ftpose.Cofs + ltpose.Cofs) / 2},
			DAngle: 0,
		}
		pGoal := rsys.Track.ToPose(tposeGoal).Point
		*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameCirc(-1, pGoal, goalRadius, cn.White, 0.004))

		pPlayer := rsys.Track.ToPose(pVeh.CurTrackPose()).Point
		if phys.Dist(pGoal, pPlayer) < goalRadius {
			gp.state = stGoalHit
		} else if didCollide {
			gp.state = stFail
		}
	case stGoalHit:
		// Player hit the goal, but doesn't score a point until we are sure player
		// didn't later collide with one of the formation cars.
		lcolor = cn.Yellow
		plDeltaDofs := rsys.Track.DriveDeltaDofs(ptpose, ltpose.Dofs)
		pfDeltaDofs := rsys.Track.DriveDeltaDofs(ptpose, ftpose.Dofs)
		bothAhead := (plDeltaDofs > 0) && (pfDeltaDofs > 0)
		bothBehind := (plDeltaDofs < 0) && (pfDeltaDofs < 0)
		if didCollide {
			gp.state = stFail
		} else if bothAhead || bothBehind {
			gp.score++
			gp.state = stSuccess
		}
	case stSuccess:
		lcolor = cn.Limegreen
		if isFar {
			gp.state = stAttempt
		}
	case stFail:
		lcolor = cn.Red
		if isFar {
			gp.state = stAttempt
		}
	}

	// Display game state (vehicle lights, message board, etc)
	for _, ce := range rsys.Collider.CurCollisions() {
		// impact points of current collisions, using red dot on the vehicle
		for i := 0; i < 2; i++ {
			*vizObj.Shapes = append(*vizObj.Shapes, viz.NewCartesGameCirc(ce.VehInfo[i].Id, ce.VehInfo[i].POI, 0.01, cn.Red, 0))
		}
	}
	pVeh.Lights().Set("top", lcolor)
	vizObj.MBText += gp.InstructionText(rsys) + "\n"
	vizObj.MBText += fmt.Sprintf("Player Score: %d\n", gp.score)

	return false, vizObj
}
