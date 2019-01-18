// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package track

import (
	"fmt"
	"math"

	"github.com/anki/goverdrive/phys"
)

// RoadPiece is the helper type used to define a track
//  - Straight or curved only; no fancy shapes like intersection or 'Y'
//  - Any anglular change through the piece is along a circular arc
//  - Maximum of angular change of +/- pi/2 radians (ie 90 degrees)
//  - Width is a parameter of the track, not individual road pieces
type RoadPiece struct {
	cenLen phys.Meters  // path length, at road center
	dAngle phys.Radians // delta angle when driving through the piece (0=>straight; +pi/2=>left turn; etc)
}

func NewRoadPiece(cenLen phys.Meters, dAngle phys.Radians) *RoadPiece {
	if cenLen <= 0 {
		panic(fmt.Sprintf("RoadPiece requires cenLen > 0; actual value is %v", cenLen))
	}
	if dAngle > phys.Radians90DegreeTurnL || dAngle < phys.Radians90DegreeTurnR {
		panic(fmt.Sprintf("RoadPiece requires (%v <= DAngle <= %v); actual value is %v", phys.Radians90DegreeTurnR, phys.Radians90DegreeTurnL, dAngle))
	}
	return &RoadPiece{cenLen: cenLen, dAngle: dAngle}
}

func (rp *RoadPiece) String() string {
	return fmt.Sprintf("RoadPice{cenLen: %v, dAngle: %v}", rp.cenLen, rp.dAngle)
}

func (rp *RoadPiece) CenLen() phys.Meters {
	return rp.cenLen
}

func (rp *RoadPiece) DAngle() phys.Radians {
	return rp.dAngle
}

func (rp *RoadPiece) IsStraight() bool {
	return rp.dAngle == 0
}

// Len computes the path length of the road piece, at the specified center
// offset.
func (rp *RoadPiece) Len(cofs phys.Meters) phys.Meters {
	d := rp.cenLen
	d -= cofs * phys.Meters(rp.dAngle)
	return d
}

// CurveRadius computes the radius of the road piece, at the specified center
// offset. Straight pieces, which have no curvature, will return 0.
func (rp *RoadPiece) CurveRadius(cofs phys.Meters) phys.Meters {
	if rp.IsStraight() {
		return 0
	}
	r := rp.cenLen / phys.Meters(math.Abs(float64(rp.dAngle)))
	if rp.dAngle < 0 {
		// curve right
		r += cofs
	} else {
		// curve left
		r -= cofs
	}
	return r
}

// DeltaPose returns the change in pose when travelling through the piece,
// assuming the canonical starting pose: origin, facing right.
func (rp *RoadPiece) DeltaPose() phys.Pose {
	if rp.IsStraight() {
		return phys.Pose{Point: phys.Point{X: rp.cenLen, Y: 0}, Theta: 0}
	}

	rc := rp.CurveRadius(0)
	angle := phys.Radians(rp.cenLen / rc)
	if rp.dAngle > 0 {
		// left turn
		pp := phys.PolarPoint{R: rc, A: phys.Radians(-(math.Pi / 2) + angle)}
		p := pp.ToPoint()
		p.Y += rc
		return phys.Pose{Point: p, Theta: rp.dAngle}
	}
	// else: right turn
	pp := phys.PolarPoint{R: rc, A: phys.Radians(+(math.Pi / 2) - angle)}
	p := pp.ToPoint()
	p.Y -= rc
	return phys.Pose{Point: p, Theta: rp.dAngle}
}
