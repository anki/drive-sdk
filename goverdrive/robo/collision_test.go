// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package robo

import (
	"fmt"
	"math"
	"testing"

	"github.com/anki/goverdrive/phys"
)

//////////////////////////////////////////////////////////////////////

const (
	nearMTolerance phys.Meters  = 1.0e-6 // XXX
	nearRTolerance phys.Radians = 1.0e-6 // XXX
)

// testEqual reports a testing error if the two values are not equal
func testEqual(t *testing.T, tag string, exp interface{}, got interface{}) {
	if exp != got {
		t.Errorf("%s error: exp=%v, got=%v", tag, exp, got)
	}
}

// testMetersAreNear reports a testing error if the two Meters values are not
// near each other
func testMetersAreNear(t *testing.T, tag string, exp phys.Meters, got phys.Meters) {
	if !phys.MetersAreNear(exp, got, nearMTolerance) {
		t.Errorf("%s error: exp=%v, got=%v", tag, exp, got)
	}
}

// testRadiansAreNear reports a testing error if the two Radians values are not
// near each other
func testRadiansAreNear(t *testing.T, tag string, exp phys.Radians, got phys.Radians) {
	if !phys.RadiansAreNear(exp, got, nearRTolerance) {
		t.Errorf("%s error: exp=%v, got=%v", tag, exp, got)
	}
}

//////////////////////////////////////////////////////////////////////

const (
	// WARNING: Some of the test vectors are hard-coded based on the vehicle
	// length and width values. Changing these may break the unit test.
	veh0Len = 0.240
	veh0Wid = 0.044
	veh1Len = 0.080
	veh1Wid = 0.040

	// XXX: These intermediate constants help write test tables
	deltaDist = 0.02
	ddd2      = deltaDist / 2
	fals      = false // XXX: same width as "true", for uniform-width table text
)

// XXX(gwenz): Use a flat struct layout so test vectors can be written concisely
type poiTestVec struct {
	x1          phys.Meters
	y1          phys.Meters
	t1          phys.Radians
	isCollision bool
	poiX        phys.Meters
	poiY        phys.Meters
}

func (v poiTestVec) String() string {
	return fmt.Sprintf("x1=%v y1=%v t1=%v isCollision=%v, poiX=%v, poiY=%v",
		v.x1, v.y1, v.t1, v.isCollision, v.poiX, v.poiY)
}

// calculation helpers, for conscise table entries
//   h = half
//   w = width
//   l = length
//   d = diagonal (45 degree)
//   p = plus
//   m = minus
func hwp() phys.Meters {
	return ((veh0Wid + veh1Wid) / 2) + deltaDist
}
func hwm() phys.Meters {
	return ((veh0Wid + veh1Wid) / 2) - deltaDist
}
func hlp() phys.Meters {
	return ((veh0Len + veh1Len) / 2) + deltaDist
}
func hlm() phys.Meters {
	return ((veh0Len + veh1Len) / 2) - deltaDist
}
func hw0m() phys.Meters {
	return (veh0Wid / 2) - deltaDist
}
func hl0m() phys.Meters {
	return (veh0Len / 2) - deltaDist
}
func hl1d() phys.Meters {
	return phys.Meters(float64(veh1Len/2-deltaDist) * math.Sqrt(2))
}
func hw1d() phys.Meters {
	return phys.Meters(float64(veh1Wid/2-deltaDist) * math.Sqrt(2))
}

