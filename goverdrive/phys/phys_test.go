// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package phys

import (
	"math"
	"testing"
)

const (
	mTol Meters  = 1.0e-6 // XXX: Need more than micrometer-level precision?
	rTol Radians = 1.0e-6 // XXX: not sure how much precision is needed for goverdrive
)

//////////////////////////////////////////////////////////////////////

type nearTestVec struct {
	m1  Meters
	m2  Meters
	tol Meters
	exp bool // expected MetersAreNear() result
}

func TestMetersAreNear(t *testing.T) {
	testTable := []nearTestVec{
		nearTestVec{m1: +0.00, m2: +0.10, tol: 0.05, exp: false},
		nearTestVec{m1: +0.00, m2: +0.10, tol: 0.10, exp: true},
		nearTestVec{m1: +0.00, m2: +0.10, tol: 0.20, exp: true},
		nearTestVec{m1: -0.00, m2: +0.00, tol: 0.00, exp: true},
		nearTestVec{m1: -0.04, m2: +0.05, tol: 0.10, exp: true},
		nearTestVec{m1: +0.04, m2: -0.05, tol: 0.10, exp: true},
		nearTestVec{m1: -0.05, m2: +0.05, tol: 0.10, exp: true},
		nearTestVec{m1: -0.06, m2: +0.06, tol: 0.10, exp: false},
		nearTestVec{m1: +0.10, m2: +0.19, tol: 0.10, exp: true},
		nearTestVec{m1: +0.10, m2: +0.20, tol: 0.10, exp: true},
		nearTestVec{m1: +0.10, m2: +0.21, tol: 0.10, exp: false},
		nearTestVec{m1: -0.10, m2: -0.19, tol: 0.10, exp: true},
		nearTestVec{m1: -0.10, m2: -0.20, tol: 0.10, exp: true},
		nearTestVec{m1: -0.10, m2: -0.21, tol: 0.10, exp: false},
		nearTestVec{m1: +0.0000001, m2: +0.00000011, tol: 0.0000000110, exp: true},
		nearTestVec{m1: +0.0000001, m2: +0.00000011, tol: 0.0000000101, exp: true},
		nearTestVec{m1: +0.0000001, m2: +0.00000011, tol: 0.0000000090, exp: false},
	}

	for i, vec := range testTable {
		got := MetersAreNear(vec.m1, vec.m2, vec.tol)
		if got != vec.exp {
			t.Errorf("Vec=%d MetersAreNear(%v, %v, %v) mismatch; exp=%v, got=%v", i, vec.m1, vec.m2, vec.tol, vec.exp, got)
		}
		// swap order of func args
		got2 := MetersAreNear(vec.m2, vec.m1, vec.tol)
		if got2 != vec.exp {
			t.Errorf("Vec=%d MetersAreNear(%v, %v, %v) mismatch; exp=%v, got=%v", i, vec.m2, vec.m1, vec.tol, vec.exp, got2)
		}
	}
}

//////////////////////////////////////////////////////////////////////

type distTestVec struct {
	p1  Point
	p2  Point
	exp Meters
}

