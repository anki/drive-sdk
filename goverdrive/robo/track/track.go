// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package track represents a physical track, which is comprised of road pieces.
// It also defines an internal coordinate system, track regions, and other
// concepts to help express game design concepts elegantly.
package track

import (
	"fmt"
	"math"

	"github.com/anki/goverdrive/phys"
)

const (
	// TrackLenMod... are parameters for standard OverDrive modular tracks
	TrackLenModStartLong  phys.Meters = 0.34
	TrackLenModStartShort phys.Meters = 0.22
	TrackLenModStraight   phys.Meters = 0.56
	TrackLenModCurve      phys.Meters = phys.Meters(math.Pi/2) * (TrackLenModStraight / 2)
)

// Rpi means Road Piece Index, ie a unique index for each road piece of the
// track. An Rpi value that <0 means invalid or not found, depending on the
// context.
type Rpi int

// Track is a representation of a physical track.
type Track struct {
	width      phys.Meters   // total width, including border lanes
	maxCofs    phys.Meters   // maximum ABS(Cofs) a vehicle can have
	pieces     []RoadPiece   // in trackwise driving order; finish line = start of pieces[0]
	entryPoses []phys.Pose   // in trackwise driving order
	entryDofs  []phys.Meters // at piece entry: distance offset from finish line, along road center
	minCorner  phys.Point    // minimum corner of the track (ie bottom-left)
	maxCorner  phys.Point    // maximum corner of the track (ie upper-right)
}

// NewTrack creates a track with a fixed width and a set of consecutive road
// pieces.
func NewTrack(width phys.Meters, maxCofs phys.Meters, pieces []RoadPiece) (*Track, error) {
	// Sanity checking for input args
	numRp := len(pieces)
	if numRp < 4 {
		return nil, fmt.Errorf("Track with %v road pieces is too small", numRp)
	}
	if width <= 0 {
		return nil, fmt.Errorf("Invalid track width: %v", width)
	}
	if maxCofs == 0 {
		maxCofs = width / 2
	}
	if maxCofs < 0 {
		return nil, fmt.Errorf("maxCofs=%v invalid; must be > 0", maxCofs)
	}

	// compute per-piece info needed by the representation
	t := Track{
		width:      width,
		maxCofs:    maxCofs,
		pieces:     pieces,
		entryPoses: make([]phys.Pose, numRp+1, numRp+1),
		entryDofs:  make([]phys.Meters, numRp+1, numRp+1),
	}

	// the finish line is ALWAYS at the origin, facing right
	t.entryPoses[0] = phys.Pose{Point: phys.Point{X: 0, Y: 0}, Theta: 0}
	t.entryDofs[0] = 0
	for i, rp := range pieces {
		t.entryDofs[i+1] = t.entryDofs[i] + rp.Len(0)
		t.entryPoses[i+1] = t.entryPoses[i].AdvancePose(rp.DeltaPose())
	}

	// Determine the corners of the "world", by examining track edges at road
	// piece boundaries.
	for i, _ := range pieces {
		for j := 0; j < 2; j++ {
			deltaPose := phys.Pose{Point: phys.Point{X: 0, Y: (-width / 2) + phys.Meters(j)*width}, Theta: 0}
			edge := t.entryPoses[i].AdvancePose(deltaPose)
			if edge.X < t.minCorner.X {
				t.minCorner.X = edge.X
			}
			if edge.Y < t.minCorner.Y {
				t.minCorner.Y = edge.Y
			}
			if edge.X > t.maxCorner.X {
				t.maxCorner.X = edge.X
			}
			if edge.Y > t.maxCorner.Y {
				t.maxCorner.Y = edge.Y
			}
		}
	}

	// The pose after doing a lap along road center should match starting pose
	if phys.RadiansAreNear(t.entryPoses[0].Theta, t.entryPoses[numRp].Theta, TrackRadiansAreEqualTol) &&
		phys.MetersAreNear(t.entryPoses[0].X, t.entryPoses[numRp].X, TrackMetersAreEqualTol) &&
		phys.MetersAreNear(t.entryPoses[0].Y, t.entryPoses[numRp].Y, TrackMetersAreEqualTol) {
		return &t, nil
	}

	return &t, fmt.Errorf("Track is not a loop: beg pose = %s, end pose =%s",
		t.entryPoses[0].String(), t.entryPoses[numRp].String())
}

// Width returns the width of the track.
func (t *Track) Width() phys.Meters {
	return t.width
}

// MaxCofs returns the maximum allowed horizontal offset. This may be greater or
// less than Track.Width()/2.
func (t *Track) MaxCofs() phys.Meters {
	return t.maxCofs
}

// NumRp returns the number of road pieces in the track.
func (t *Track) NumRp() int {
	return len(t.pieces)
}

// CenLen returns the track length along road center.
func (t *Track) CenLen() phys.Meters {
	return t.entryDofs[len(t.pieces)]
}