// TestCollisionCalcPointOfImpact tests the calcPointsOfImpact() function, which
// is a helper for collision detection.
func TestCollisionCalcPointOfImpact(t *testing.T) {
	// This function has a lot of loops to get good coverage without writing a ton
	// of test vectors.
	//   - testTable is the "base" conditions to check
	//   - Vehicle 0 is the bigger vehicle and is based at the origin with Theta=0
	//   - One set of inner loops translate all coordiantes to different places,
	//     trying to get good coverage of vehicles in a mix of Cartesian quadrants
	//   - Another set of inner loops tries permutations of reversing the direction
	//     the vehicle is facing. (A 180 degree pose change should have same four
	//     vehicle recangle corners, and hence the same collision point.)
	testTable := []poiTestVec{
		// Test two corners of Vehicle 0 are inside Vehicle 1
		//         x1      y1      t1   clsn  poiX     poiY
		poiTestVec{+hlp(), 0.0000, 0.0, fals, +0.0000, 0.00000},
		poiTestVec{+hlm(), 0.0000, 0.0, true, +hl0m(), 0.00000},
		poiTestVec{-hlp(), 0.0000, 0.0, fals, +0.0000, 0.00000},
		poiTestVec{-hlm(), 0.0000, 0.0, true, -hl0m(), 0.00000},
		poiTestVec{0.0000, +hwp(), 0.0, fals, +0.0000, 0.00000},
		poiTestVec{0.0000, +hwm(), 0.0, true, +0.0000, +hw0m()},
		poiTestVec{0.0000, -hwp(), 0.0, fals, +0.0000, 0.00000},
		poiTestVec{0.0000, -hwm(), 0.0, true, +0.0000, -hw0m()},

		// Test one corners of Vehicle 0 is inside Vehicle 1, and vice versa
		//         x1      y1      t1   clsn  poiX            poiY
		poiTestVec{+hlp(), +hwm(), 0.0, fals, +0.0000 + 0.00, 0.00000 + 0.00},
		poiTestVec{+hlm(), +hwm(), 0.0, true, +hl0m() + ddd2, +hw0m() + ddd2},
		poiTestVec{-hlp(), +hwm(), 0.0, fals, +0.0000 + 0.00, 0.00000 + 0.00},
		poiTestVec{-hlm(), +hwm(), 0.0, true, -hl0m() - ddd2, +hw0m() + ddd2},
		poiTestVec{+hlp(), -hwm(), 0.0, fals, +0.0000 + 0.00, 0.00000 + 0.00},
		poiTestVec{+hlm(), -hwm(), 0.0, true, +hl0m() + ddd2, -hw0m() - ddd2},
		poiTestVec{-hlp(), -hwm(), 0.0, fals, +0.0000 + 0.00, 0.00000 + 0.00},
		poiTestVec{-hlm(), -hwm(), 0.0, true, -hl0m() - ddd2, -hw0m() - ddd2},

		// Test Vehicle 1 is completely contained in Vehicle 0
		//         x1      y1      t1   clsn  poiX            poiY
		poiTestVec{0.0000, 0.0000, 0.0, true, +0.0000 + 0.00, 0.00000 + 0.00},
		poiTestVec{0.0000, 0.0000, 0.2, true, +0.0000 + 0.00, 0.00000 + 0.00},

		// Test Vehicle rotated 45 degrees, positioned such that exactly one corner
		// of Vehicle 0 is inside Vehicle 1 rectangle. (This makes checking the
		// collision point easier.)
		// XXX(gwenz): This is brittle, and took some effort to get working values.
		//         x1                   y1                               t1               clsn  poiX          poiY
		poiTestVec{+veh0Len/2 + hl1d(), +veh0Wid/2 + veh1Wid/2 + hw1d(), 1 * math.Pi / 4, true, +veh0Len / 2, +veh0Wid / 2},
		poiTestVec{-veh0Len/2 - hl1d(), +veh0Wid/2 + veh1Wid/2 + hw1d(), 3 * math.Pi / 4, true, -veh0Len / 2, +veh0Wid / 2},
		poiTestVec{+veh0Len/2 + hl1d(), -veh0Wid/2 - veh1Wid/2 - hw1d(), 3 * math.Pi / 4, true, +veh0Len / 2, -veh0Wid / 2},
		poiTestVec{+veh0Len/2 + hl1d(), -veh0Wid/2 - veh1Wid/2 - hw1d(), 1 * math.Pi / 4, fals, +veh0Len / 2, -veh0Wid / 2},
		poiTestVec{-veh0Len/2 - hl1d(), -veh0Wid/2 - veh1Wid/2 - hw1d(), 1 * math.Pi / 4, true, -veh0Len / 2, -veh0Wid / 2},
		poiTestVec{-veh0Len/2 - hl1d(), -veh0Wid/2 - veh1Wid/2 - hw1d(), 3 * math.Pi / 4, fals, -veh0Len / 2, -veh0Wid / 2},
	}

	for i, vec := range testTable {
		vecStr := fmt.Sprintf("Vec %d: poiTestVec=%s", i, vec.String())

		for _, rad := range []phys.Meters{0, 0.01, 0.1, 1.0} {
			for rho := float64(0); rho < (2 * math.Pi); rho += (math.Pi / 4) {
				xlatePoint := phys.PolarPoint{R: rad, A: phys.Radians(rho)}.ToPoint()
				// translate both vehicle poses by PolarPoint{rad, rho}
				vehPose := [2]phys.Pose{
					phys.Pose{Point: phys.Point{X: 0.0000, Y: 0.0000}, Theta: 0.0000},
					phys.Pose{Point: phys.Point{X: vec.x1, Y: vec.y1}, Theta: vec.t1},
				}
				for j := range vehPose {
					vehPose[j].X += xlatePoint.X
					vehPose[j].Y += xlatePoint.Y
				}
				// translate expected point-of-impact by PolarPoint{rad, rho}
				expPoiX := vec.poiX + xlatePoint.X
				expPoiY := vec.poiY + xlatePoint.Y

				// rotate each vehicle by +/- 180 degrees => should not affect the result
				for dt0 := -1; dt0 < 2; dt0++ {
					for dt1 := -1; dt1 < 2; dt1++ {
						// test vector -> function input type
						vehPose[0].Theta = phys.NormalizeRadians(phys.Radians(math.Pi*float64(dt0)) + vehPose[0].Theta)
						vehPose[1].Theta = phys.NormalizeRadians(phys.Radians(math.Pi*float64(dt1)) + vehPose[1].Theta)
						inputs := [2]vehCollisionInputs{
							vehCollisionInputs{dofs: 0, pose: vehPose[0], len: veh0Len, width: veh0Wid},
							vehCollisionInputs{dofs: 0, pose: vehPose[1], len: veh1Len, width: veh1Wid},
						} // Note:      Dofs ^^^^ is unused

						isCollision, poi := calcPointOfImpact(inputs)
						testEqual(t, fmt.Sprintf("%s isCollision", vecStr), vec.isCollision, isCollision)
						if isCollision {
							testMetersAreNear(t, fmt.Sprintf("%s PointOfImpact.X", vecStr), expPoiX, poi.X)
							testMetersAreNear(t, fmt.Sprintf("%s PointOfImpact.Y", vecStr), expPoiY, poi.Y)
						}
					}
				}
			}
		}
	}
}