func TestDist(t *testing.T) {
	testTable := []distTestVec{
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: +0.00, Y: +0.00}, exp: 0.00},
		distTestVec{p1: Point{X: -0.00, Y: -0.00}, p2: Point{X: +0.00, Y: +0.00}, exp: 0.00},
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: +0.10, Y: +0.00}, exp: 0.10}, // on-axis (90 degrees), from origin
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: +0.00, Y: +0.10}, exp: 0.10},
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: -0.10, Y: +0.00}, exp: 0.10},
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: +0.00, Y: -0.10}, exp: 0.10},
		distTestVec{p1: Point{X: +0.10, Y: +0.40}, p2: Point{X: +0.20, Y: +0.40}, exp: 0.10}, // on-axis (90 degrees), not from origin
		distTestVec{p1: Point{X: +0.20, Y: +0.30}, p2: Point{X: +0.10, Y: +0.30}, exp: 0.10},
		distTestVec{p1: Point{X: +0.30, Y: +0.20}, p2: Point{X: +0.30, Y: +0.10}, exp: 0.10},
		distTestVec{p1: Point{X: +0.40, Y: +0.10}, p2: Point{X: +0.40, Y: +0.20}, exp: 0.10},
		distTestVec{p1: Point{X: +0.10, Y: -0.40}, p2: Point{X: +0.20, Y: -0.40}, exp: 0.10}, // on-axis (90 degrees), not from origin
		distTestVec{p1: Point{X: +0.20, Y: -0.30}, p2: Point{X: +0.10, Y: -0.30}, exp: 0.10},
		distTestVec{p1: Point{X: +0.30, Y: -0.20}, p2: Point{X: +0.30, Y: -0.10}, exp: 0.10},
		distTestVec{p1: Point{X: +0.40, Y: -0.10}, p2: Point{X: +0.40, Y: -0.20}, exp: 0.10},
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: +1.00, Y: +1.00}, exp: Meters(math.Sqrt(2.0))}, // 45 degress, from origin
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: -1.00, Y: +1.00}, exp: Meters(math.Sqrt(2.0))},
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: +1.00, Y: -1.00}, exp: Meters(math.Sqrt(2.0))},
		distTestVec{p1: Point{X: +0.00, Y: +0.00}, p2: Point{X: -1.00, Y: -1.00}, exp: Meters(math.Sqrt(2.0))},
	}
	for i, vec := range testTable {
		got := Dist(vec.p1, vec.p2)
		if !MetersAreNear(got, vec.exp, mTol) {
			t.Errorf("Vec=%d Dist(%v, %v) mismatch; exp=%v, got=%v", i, vec.p1, vec.p2, vec.exp, got)
		}
		// swap order of func args
		got2 := Dist(vec.p2, vec.p1)
		if !MetersAreNear(got2, vec.exp, mTol) {
			t.Errorf("Vec=%d Dist(%v, %v) mismatch; exp=%v, got=%v", i, vec.p2, vec.p1, vec.exp, got2)
		}
	}
}

//////////////////////////////////////////////////////////////////////

type p2ppTestVec struct {
	p  Point
	pp PolarPoint
}

