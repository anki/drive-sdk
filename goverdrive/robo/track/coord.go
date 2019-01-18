// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package track

import (
	"github.com/anki/goverdrive/phys"
)

const (
	// TrackDistTol is the general tolerance for checking if two Meters values are
	// equal, in a track point context.
	TrackMetersAreEqualTol  phys.Meters  = 1e-6
	TrackRadiansAreEqualTol phys.Radians = 1e-3
)

// Point is a position in the track coordinate space.
//   - Is absolute, ie it is independent of driving direction
//   - The distance offset is cyclical, like the angle for Polar Coordinates
//   - Is non-Euclidean, because points "bend" to match the shape of the track
//   - Dofs = distance offset = trackwise distance from finish line, along road center
//   - Cofs = center offset, ie from road center
//   - In trackwise driving direction, Cofs>0 means left-of-road-center
//   - NOTE: "trackwise" is like "clockwise", and means natural forward direction,
//     as determined by the start piece (for modular tracks). Counter-trackwise
//     means driving opposite of the natural forward direction of the track.
//
// Conversions:
//   - track.Point -> phys.Point  = straightforward
//   - phys.Point  -> track.Point = not as straightforward
type Point struct {
	Dofs phys.Meters // vert offset
	Cofs phys.Meters // horz offset
}

// Pose adds an orientation (front of vehicle) to a track position
//   0   = facing   the track's forward direction AT THAT PARTICULAR Point
//   pi  = opposite the track's forward direction AT THAT PARTICULAR Point
//   >0  = left (ccw rotation)
//   <0  = right (cw rotation)
// More generally:
//   ABS(DAngle) <= (pi/2)         ==> facing trackwise
//   (pi/2) < ABS(DAngle) <= (pi)  ==> facing counter-trackwise
//   ABS(DAngle) > pi              ==> undefined (bad things may happen)
type Pose struct {
	Point
	DAngle phys.Radians
}

// Vel is the velocity vector of an object, in track coordinate system. D<0
// means vehicle is moving counter-trackwise.
type Vel struct {
	D phys.MetersPerSec // velocity of distance offset (from finish line)
	C phys.MetersPerSec // velocity of center offset (from road center)
}