// Len computes the total driving path length of the track, along a given
// horizontal offset.
func (t *Track) Len(cofs phys.Meters) phys.Meters {
	var len phys.Meters = 0
	for _, rp := range t.pieces {
		len += rp.Len(cofs)
	}
	return len
}

// MinCorner is the botom-left corner of the rectangle that completely encloses
// the track.
func (t *Track) MinCorner() phys.Point {
	return t.minCorner
}

// MaxCorner is the upper-right corner of the rectangle that completely encloses
// the track.
func (t *Track) MaxCorner() phys.Point {
	return t.maxCorner
}

// RpiAt returns the Road Piece Index corresponding to a distance offset.
func (t *Track) RpiAt(dofs phys.Meters) Rpi {
	// XXX: Linear search
	for i, _ := range t.entryDofs {
		if t.entryDofs[i] > dofs {
			return Rpi(i - 1)
		}
	}
	panic(fmt.Sprintf("RpiAt(%v) with track len %v: Could not find road piece index", dofs, t.entryDofs[len(t.pieces)]))
}

// assertValidDofs causes a panic of the Dofs value is not in an appropriate
// range for the track.
func (t *Track) assertValidDofs(dofs phys.Meters) {
	if dofs < 0 {
		panic(fmt.Sprintf("dofs=%v is invalid, must be >= 0", dofs))
	}
	if dofs > t.CenLen() {
		panic(fmt.Sprintf("dofs=%v is invalid, must be <= track.CenLen()=%v", dofs, t.CenLen()))
	}
}

// assertValidRpi causes a panic if the index is not valid for the track.
func (t *Track) assertValidRpi(i Rpi) {
	if int(i) >= len(t.pieces) {
		panic(fmt.Sprintf("rpi==%d is not valid for track with only %d pieces", i, len(t.pieces)))
	}
}

// Rp returns a Road Piece in the track, from it's Road Piece Index.
func (t *Track) Rp(i Rpi) RoadPiece {
	t.assertValidRpi(i)
	return t.pieces[i]
}

// RpEntryDofs returns the Dofs value at the start of the road piece, in the
// trackwise driving direction.
func (t *Track) RpEntryDofs(i Rpi) phys.Meters {
	t.assertValidRpi(i)
	return t.entryDofs[i]
}

// RpEntryPose returns the pose of the vehicle when it enters a road piece. This
// pose is for road center, trackwise driving direction.
func (t *Track) RpEntryPose(i Rpi) phys.Pose {
	if int(i) == len(t.pieces) {
		i = 0
	}
	t.assertValidRpi(i)
	return t.entryPoses[i]
}

// RpCurveCenter returns the center point of the cirlce's radius of curvature.
// Straight pieces, which have no curvature, return the point of entry.
func (t *Track) RpCurveCenter(i Rpi) phys.Point {
	t.assertValidRpi(i)
	rp := t.pieces[i]

	// special case: straight
	if rp.DAngle() == 0 {
		return phys.Point{X: t.entryPoses[i].X, Y: t.entryPoses[i].Y}
	}

	dy := rp.CurveRadius(0)
	if rp.DAngle() < 0 {
		dy = -dy
	}
	pose := t.entryPoses[i].AdvancePose(phys.Pose{Point: phys.Point{X: 0, Y: dy}, Theta: 0})
	return phys.Point{X: pose.X, Y: pose.Y}
}

// RpiAndRpDofs maps a distance offset to a road piece and road-center distance
// into the road piece, for trackwise driving direction.
func (t *Track) RpiAndRpDofs(dofs phys.Meters) (Rpi, phys.Meters) {
	t.assertValidDofs(dofs)
	rpi := len(t.pieces) - 1
	for ; (rpi > 0) && (t.entryDofs[rpi] > dofs); rpi-- {
	}
	rpDofs := dofs - t.entryDofs[rpi]
	if rpDofs > t.pieces[rpi].CenLen() {
		// XXX: This can happen due to arithmetic precision errors
		rpDofs = t.pieces[rpi].CenLen()
	}
	return Rpi(rpi), rpDofs
}

// ToPose converts a Pose to a canonical Cartesian pose.
func (t *Track) ToPose(tp Pose) phys.Pose {
	tp.Dofs = t.NormalizeDofs(tp.Dofs)
	rpi, rpDofs := t.RpiAndRpDofs(tp.Dofs)
	rp := t.Rp(rpi)
	pose := t.RpEntryPose(rpi)

	// adjust Dofs
	percent := float64(rpDofs / rp.CenLen())
	if math.Abs(percent) > 1.0e-6 {
		rp2 := NewRoadPiece(phys.Meters(percent)*rp.CenLen(), phys.Radians(percent*float64(rp.DAngle())))
		pose = pose.AdvancePose(rp2.DeltaPose())
	}

	// adjust Cofs
	pose = pose.AdvancePose(phys.Pose{Point: phys.Point{X: 0, Y: tp.Cofs}, Theta: 0})

	// adjust Theta
	pose.Theta = phys.NormalizeRadians(pose.Theta + tp.DAngle)

	return pose
}