// TestPointConv tests conversions between Point and PolarPoint
func TestPointConv(t *testing.T) {
	const pi Radians = Radians(math.Pi) // for brevity

	/// Spot-check edge cases, different quadrants, etc
	// Point -> PolarPoint
	testTable1 := []p2ppTestVec{
		p2ppTestVec{p: Point{X: +00.000000, Y: +00.000000}, pp: PolarPoint{R: +00.000000, A: +0.00 * pi}},
		p2ppTestVec{p: Point{X: +10.000001, Y: +00.000000}, pp: PolarPoint{R: +10.000001, A: +0.00 * pi}},
		p2ppTestVec{p: Point{X: -10.000001, Y: +00.000000}, pp: PolarPoint{R: +10.000001, A: +1.00 * pi}},
		p2ppTestVec{p: Point{X: +00.000000, Y: +10.000001}, pp: PolarPoint{R: +10.000001, A: +0.50 * pi}},
		p2ppTestVec{p: Point{X: -00.000000, Y: -10.000001}, pp: PolarPoint{R: +10.000001, A: -0.50 * pi}},
		p2ppTestVec{p: Point{X: -05.000000, Y: +05.000000}, pp: PolarPoint{R: Meters(5 * math.Sqrt(2.0)), A: pi * 0.75}},
	}

	for i, vec := range testTable1 {
		pp := vec.p.ToPolarPoint()
		if !MetersAreNear(vec.pp.R, pp.R, mTol) || !RadiansAreNear(vec.pp.A, pp.A, rTol) {
			t.Errorf("Vec=%d ToPolarPoint(%s) mismatch; exp=%s, got=%s", i, vec.p.String(), vec.pp.String(), pp.String())
		}
	}

	// PolarPoint -> Point
	testTable2 := []p2ppTestVec{
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +0.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +1.00 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +1.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +2.00 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: -2.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: -0.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: -1.00 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: -1.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: -2.00 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: -2.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +0.50 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +0.00 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +10.000001, A: +0.00 * pi}, p: Point{X: +10.000001, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +00.000000, A: +2.00 * pi}, p: Point{X: +00.000000, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +10.000001, A: +2.00 * pi}, p: Point{X: +10.000001, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +08.000002, A: +1.00 * pi}, p: Point{X: -08.000002, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +05.000005, A: +3.00 * pi}, p: Point{X: -05.000005, Y: +00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +06.000006, A: +2.50 * pi}, p: Point{X: +00.000000, Y: +06.000006}},
		p2ppTestVec{pp: PolarPoint{R: +07.000007, A: +3.50 * pi}, p: Point{X: -00.000000, Y: -07.000007}},
		p2ppTestVec{pp: PolarPoint{R: +08.000008, A: +12.0 * pi}, p: Point{X: +08.000008, Y: -00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +08.000008, A: -11.0 * pi}, p: Point{X: -08.000008, Y: -00.000000}},
		p2ppTestVec{pp: PolarPoint{R: +03.000003, A: -11.5 * pi}, p: Point{X: -00.000000, Y: +03.000003}},
		p2ppTestVec{pp: PolarPoint{R: +02.000002, A: -13.5 * pi}, p: Point{X: -00.000000, Y: +02.000002}},
		p2ppTestVec{pp: PolarPoint{R: +02.000002, A: -14.5 * pi}, p: Point{X: -00.000000, Y: -02.000002}},
	}

	for i, vec := range testTable2 {
		p := vec.pp.ToPoint()
		if !MetersAreNear(vec.p.X, p.X, mTol) || !MetersAreNear(vec.p.Y, p.Y, mTol) {
			t.Errorf("Vec=%d ToPoint(%s) mismatch; exp=%s, got=%s", i, vec.pp.String(), vec.p.String(), p.String())
		}
	}

	/// Sweep: PolarPoint -> Point -> PolarPoint
	for r := 0.1; r < 10.0; r += 0.317 {
		for a := (-4 * math.Pi); a < (4 * math.Pi); a += 0.101 {
			pp1 := PolarPoint{R: Meters(r), A: Radians(a)}
			pp1n := PolarPoint{R: pp1.R, A: NormalizeRadians(pp1.A)}
			p := pp1.ToPoint()
			pp2 := p.ToPolarPoint()
			if !MetersAreNear(pp1n.R, pp2.R, mTol) || !RadiansAreNear(pp1n.A, pp2.A, rTol) {
				t.Errorf("PolarPoint->Point->PolarPoint mismatch for %s; exp=%s, got=%s", pp1.String(), pp1n.String(), pp2.String())
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////

type poseTestVec struct {
	start Pose
	delta Pose
	exp   Pose
}

// makePose is for brevity to specify test tables
func makePose(x, y Meters, theta Radians) Pose {
	return Pose{Point: Point{X: x, Y: y}, Theta: theta}
}

func TestAdvancePose(t *testing.T) {
	const pi Radians = Radians(math.Pi) // for brevity

	testTable := []poseTestVec{
		// advance by {0,0,0} => no change froms start
		poseTestVec{start: makePose(+1.000, +2.000, +0.500*pi), delta: makePose(+0.000, +0.000, +0.000*pi), exp: makePose(+1.000, +2.000, +0.500*pi)},
		poseTestVec{start: makePose(-2.000, +3.000, +0.600*pi), delta: makePose(+0.000, +0.000, +0.000*pi), exp: makePose(-2.000, +3.000, +0.600*pi)},
		poseTestVec{start: makePose(+3.000, -4.000, -0.700*pi), delta: makePose(+0.000, +0.000, +0.000*pi), exp: makePose(+3.000, -4.000, -0.700*pi)},
		poseTestVec{start: makePose(-4.000, -5.000, -0.800*pi), delta: makePose(+0.000, +0.000, +0.000*pi), exp: makePose(-4.000, -5.000, -0.800*pi)},
		// start from origin
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(+0.000, +0.000, +0.000*pi), exp: makePose(+0.000, +0.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+1.000, +0.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(-1.000, +0.000, +0.000*pi), exp: makePose(-1.000, +0.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+0.000, +1.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(+0.000, -1.000, +0.000*pi), exp: makePose(+0.000, -1.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(+0.000, +0.000, +0.900*pi), exp: makePose(+0.000, +0.000, +0.900*pi)},
		poseTestVec{start: makePose(+0.000, +0.000, +0.000*pi), delta: makePose(+0.000, +0.000, -0.900*pi), exp: makePose(+0.000, +0.000, -0.900*pi)},
		// start from non-orgin, theta=0
		poseTestVec{start: makePose(+1.000, +0.000, +0.000*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+2.000, +0.000, +0.000*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.000*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+1.000, +1.000, +0.000*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.000*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+1.000, +0.000, +1.000*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.000*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(+2.000, +2.000, +0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.000*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+1.000, +1.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.000*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+0.000, +2.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.000*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+0.000, +1.000, +1.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.000*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(+1.000, +3.000, +0.500*pi)},
		// start from non-orgin, theta=pi
		poseTestVec{start: makePose(+1.000, +0.000, +1.000*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+0.000, +0.000, +1.000*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +1.000*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+1.000, -1.000, +1.000*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +1.000*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+1.000, +0.000, +0.000*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +1.000*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(+0.000, -2.000, -0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +1.000*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(-1.000, +1.000, +1.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +1.000*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+0.000, +0.000, +1.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +1.000*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+0.000, +1.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +1.000*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(-1.000, -1.000, -0.500*pi)},
		// start from non-orgin, theta=(pi/2)
		poseTestVec{start: makePose(+1.000, +0.000, +0.500*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+1.000, +1.000, +0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.500*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+0.000, +0.000, +0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.500*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+1.000, +0.000, -0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.500*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(-1.000, +1.000, +1.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.500*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+0.000, +2.000, +0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.500*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(-1.000, +1.000, +0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.500*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+0.000, +1.000, -0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, +0.500*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(-2.000, +2.000, +1.000*pi)},
		// start from non-orgin, theta=(-pi/2)
		poseTestVec{start: makePose(+1.000, +0.000, -0.500*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+1.000, -1.000, -0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, -0.500*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+2.000, +0.000, -0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, -0.500*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+1.000, +0.000, +0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, -0.500*pi), delta: makePose(+1.000, +2.000, +0.500*pi), exp: makePose(+3.000, -1.000, +0.000*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, -0.500*pi), delta: makePose(+1.000, +0.000, +0.000*pi), exp: makePose(+0.000, +0.000, -0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, -0.500*pi), delta: makePose(+0.000, +1.000, +0.000*pi), exp: makePose(+1.000, +1.000, -0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, -0.500*pi), delta: makePose(+0.000, +0.000, +1.000*pi), exp: makePose(+0.000, +1.000, +0.500*pi)},
		poseTestVec{start: makePose(+0.000, +1.000, -0.500*pi), delta: makePose(+1.000, -2.000, +0.400*pi), exp: makePose(-2.000, +0.000, -0.100*pi)},
		// theta=(+pi/4) and (-pi/4)
		poseTestVec{start: makePose(+0.000, +0.000, +0.250*pi), delta: makePose(+1.000, +1.000, +0.250*pi), exp: makePose(+0.000, Meters(math.Sqrt(2.0)), +0.500*pi)},
		poseTestVec{start: makePose(+1.000, +0.000, +0.250*pi), delta: makePose(+1.000, +1.000, -0.250*pi), exp: makePose(+1.000, Meters(math.Sqrt(2.0)), +0.000*pi)},
		poseTestVec{start: makePose(+1.000, +1.000, +0.250*pi), delta: makePose(-2.0*Meters(math.Sqrt(2.0)), +0.000, -1.000*pi), exp: makePose(-1.000, -1.000, -0.750*pi)},
		poseTestVec{start: makePose(+1.000, -1.000, -0.250*pi), delta: makePose(-2.0*Meters(math.Sqrt(2.0)), +0.000, +0.500*pi), exp: makePose(-1.000, +1.000, +0.250*pi)},
	}

	for i, vec := range testTable {
		gotPose := vec.start.AdvancePose(vec.delta)
		if !MetersAreNear(gotPose.X, vec.exp.X, mTol) ||
			!MetersAreNear(gotPose.Y, vec.exp.Y, mTol) ||
			!RadiansAreNear(gotPose.Theta, vec.exp.Theta, rTol) {
			t.Errorf("Vec=%d %s.AdvancePose(%s) mismatch; exp=%s, got=%s", i, vec.start.String(), vec.delta.String(), vec.exp.String(), gotPose.String())
		}
	}
}
