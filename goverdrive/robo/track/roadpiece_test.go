// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package track

import (
	"math"
	"testing"

	"github.com/anki/goverdrive/phys"
)

//////////////////////////////////////////////////////////////////////

// TestStraights creates straight road pieces and checks their properties
func TestStraights(t *testing.T) {
	straightCenLens := []phys.Meters{
		0.001,
		0.020,
		0.300,
		4.000,
	}
	for _, cenLen := range straightCenLens {
		rp := NewRoadPiece(cenLen, phys.Radians(0))

		if rp.CenLen() != cenLen {
			t.Errorf("rp.CenLen()=%v is wrong. Exp=%v", rp.CenLen(), cenLen)
		}
		if rp.DAngle() != 0 {
			t.Errorf("rp.DAngle()=%v is wrong. Exp=%v", rp.DAngle(), 0)
		}
		if !rp.IsStraight() {
			t.Errorf("rp.IsStraight()=%v is wrong. Exp=true", rp.IsStraight())
		}

		expDp := phys.Pose{Point: phys.Point{X: cenLen, Y: 0}, Theta: 0}
		gotDp := rp.DeltaPose()
		if gotDp != expDp {
			t.Errorf("rp.DeltaPose()=%v is wrong for rp=%v. Exp=%v", gotDp, rp, expDp)
		}

		// all hofs values should have same length and radius of curvature
		var hofs phys.Meters
		for hofs = -1.1; hofs <= 1.2; hofs += 0.1 {
			len := rp.Len(hofs)
			if len != cenLen {
				t.Errorf("rp.Len(%v)=%v is wrong. Exp=%v", hofs, len, cenLen)
			}
			cr := rp.CurveRadius(hofs)
			if cr != 0 {
				t.Errorf("rp.CurveRadius()=%v is wrong. Exp=0", cr)
			}
		}
	}
}

// TestRightAngleCurves creates 90-degree road pieces and checks their
// properties
func TestRightAngleCurves(t *testing.T) {
	for i := 0; i < 10; i++ { // sweep curve radius values
		radius := phys.Meters(float64(i+1) * 0.1) // curve radius (at road center)
		cenLen := phys.Meters(math.Pi/2) * radius
		for j := 0; j < 2; j++ {
			// 0 => right turn; 1 => left turn
			dAngle := phys.Radians(-math.Pi/2 + float64(j)*math.Pi)
			rp := NewRoadPiece(cenLen, dAngle)

			if rp.CenLen() != cenLen {
				t.Errorf("rp.CenLen()=%v is wrong for rp=%v. Exp=%v", rp.CenLen(), rp, cenLen)
			}
			if rp.DAngle() != dAngle {
				t.Errorf("rp.DAngle()=%v is wrong for rp=%v. Exp=%v", rp.DAngle(), rp, dAngle)
			}
			if rp.IsStraight() {
				t.Errorf("rp.IsStraight()=%v is wrong for rp=%v. Exp=false", rp.IsStraight(), rp)
			}

			expDp := phys.Pose{Point: phys.Point{X: radius, Y: radius}, Theta: rp.DAngle()}
			if rp.DAngle() < 0 { // right turn
				expDp.Y = -radius
			}
			gotDp := rp.DeltaPose()
			if !phys.RadiansAreNear(gotDp.Theta, expDp.Theta, nearRTolerance) ||
				!phys.MetersAreNear(gotDp.X, expDp.X, nearMTolerance) ||
				!phys.MetersAreNear(gotDp.Y, expDp.Y, nearMTolerance) {
				t.Errorf("rp.DeltaPose()=%v is wrong for rp=%v. Exp=%v", gotDp, rp, expDp)
			}

			// radius of curvature and path len depend on the horizontal offset
			hofsList := []phys.Meters{
				+0.00, // road center
				+0.07, // left-of-center
				-0.05, // right-of-center
				+radius,
				-radius,
			}
			for _, hofs := range hofsList {
				hofsRadius := radius + hofs // default: right turn
				if rp.DAngle() > 0 {
					// left turn => larger (+hofs) means smaller curve radius
					hofsRadius = radius - hofs
				}
				expLen := phys.Meters(math.Pi/2) * hofsRadius
				gotLen := rp.Len(hofs)
				if !phys.MetersAreNear(gotLen, expLen, nearMTolerance) {
					t.Errorf("rp.Len(%v)=%v is wrong for rp=%v. Exp=%v", hofs, gotLen, rp, expLen)
				}
				gotCr := rp.CurveRadius(hofs)
				if !phys.MetersAreNear(gotCr, hofsRadius, nearMTolerance) {
					t.Errorf("r.CurveRadius(%v)=%v is wrong for rp=%v. Exp=%v", hofs, gotCr, rp, hofsRadius)
				}
			}
		}
	}
}

// TestOtherCurves creates curved road pieces that are NOT 90 degrees turns, and
// checks their properties.
func TestOtherCurves(t *testing.T) {
	rpList := make([]RoadPiece, 4, 4)
	rpList[0] = *NewRoadPiece(phys.Meters(1.0), phys.Radians(+2.0*math.Pi/6)) // L 60 Degrees
	rpList[1] = *NewRoadPiece(phys.Meters(0.5), phys.Radians(+1.0*math.Pi/6)) // L 30 Degrees
	rpList[2] = *NewRoadPiece(phys.Meters(0.4), phys.Radians(-1.0*math.Pi/6)) // R 30 Degrees
	rpList[3] = *NewRoadPiece(phys.Meters(0.8), phys.Radians(-2.0*math.Pi/6)) // R 60 Degrees

	for _, rp := range rpList {
		if rp.IsStraight() {
			t.Errorf("rp.IsStraight()=%v is wrong. Exp=false", rp.IsStraight())
		}
		expCr := rp.CenLen() / phys.Meters(math.Abs(float64(rp.DAngle())))
		gotCr := rp.CurveRadius(0)
		if !phys.MetersAreNear(gotCr, expCr, nearMTolerance) {
			t.Errorf("rp.CurveRadius(0)=%v is wrong for rp=%v. Exp=%v", gotCr, rp, expCr)
		}
	}

	for i := 0; i < 2; i++ {
		rp1 := rpList[2*i]
		rp2 := rpList[2*i+1]
		cr1 := rp1.CurveRadius(0)
		cr2 := rp2.CurveRadius(0)
		if !phys.MetersAreNear(cr1, cr2, nearMTolerance) {
			t.Errorf("rp1.CurveRadius(0)=%v does not equal rp2.CurveRadius(0)=%v", cr1, cr2)
		}

		expDeltaPose := phys.Pose{Point: phys.Point{X: cr1, Y: cr1}, Theta: phys.Radians(math.Pi / 2)}
		if rp1.DAngle() < 0 {
			expDeltaPose.Y = -expDeltaPose.Y
			expDeltaPose.Theta = -expDeltaPose.Theta
		}
		gotDeltaPose := rp1.DeltaPose().AdvancePose(rp2.DeltaPose())
		if !phys.MetersAreNear(gotDeltaPose.X, expDeltaPose.X, nearMTolerance) ||
			!phys.MetersAreNear(gotDeltaPose.Y, expDeltaPose.Y, nearMTolerance) ||
			!phys.RadiansAreNear(gotDeltaPose.Theta, expDeltaPose.Theta, nearRTolerance) {
			t.Errorf("rp.DeltaPose()=%v is wrong for rp=%v + rp=%v. Exp=%v", gotDeltaPose, rp1, rp2, expDeltaPose)
		}

	}
}