// NormalizeDofs adjusts a Dofs by increments of the track len, to make sure
// that (0 <= Dofs < track.CenLen()).
func (t *Track) NormalizeDofs(dofs phys.Meters) phys.Meters {
	for ; dofs < 0; dofs += t.CenLen() {
	}
	for ; dofs >= t.CenLen(); dofs -= t.CenLen() {
	}
	return dofs
}

// DofsDist returns the shortest distance between two track distance offsets,
// taking the looping nature of the track into account. Always positive.
func (t *Track) DofsDist(dofs1, dofs2 phys.Meters) phys.Meters {
	dist := math.Abs(float64(dofs1) - float64(dofs2))
	if dist > (float64(t.CenLen()) / 2) {
		dist = float64(t.CenLen()) - dist
	}
	return phys.Meters(dist)
}

func (t *Track) isFacingTrackwise(dangle phys.Radians) bool {
	if (dangle > math.Pi) || (dangle < -math.Pi) {
		panic(fmt.Sprintf("isFacingTrackwise(dangle==%v) is invalid; must be in range [-pi,pi]", dangle))
	}
	return (math.Abs(float64(dangle)) <= (math.Pi / 2))
}

// DriveDofsDist returns the Dofs distance needed to drive from Point A to dofs,
// in the track direction that Pose A is facing. For example, if dofs is a
// little behind Point A, then A would need to drive most of the track length to
// get to it.
func (t *Track) DriveDofsDist(a Pose, dofs phys.Meters) phys.Meters {
	if t.isFacingTrackwise(a.DAngle) {
		return t.NormalizeDofs(dofs - a.Dofs)
	}
	return t.NormalizeDofs(a.Dofs - dofs)
}

// DriveDist calculates the driving distance needed to reach dofs starting from
// p, driving along center offset p.Cofs, in direction p.DAngle. Always
// positive.
func (t *Track) DriveDist(p Pose, dofs phys.Meters) phys.Meters {
	dofs1 := p.Dofs
	dofs2 := dofs
	if !t.isFacingTrackwise(p.DAngle) {
		dofs1, dofs2 = dofs2, dofs1
	}
	rpi1, rpDofs1 := t.RpiAndRpDofs(dofs1)
	rpi2, rpDofs2 := t.RpiAndRpDofs(dofs2)

	if (rpi1 == rpi2) && (dofs2 >= dofs1) {
		dist := dofs2 - dofs1
		rp := t.Rp(rpi1)
		if !rp.IsStraight() {
			dist *= (rp.CurveRadius(p.Cofs) / rp.CurveRadius(0))
		}
		return dist
	}

	rpCount := int(rpi2 - rpi1)
	if rpi2 <= rpi1 {
		rpCount += t.NumRp()
	}
	dist := phys.Meters(0)
	for i := 0; i < rpCount; i++ {
		j := i + int(rpi1)
		if j >= t.NumRp() {
			j -= t.NumRp()
		}
		rpi := Rpi(j)
		rp := t.Rp(rpi)
		if rpi == rpi1 {
			dist += rp.Len(p.Cofs) * ((rp.CenLen() - rpDofs1) / rp.CenLen())
		} else {
			dist += rp.Len(p.Cofs)
		}
	}
	rp2 := t.Rp(rpi2)
	dist += rp2.Len(p.Cofs) * (rpDofs2 / rp2.CenLen())
	return dist
}

// DriveDeltaDist calculates the driving distance needed to reach dofs starting
// from p, driving along center offset p.Cofs. However, if this distance is more
// than half of the track length at p.Cofs, it is considered BEHIND Pose p, and
// a negative distance is returned.
func (t *Track) DriveDeltaDist(p Pose, dofs phys.Meters) phys.Meters {
	len := t.Len(p.Cofs)
	dd := t.DriveDist(p, dofs)
	if dd > (len / 2) {
		dd -= len
	}
	return dd
}

// DriveDeltaDofs returns the Dofs delta from Track Pose A to dofs, as if Pose
// were the origin.
//   - Bounds: (-t.CenLen()/2) <= DeltaDriveDofs <= (t.CenLen()/2)
//   - If dofs is is more than half of the track length ahead in the direction
//     the Pose is facing, then it is considered BEHIND the Pose, so that
//     DriveDeltaDofs<0.
func (t *Track) DriveDeltaDofs(a Pose, dofs phys.Meters) phys.Meters {
	driveDofs := t.DriveDofsDist(a, dofs)
	if driveDofs > (t.CenLen() / 2) {
		return driveDofs - t.CenLen()
	}
	return driveDofs
}

// DriveDeltaCofs returns the Cofs distance from a Track Pose A to cofs, as if
// Pose were the origin. For example, if Pose is facing counter-trackwise and
// Point is to its left, then DriveDeltaCofs>0.
func (t *Track) DriveDeltaCofs(a Pose, cofs phys.Meters) phys.Meters {
	deltaCofs := cofs - a.Cofs
	if t.isFacingTrackwise(a.DAngle) {
		return deltaCofs
	}
	return -deltaCofs
}
