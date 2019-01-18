// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package phys provides the foundation for modeling the physical world. For
// example, it defines the units of length and angle, and coordinate types.
package phys

import (
	"math"
)

//////////////////////////////////////////////////////////////////////
/// Time
//////////////////////////////////////////////////////////////////////

// SimTime is the measurement of simulation time. It starts at 0 and increases
// with every Tick.
type SimTime uint64 // nanoseconds

const (
	SimNanosecond  = 1
	SimMicrosecond = 1e3
	SimMillisecond = 1e6
	SimSecond      = 1e9
)

//////////////////////////////////////////////////////////////////////
/// Physical units
//////////////////////////////////////////////////////////////////////

// Meters is a unit of distance
type Meters float64

// MetersPerSec is a unit of speed and velocity
type MetersPerSec float64

// MetersPerSec2 is a unit of acceleration
type MetersPerSec2 float64

// Radians is a unit of angle. 360 degrees = 2*pi Radians.
type Radians float64

// Grams is a unit of mass
type Grams float32

//////////////////////////////////////////////////////////////////////
// Constants
//////////////////////////////////////////////////////////////////////

const (
	Radians90DegreeTurnR Radians = -(math.Pi / 2)
	Radians90DegreeTurnL Radians = +(math.Pi / 2)
)

//////////////////////////////////////////////////////////////////////
/// Conversions, Comparisons, etc
//////////////////////////////////////////////////////////////////////

func isNear(v1, v2, tolerance float64) bool {
	return math.Abs(v1-v2) <= math.Abs(tolerance)
}

// MetersAreNear returns true if two Meters values are near each other, within a
// specified tolerance.
func MetersAreNear(m1, m2, tolerance Meters) bool {
	return isNear(float64(m1), float64(m2), float64(tolerance))
}

// MetersPerSecAreNear returns true if two MetersPerSec values are near each
// other, within a specified tolerance.
func MetersPerSecAreNear(m1, m2, tolerance MetersPerSec) bool {
	return isNear(float64(m1), float64(m2), float64(tolerance))
}

// RadiansAreNear returns true if two Radians values are near each other, within
// a specified tolerance.
func RadiansAreNear(a1, a2, tolerance Radians) bool {
	return isNear(float64(a1), float64(a2), float64(tolerance))
}

// NormalizeRadians adjusts a Radians value to be in the range [-Pi, +Pi].
func NormalizeRadians(a Radians) Radians {
	for ; a < (-math.Pi); a += (2 * math.Pi) {
	}
	for ; a > (+math.Pi); a -= (2 * math.Pi) {
	}
	return a
}
