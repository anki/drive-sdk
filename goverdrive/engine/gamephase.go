// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package engine

import (
	"fmt"

	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/viz"
	"github.com/faiface/pixel/pixelgl"
)

// VehRanking is for reporting the rank of a vehicle, compared to other
// vehicles. The precise meaning of ranking is intentionally vague.
type VehRanking struct {
	VehId       int    // handle to the vehicle
	Rank        int    // (typical) 1=1st place, 2=2nd place, etc
	ScoreString string // arbitrary string representation of a vehicle's score (eg "10 points")
}

func (vr *VehRanking) String() string {
	return fmt.Sprintf("VehId %d:  Rank: %d  Score: %s", vr.VehId, vr.Rank, vr.ScoreString)
}

// VehRankingSorter is a wrapper to use sort.Interface
type VehRankingSorter struct {
	Rankings []VehRanking
}

// Len implements function needed for sort.Interface
func (vrs *VehRankingSorter) Len() int {
	return len(vrs.Rankings)
}

// Less implements comparison needed for sort.Interface
func (vrs *VehRankingSorter) Less(i, j int) bool {
	return vrs.Rankings[i].Rank < vrs.Rankings[j].Rank
}

// Swap implements function needed for sort.Interface
func (vrs *VehRankingSorter) Swap(i, j int) {
	vrs.Rankings[i], vrs.Rankings[j] = vrs.Rankings[j], vrs.Rankings[i]
}

//////////////////////////////////////////////////////////////////////

// GamePhaseVizObjects is a wrapper with all of the game-specific objects
// (beyond the built-in track and vehicles) from the GamePhase, which need to
// visualized at a particular moment in time.
type GamePhaseVizObjects struct {
	Regions *[]*viz.TrackRegion
	Shapes  *[]*viz.GameShape
	MBText  string // message board
}

// EmptyGamePhaseVizObjects returns a GamePhaseVizObjects that has been properly
// initialized with empty slices.
func EmptyGamePhaseVizObjects() GamePhaseVizObjects {
	emptyReg := make([]*viz.TrackRegion, 0)
	emptyShp := make([]*viz.GameShape, 0)
	return GamePhaseVizObjects{
		Regions: &emptyReg,
		Shapes:  &emptyShp,
		MBText:  "",
	}
}

//////////////////////////////////////////////////////////////////////

// GamePhase is the meat of gameplay. It drives interactions between the track
// and vehicles. A full "game" has one or more game phases.
//   - The same track is used for all game phases
//   - The same set of vehicles is used for all game phases
//   - The state of the vehicles can be preserved between game phases
//   - The vehicles have a ranking
type GamePhase interface {
	// InstructionText returns a text instruction string for the game phase. It
	// can be multi-line.
	InstructionText(rsys *robo.System) string

	// Start does any one-time setup needed for the game phase. No time passes,
	// but the vehicle states may be changed, eg to reposition for a fake lineup.
	Start(rsys *robo.System)

	// Update is the "tick" to run the game logic.
	//   - The robotics system is available to query and command; it includes the time
	//   - User input can be retrieved from the window
	//   - Game-specific objects are returned for visualization
	//   - When the game phase is done, true is returned
	Update(rsys *robo.System, win *pixelgl.Window) (bool, GamePhaseVizObjects)

	// Stop terminates the game phase, and computes final vehicle rankings. No
	// time passes.
	Stop(rsys *robo.System)

	// VehRankings returns the ranking of each vehicle.
	VehRankings() []VehRanking
}
