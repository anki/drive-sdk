// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package track

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/anki/goverdrive/phys"
)

//////////////////////////////////////////////////////////////////////

const (
	defTrackWidth               = 0.2
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

type testDofsVec struct {
	dofs      phys.Meters
	expRpi    Rpi
	expRpDofs phys.Meters
}

type testPoseVec struct {
	dofs    phys.Meters
	expPose phys.Pose
}

// TestLeftQuadraTracks checks basic track properties like width, length, entry
// poses, and road piece indexing. This is a convenient track for testing, since
// it has easily-computable geometry.
func TestLeftQuadraTracks(t *testing.T) {
	for n := 1; n <= 5; n++ { // number of consecutive straights
		topo := strings.Repeat("S", n) + "L" +
			/******/ strings.Repeat("S", n) + "L" +
			/******/ strings.Repeat("S", n) + "L" +
			/******/ strings.Repeat("S", n) + "L"
		width := phys.Meters(float32(n) * 0.1)
		numRp := len(topo) + 1
		track, err := NewModularTrack(width, width/2, topo)
		if err != nil {
			t.Fatalf(err.Error())
		}

		// Verify basic track properties
		testEqual(t, topo+" track.Width()", width, track.Width())
		testEqual(t, topo+" track.NumRp()", numRp, track.NumRp())

		// Get one of the 90-degree curves in the track
		curveRpi := Rpi(n) // first curve is after n straights
		track.assertValidRpi(curveRpi)
		curveRp := track.Rp(curveRpi)
		testEqual(t, topo+" curveRp.IsStraight()", false, curveRp.IsStraight())

		// Verify overall track length at different center offsets
		for h := 0; h < 3; h++ {
			cofs := phys.Meters(-0.3 + (float64(h) * 0.3))
			radius := curveRp.CurveRadius(cofs)
			curveLen := phys.Meters(math.Pi/2) * radius
			expLen := (phys.Meters(4*n) * TrackLenModStraight) + (phys.Meters(4) * curveLen)
			gotLen := track.Len(cofs)
			testMetersAreNear(t, topo+fmt.Sprintf(" track.Len(%v)", cofs), expLen, gotLen)
			if phys.MetersAreNear(cofs, 0, nearMTolerance) {
				testMetersAreNear(t, topo+" track.CenLen()", expLen, gotLen)
			}
		}

		// Verify overall size of the "world"
		centerRadius := curveRp.CurveRadius(0) // road center
		outerRadius := curveRp.CurveRadius(-width / 2)
		expWorldWidth := (phys.Meters(n) * TrackLenModStraight) + (phys.Meters(2) * outerRadius)
		gotWorldWidth := track.MaxCorner().X - track.MinCorner().X
		testMetersAreNear(t, topo+" WorldWidth", expWorldWidth, gotWorldWidth)
		expWorldHeight := (phys.Meters(n) * TrackLenModStraight) + (phys.Meters(2) * outerRadius)
		gotWorldHeight := track.MaxCorner().X - track.MinCorner().X
		testMetersAreNear(t, topo+" WorldHeight", expWorldHeight, gotWorldHeight)

		// Verify RpiAt() and RpiAndRpDofs() in a few places
		tlen := track.Len(0)
		dofsTestTable := []testDofsVec{
			testDofsVec{dofs: TrackLenModStartShort / 2.0 /********/, expRpi: Rpi(0) /*****/, expRpDofs: TrackLenModStartShort / 2.0},
			testDofsVec{dofs: tlen - TrackLenModStartLong /********/, expRpi: Rpi(numRp - 1), expRpDofs: 0},
			testDofsVec{dofs: tlen - TrackLenModStartLong - 0.1 /**/, expRpi: Rpi(numRp - 2), expRpDofs: TrackLenModCurve - 0.1},
		}
		for _, v := range dofsTestTable {
			rpi, rpDofs := track.RpiAndRpDofs(v.dofs)
			testEqual(t, topo+fmt.Sprintf(" RpiAt(%v)", v.dofs), Rpi(v.expRpi), track.RpiAt(v.dofs))
			testEqual(t, topo+fmt.Sprintf(" RpiAndRpDofs(%v).Rpi", v.dofs), v.expRpi, rpi)
			testMetersAreNear(t, topo+fmt.Sprintf(" RpiAndRpDofs(%v).RpDofs", v.dofs), v.expRpDofs, rpDofs)
		}

		// Verify RpEntryPose() in a few places
		for i := 1; i < (n + 1); i++ {
			// First set of straights extends in +X direction
			rpi := Rpi(i)
			expEntryPose := phys.Pose{Point: phys.Point{
				X: TrackLenModStartShort + (phys.Meters(i-1) * TrackLenModStraight),
				Y: 0},
				Theta: 0}
			gotEntryPose := track.RpEntryPose(rpi)
			testMetersAreNear(t, topo+fmt.Sprintf(" RpEntryPose(%v).X", rpi), expEntryPose.X, gotEntryPose.X)
			testMetersAreNear(t, topo+fmt.Sprintf(" RpEntryPose(%v).Y", rpi), expEntryPose.Y, gotEntryPose.Y)
			testRadiansAreNear(t, topo+fmt.Sprintf(" RpEntryPose(%v).Theta", rpi), expEntryPose.Theta, gotEntryPose.Theta)

			// Second set of straights goes in +Y direction
			rpi = Rpi(n + i)
			expEntryPose = phys.Pose{Point: phys.Point{
				X: centerRadius + (phys.Meters(n-1) * TrackLenModStraight) + TrackLenModStartShort,
				Y: centerRadius + (phys.Meters(i-1) * TrackLenModStraight)},
				Theta: phys.Radians(math.Pi / 2)}
			gotEntryPose = track.RpEntryPose(rpi)
			testMetersAreNear(t, topo+fmt.Sprintf(" RpEntryPose(%v).X", rpi), expEntryPose.X, gotEntryPose.X)
			testMetersAreNear(t, topo+fmt.Sprintf(" RpEntryPose(%v).Y", rpi), expEntryPose.Y, gotEntryPose.Y)
			testRadiansAreNear(t, topo+fmt.Sprintf(" RpEntryPose(%v).Theta", rpi), expEntryPose.Theta, gotEntryPose.Theta)
		}

		// Verify that Dofs <-> Pose is consistent at entry point of each road piece
		for i := 0; i < track.NumRp(); i++ {
			dofs := track.RpEntryDofs(Rpi(i))
			rpi, rpDofs := track.RpiAndRpDofs(dofs)
			testEqual(t, topo+" rpi", Rpi(i), rpi)
			testMetersAreNear(t, topo+" rpDofs", 0, rpDofs)
			gotPose := track.ToPose(Pose{Point: Point{Dofs: dofs, Cofs: 0}, DAngle: 0})
			testMetersAreNear(t, topo+fmt.Sprintf(" ToPose(%v, 0).X", dofs), track.RpEntryPose(rpi).X, gotPose.X)
			testMetersAreNear(t, topo+fmt.Sprintf(" ToPose(%v, 0).Y", dofs), track.RpEntryPose(rpi).Y, gotPose.Y)
			testRadiansAreNear(t, topo+fmt.Sprintf(" ToPose(%v, 0).Theta", dofs), track.RpEntryPose(rpi).Theta, gotPose.Theta)
		}

		// Verify ToPose() in a few more places
		poseTestTable := []testPoseVec{
			testPoseVec{
				dofs:    0.1,
				expPose: phys.Pose{Point: phys.Point{X: 0.1, Y: 0}, Theta: 0}},
			testPoseVec{
				dofs:    0.1 + TrackLenModStartShort + (phys.Meters(n-1) * TrackLenModStraight) + TrackLenModCurve,
				expPose: phys.Pose{Point: phys.Point{X: phys.Meters(n-1)*TrackLenModStraight + TrackLenModStartShort + centerRadius, Y: 0.1 + centerRadius}, Theta: phys.Radians(math.Pi / 2)}},
			testPoseVec{
				dofs:    0.1 + TrackLenModStartShort + (phys.Meters(n-1) * TrackLenModStraight) + (phys.Meters(2*n) * TrackLenModStraight) + phys.Meters(3)*TrackLenModCurve,
				expPose: phys.Pose{Point: phys.Point{X: -TrackLenModStartLong - centerRadius, Y: (phys.Meters(n) * TrackLenModStraight) + centerRadius - 0.1}, Theta: phys.Radians(-math.Pi / 2)}},
		}
		for _, v := range poseTestTable {
			pose := track.ToPose(Pose{Point: Point{Dofs: v.dofs, Cofs: 0}, DAngle: 0})
			testMetersAreNear(t, topo+fmt.Sprintf(" ToPose(%v, 0).X", v.dofs), v.expPose.X, pose.X)
			testMetersAreNear(t, topo+fmt.Sprintf(" ToPose(%v, 0).Y", v.dofs), v.expPose.Y, pose.Y)
			testRadiansAreNear(t, topo+fmt.Sprintf(" ToPose(%v, 0).Theta", v.dofs), v.expPose.Theta, pose.Theta)
		}

		// Verify the center point of all four curves on the track
		curveCenterDist := phys.Meters(n) * TrackLenModStraight
		curveCenters := make([]phys.Point, 4, 4) // quadra corners, in driving order
		curveCenters[3] = phys.Point{            // lower-left
			X: -TrackLenModStartLong,
			Y: centerRadius}
		curveCenters[0] = phys.Point{ // lower-right
			X: curveCenters[3].X + curveCenterDist,
			Y: curveCenters[3].Y}
		curveCenters[1] = phys.Point{ // upper-right
			X: curveCenters[3].X + curveCenterDist,
			Y: curveCenters[3].Y + curveCenterDist}
		curveCenters[2] = phys.Point{ // upper-left
			X: curveCenters[3].X,
			Y: curveCenters[3].Y + curveCenterDist}
		for i := 0; i < 4; i++ {
			rpi := Rpi(n + i*(n+1)) // curve[i]
			testMetersAreNear(t, topo+fmt.Sprintf(" RpCurveCenter(%v).X", rpi), curveCenters[i].X, track.RpCurveCenter(rpi).X)
			testMetersAreNear(t, topo+fmt.Sprintf(" RpCurveCenter(%v).Y", rpi), curveCenters[i].Y, track.RpCurveCenter(rpi).Y)
		}
	}
}

//////////////////////////////////////////////////////////////////////

type testIsFacingTWVec struct {
	dangle phys.Radians
	exp    bool
}

type testDriveDofsDistVec struct {
	frDofs   phys.Meters
	frDangle phys.Radians
	toDofs   phys.Meters
	exp      phys.Meters
}

// TestTrackCoord tests miscellaneous functions related to track coordinates,
// poses, distance measurements, etc.
func TestTrackCoord(t *testing.T) {
	trk, err := NewModularTrack(defTrackWidth, defTrackWidth/2, "SLLSSLLS") // Left Capsule
	if err != nil {
		t.Fatalf(err.Error())
	}
	trkLen := trk.CenLen()

	fmt.Println("Check Track.NormalizeDofs()..")
	for i := -3; i < 3; i++ {
		for _, expNDofs := range []phys.Meters{0.1, 0.5, 1.0} {
			dofs := expNDofs + (phys.Meters(i) * trkLen)
			gotNDofs := trk.NormalizeDofs(dofs)
			testMetersAreNear(t, fmt.Sprintf("NormalizeDofs(%v)", dofs), expNDofs, gotNDofs)
		}
	}

	fmt.Println("Check Track.isFacingTrackwise()..")
	// Track.isFacingTrackwise() requires input angle to be in range [-pi,pi]
	isFacingTWTestTable := []testIsFacingTWVec{
		testIsFacingTWVec{dangle: +0.0 * math.Pi, exp: true},
		testIsFacingTWVec{dangle: +0.1 * math.Pi, exp: true},
		testIsFacingTWVec{dangle: +0.4 * math.Pi, exp: true},
		testIsFacingTWVec{dangle: -0.1 * math.Pi, exp: true},
		testIsFacingTWVec{dangle: -0.4 * math.Pi, exp: true},

		testIsFacingTWVec{dangle: +1.0 * math.Pi, exp: false},
		testIsFacingTWVec{dangle: +0.9 * math.Pi, exp: false},
		testIsFacingTWVec{dangle: +0.6 * math.Pi, exp: false},

		testIsFacingTWVec{dangle: -1.0 * math.Pi, exp: false},
		testIsFacingTWVec{dangle: -0.9 * math.Pi, exp: false},
		testIsFacingTWVec{dangle: -0.6 * math.Pi, exp: false},
	}
	for i, v := range isFacingTWTestTable {
		label := fmt.Sprintf("Vec %d: isFacingTrackwise(%v)", i, v.dangle)
		got := trk.isFacingTrackwise(v.dangle)
		testEqual(t, label, v.exp, got)
	}

	fmt.Println("Check Track.DriveDofsDist()..")
	driveDofsDistTestTable := []testDriveDofsDistVec{
		testDriveDofsDistVec{frDofs: 0.1, frDangle: +0.0 * math.Pi, toDofs: 0.3, exp: 0.2},
		testDriveDofsDistVec{frDofs: 0.1, frDangle: +1.0 * math.Pi, toDofs: 0.3, exp: trkLen - 0.2},
		testDriveDofsDistVec{frDofs: 0.1, frDangle: -1.0 * math.Pi, toDofs: 0.3, exp: trkLen - 0.2},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: +0.0 * math.Pi, toDofs: trkLen - 0.400, exp: 0.0},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: +1.0 * math.Pi, toDofs: trkLen - 0.400, exp: 0.0},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: -1.0 * math.Pi, toDofs: trkLen - 0.400, exp: 0.0},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: +0.0 * math.Pi, toDofs: trkLen - 0.399, exp: 0.001},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: +1.0 * math.Pi, toDofs: trkLen - 0.399, exp: trkLen - 0.001},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: -1.0 * math.Pi, toDofs: trkLen - 0.399, exp: trkLen - 0.001},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: +0.0 * math.Pi, toDofs: trkLen - 0.401, exp: trkLen - 0.001},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: +1.0 * math.Pi, toDofs: trkLen - 0.401, exp: 0.001},
		testDriveDofsDistVec{frDofs: trkLen - 0.4, frDangle: -1.0 * math.Pi, toDofs: trkLen - 0.401, exp: 0.001},
	}
	for i, v := range driveDofsDistTestTable {
		pose := Pose{Point: Point{Dofs: v.frDofs, Cofs: 0}, DAngle: v.frDangle}
		label := fmt.Sprintf("Vec %d: DriveDofsDist(%v, %v)", i, pose, v.toDofs)
		got := trk.DriveDofsDist(pose, v.toDofs)
		testMetersAreNear(t, label, v.exp, got)
	}

}

//////////////////////////////////////////////////////////////////////

type testDofsDistVec struct {
	a   phys.Meters
	b   phys.Meters
	exp phys.Meters
}

// TestDofsDist checks Track.DofsDist()
func TestDofDist(t *testing.T) {
	trk, err := NewModularTrack(defTrackWidth, defTrackWidth/2, "SLLSSLLS") // Left Capsule
	if err != nil {
		t.Fatalf(err.Error())
	}
	trkLen := trk.CenLen()

	dofsDistTestTable := []testDofsDistVec{
		testDofsDistVec{a: 0.0, b: 0.0, exp: 0.0},
		testDofsDistVec{a: 0.1, b: 0.2, exp: 0.1},
		testDofsDistVec{a: 0.2, b: 0.1, exp: 0.1},
		testDofsDistVec{a: 0.1, b: 1.0, exp: 0.9},
		testDofsDistVec{a: 1.0, b: 0.1, exp: 0.9},

		testDofsDistVec{a: trkLen - 0.1, b: trkLen - 0.4, exp: 0.3},
		testDofsDistVec{a: trkLen - 0.4, b: trkLen - 0.1, exp: 0.3},
		testDofsDistVec{a: trkLen - 0.2, b: trkLen - 0.6, exp: 0.4},
		testDofsDistVec{a: trkLen - 0.6, b: trkLen - 0.2, exp: 0.4},

		testDofsDistVec{a: trkLen - 0.1, b: 0.1 /******/, exp: 0.2},
		testDofsDistVec{a: 0.1 /******/, b: trkLen - 0.1, exp: 0.2},
		testDofsDistVec{a: trkLen - 0.5, b: 0.1 /******/, exp: 0.6},
		testDofsDistVec{a: 0.1 /******/, b: trkLen - 0.5, exp: 0.6},
	}
	for _, v := range dofsDistTestTable {
		label := fmt.Sprintf("Track.DofsDist(%v, %v)", v.a, v.b)
		got := trk.DofsDist(v.a, v.b)
		testMetersAreNear(t, label, v.exp, got)
	}
}

//////////////////////////////////////////////////////////////////////

type testDriveDeltaDofsVec struct {
	frDofs   phys.Meters
	frDangle phys.Radians
	toDofs   phys.Meters
	exp      phys.Meters
}

type testDriveDeltaCofsVec struct {
	frCofs   phys.Meters
	frDangle phys.Radians
	toCofs   phys.Meters
	exp      phys.Meters
}

// TestDriveDelta tests Track.DriveDeltaXyz() functions
func TestDriveDelta(t *testing.T) {
	for _, topo := range []string{"SLLSSLLS", "SRSLLLSSRR"} { // Left Capsule, Right Loopback
		trk, err := NewModularTrack(defTrackWidth, defTrackWidth/2, topo)
		if err != nil {
			t.Fatalf(err.Error())
		}
		trkLen := trk.CenLen()
		halfTrkLen := trkLen / 2

		dofsTestTable := []testDriveDeltaDofsVec{
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +0.0 * math.Pi, toDofs: 0.0, exp: 0},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +1.0 * math.Pi, toDofs: 0.0, exp: 0},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: -1.0 * math.Pi, toDofs: 0.0, exp: 0},
			testDriveDeltaDofsVec{frDofs: 1.0, frDangle: +0.0 * math.Pi, toDofs: 1.0, exp: 0},
			testDriveDeltaDofsVec{frDofs: 1.0, frDangle: +1.0 * math.Pi, toDofs: 1.0, exp: 0},
			testDriveDeltaDofsVec{frDofs: 1.0, frDangle: -1.0 * math.Pi, toDofs: 1.0, exp: 0},

			testDriveDeltaDofsVec{frDofs: 1.0, frDangle: +0.0 * math.Pi, toDofs: 1.2, exp: +0.2},
			testDriveDeltaDofsVec{frDofs: 1.2, frDangle: +0.0 * math.Pi, toDofs: 1.0, exp: -0.2},
			testDriveDeltaDofsVec{frDofs: 1.0, frDangle: +1.0 * math.Pi, toDofs: 1.2, exp: -0.2},
			testDriveDeltaDofsVec{frDofs: 1.2, frDangle: +1.0 * math.Pi, toDofs: 1.0, exp: +0.2},

			testDriveDeltaDofsVec{frDofs: trkLen - 0.3, frDangle: +0.0 * math.Pi, toDofs: 0.2 /******/, exp: +0.5},
			testDriveDeltaDofsVec{frDofs: trkLen - 0.3, frDangle: -1.0 * math.Pi, toDofs: 0.2 /******/, exp: -0.5},
			testDriveDeltaDofsVec{frDofs: trkLen - 0.3, frDangle: +0.0 * math.Pi, toDofs: trkLen - 0.6, exp: -0.3},
			testDriveDeltaDofsVec{frDofs: trkLen - 0.3, frDangle: +1.0 * math.Pi, toDofs: trkLen - 0.6, exp: +0.3},

			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +0.0 * math.Pi, toDofs: halfTrkLen, exp: halfTrkLen},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +1.0 * math.Pi, toDofs: halfTrkLen, exp: halfTrkLen},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: -1.0 * math.Pi, toDofs: halfTrkLen, exp: halfTrkLen},

			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +0.0 * math.Pi, toDofs: halfTrkLen - 0.01, exp: +(halfTrkLen - 0.01)},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +1.0 * math.Pi, toDofs: halfTrkLen - 0.01, exp: -(halfTrkLen - 0.01)},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: +0.0 * math.Pi, toDofs: halfTrkLen + 0.01, exp: -(halfTrkLen - 0.01)},
			testDriveDeltaDofsVec{frDofs: 0.0, frDangle: -1.0 * math.Pi, toDofs: halfTrkLen + 0.01, exp: +(halfTrkLen - 0.01)},
		}
		for i, v := range dofsTestTable {
			pose := Pose{Point: Point{Dofs: v.frDofs, Cofs: 0.1}, DAngle: v.frDangle}
			label := fmt.Sprintf("Vec %d: Track.DriveDeltaDofs(%v, %v)", i, pose, v.toDofs)
			got := trk.DriveDeltaDofs(pose, v.toDofs)
			testMetersAreNear(t, label, v.exp, got)
		}

		cofsTestTable := []testDriveDeltaCofsVec{
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +0.0 * math.Pi, toCofs: 0.0, exp: 0},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +1.0 * math.Pi, toCofs: 0.0, exp: 0},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: -1.0 * math.Pi, toCofs: 0.0, exp: 0},

			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +0.0 * math.Pi, toCofs: +0.1, exp: +0.1},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +0.0 * math.Pi, toCofs: +0.3, exp: +0.3},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +0.0 * math.Pi, toCofs: -0.1, exp: -0.1},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +0.0 * math.Pi, toCofs: -0.4, exp: -0.4},

			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +1.0 * math.Pi, toCofs: +0.1, exp: -0.1},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +1.0 * math.Pi, toCofs: +0.3, exp: -0.3},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +1.0 * math.Pi, toCofs: -0.1, exp: +0.1},
			testDriveDeltaCofsVec{frCofs: 0.0, frDangle: +1.0 * math.Pi, toCofs: -0.4, exp: +0.4},
		}
		for i, v := range cofsTestTable {
			pose := Pose{Point: Point{Dofs: 0, Cofs: v.frCofs}, DAngle: v.frDangle}
			point := Point{Dofs: 0.01, Cofs: v.toCofs}
			label := fmt.Sprintf("Vec %d: Track.DriveDeltaCofs(%v, %v)", i, pose, point)
			got := trk.DriveDeltaCofs(pose, point.Cofs)
			testMetersAreNear(t, label, v.exp, got)
		}
	}
}

//////////////////////////////////////////////////////////////////////

type testDriveDistStrtVec struct {
	dofs1   phys.Meters
	dangle1 phys.Radians
	dofs2   phys.Meters
	exp     phys.Meters
}

func TestDriveDistStraights(t *testing.T) {
	// XXX: Use non-looping track, for easier test cases
	pieces := []RoadPiece{
		*NewRoadPiece(1.0, 0.0),
		*NewRoadPiece(1.0, 0.0),
		*NewRoadPiece(1.0, 0.0),
		*NewRoadPiece(1.0, 0.0),
	}
	trk, _ := NewTrack(defTrackWidth, defTrackWidth/2, pieces)

	testTable := []testDriveDistStrtVec{
		// Edge case: zero distance
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 0.0, exp: 0.0},
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 0.0},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 0.0},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 1.0 * math.Pi, dofs2: 1.0, exp: 0.0},

		// Within one road piece
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 0.1, exp: 0.1},
		testDriveDistStrtVec{dofs1: 0.2, dangle1: 0.0 * math.Pi, dofs2: 0.4, exp: 0.2},
		testDriveDistStrtVec{dofs1: 0.3, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 0.7},
		testDriveDistStrtVec{dofs1: 0.6, dangle1: 1.0 * math.Pi, dofs2: 0.5, exp: 0.1},
		testDriveDistStrtVec{dofs1: 0.5, dangle1: 1.0 * math.Pi, dofs2: 0.3, exp: 0.2},
		testDriveDistStrtVec{dofs1: 0.4, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 0.4},

		testDriveDistStrtVec{dofs1: 1.0, dangle1: 0.0 * math.Pi, dofs2: 1.1, exp: 0.1},
		testDriveDistStrtVec{dofs1: 1.2, dangle1: 0.0 * math.Pi, dofs2: 1.4, exp: 0.2},
		testDriveDistStrtVec{dofs1: 1.3, dangle1: 0.0 * math.Pi, dofs2: 2.0, exp: 0.7},
		testDriveDistStrtVec{dofs1: 1.6, dangle1: 1.0 * math.Pi, dofs2: 1.5, exp: 0.1},
		testDriveDistStrtVec{dofs1: 1.5, dangle1: 1.0 * math.Pi, dofs2: 1.3, exp: 0.2},
		testDriveDistStrtVec{dofs1: 1.4, dangle1: 1.0 * math.Pi, dofs2: 1.0, exp: 0.4},

		testDriveDistStrtVec{dofs1: 2.0, dangle1: 0.0 * math.Pi, dofs2: 2.1, exp: 0.1},
		testDriveDistStrtVec{dofs1: 2.2, dangle1: 0.0 * math.Pi, dofs2: 2.4, exp: 0.2},
		testDriveDistStrtVec{dofs1: 2.3, dangle1: 0.0 * math.Pi, dofs2: 3.0, exp: 0.7},
		testDriveDistStrtVec{dofs1: 2.6, dangle1: 1.0 * math.Pi, dofs2: 2.5, exp: 0.1},
		testDriveDistStrtVec{dofs1: 2.5, dangle1: 1.0 * math.Pi, dofs2: 2.3, exp: 0.2},
		testDriveDistStrtVec{dofs1: 2.4, dangle1: 1.0 * math.Pi, dofs2: 2.0, exp: 0.4},

		testDriveDistStrtVec{dofs1: 3.0, dangle1: 0.0 * math.Pi, dofs2: 3.1, exp: 0.1},
		testDriveDistStrtVec{dofs1: 3.2, dangle1: 0.0 * math.Pi, dofs2: 3.4, exp: 0.2},
		testDriveDistStrtVec{dofs1: 3.3, dangle1: 0.0 * math.Pi, dofs2: 0.0, exp: 0.7},
		testDriveDistStrtVec{dofs1: 3.6, dangle1: 1.0 * math.Pi, dofs2: 3.5, exp: 0.1},
		testDriveDistStrtVec{dofs1: 3.5, dangle1: 1.0 * math.Pi, dofs2: 3.3, exp: 0.2},
		testDriveDistStrtVec{dofs1: 3.4, dangle1: 1.0 * math.Pi, dofs2: 3.0, exp: 0.4},

		// Across two road pieces
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 2.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 1.9, exp: 1.9},
		testDriveDistStrtVec{dofs1: 0.2, dangle1: 0.0 * math.Pi, dofs2: 2.0, exp: 1.8},
		testDriveDistStrtVec{dofs1: 2.0, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 1.9, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 1.9},
		testDriveDistStrtVec{dofs1: 1.8, dangle1: 1.0 * math.Pi, dofs2: 0.3, exp: 1.5},

		testDriveDistStrtVec{dofs1: 1.0, dangle1: 0.0 * math.Pi, dofs2: 3.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 0.0 * math.Pi, dofs2: 2.9, exp: 1.9},
		testDriveDistStrtVec{dofs1: 1.2, dangle1: 0.0 * math.Pi, dofs2: 3.0, exp: 1.8},
		testDriveDistStrtVec{dofs1: 3.0, dangle1: 1.0 * math.Pi, dofs2: 1.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 2.9, dangle1: 1.0 * math.Pi, dofs2: 1.0, exp: 1.9},
		testDriveDistStrtVec{dofs1: 2.8, dangle1: 1.0 * math.Pi, dofs2: 1.3, exp: 1.5},

		testDriveDistStrtVec{dofs1: 2.0, dangle1: 0.0 * math.Pi, dofs2: 0.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 2.0, dangle1: 0.0 * math.Pi, dofs2: 3.9, exp: 1.9},
		testDriveDistStrtVec{dofs1: 2.2, dangle1: 0.0 * math.Pi, dofs2: 0.0, exp: 1.8},
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 1.0 * math.Pi, dofs2: 2.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 3.9, dangle1: 1.0 * math.Pi, dofs2: 2.0, exp: 1.9},
		testDriveDistStrtVec{dofs1: 3.8, dangle1: 1.0 * math.Pi, dofs2: 2.3, exp: 1.5},

		testDriveDistStrtVec{dofs1: 3.0, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 3.0, dangle1: 0.0 * math.Pi, dofs2: 0.9, exp: 1.9},
		testDriveDistStrtVec{dofs1: 3.2, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 1.8},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 1.0 * math.Pi, dofs2: 3.0, exp: 2.0},
		testDriveDistStrtVec{dofs1: 0.9, dangle1: 1.0 * math.Pi, dofs2: 3.0, exp: 1.9},
		testDriveDistStrtVec{dofs1: 0.8, dangle1: 1.0 * math.Pi, dofs2: 3.3, exp: 1.5},

		// Across three road pieces
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 3.0, exp: 3.0},
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 2.8, exp: 2.8},
		testDriveDistStrtVec{dofs1: 0.3, dangle1: 0.0 * math.Pi, dofs2: 3.0, exp: 2.7},
		testDriveDistStrtVec{dofs1: 0.4, dangle1: 0.0 * math.Pi, dofs2: 2.6, exp: 2.2},
		testDriveDistStrtVec{dofs1: 3.0, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 3.0},
		testDriveDistStrtVec{dofs1: 2.7, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 2.7},
		testDriveDistStrtVec{dofs1: 3.0, dangle1: 1.0 * math.Pi, dofs2: 0.4, exp: 2.6},
		testDriveDistStrtVec{dofs1: 2.6, dangle1: 1.0 * math.Pi, dofs2: 0.5, exp: 2.1},

		testDriveDistStrtVec{dofs1: 2.0, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 3.0},
		testDriveDistStrtVec{dofs1: 2.0, dangle1: 0.0 * math.Pi, dofs2: 0.8, exp: 2.8},
		testDriveDistStrtVec{dofs1: 2.3, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 2.7},
		testDriveDistStrtVec{dofs1: 2.4, dangle1: 0.0 * math.Pi, dofs2: 0.6, exp: 2.2},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 1.0 * math.Pi, dofs2: 2.0, exp: 3.0},
		testDriveDistStrtVec{dofs1: 0.7, dangle1: 1.0 * math.Pi, dofs2: 2.0, exp: 2.7},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 1.0 * math.Pi, dofs2: 2.4, exp: 2.6},
		testDriveDistStrtVec{dofs1: 0.6, dangle1: 1.0 * math.Pi, dofs2: 2.5, exp: 2.1},

		// Accross four road pieces
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 0.0 * math.Pi, dofs2: 3.9, exp: 3.9},
		testDriveDistStrtVec{dofs1: 0.2, dangle1: 0.0 * math.Pi, dofs2: 0.0, exp: 3.8},
		testDriveDistStrtVec{dofs1: 0.4, dangle1: 0.0 * math.Pi, dofs2: 0.1, exp: 3.7},
		testDriveDistStrtVec{dofs1: 3.9, dangle1: 1.0 * math.Pi, dofs2: 0.0, exp: 3.9},
		testDriveDistStrtVec{dofs1: 0.0, dangle1: 1.0 * math.Pi, dofs2: 0.2, exp: 3.8},
		testDriveDistStrtVec{dofs1: 0.1, dangle1: 1.0 * math.Pi, dofs2: 0.4, exp: 3.7},

		testDriveDistStrtVec{dofs1: 1.0, dangle1: 0.0 * math.Pi, dofs2: 0.9, exp: 3.9},
		testDriveDistStrtVec{dofs1: 1.2, dangle1: 0.0 * math.Pi, dofs2: 1.0, exp: 3.8},
		testDriveDistStrtVec{dofs1: 1.4, dangle1: 0.0 * math.Pi, dofs2: 1.1, exp: 3.7},
		testDriveDistStrtVec{dofs1: 0.9, dangle1: 1.0 * math.Pi, dofs2: 1.0, exp: 3.9},
		testDriveDistStrtVec{dofs1: 1.0, dangle1: 1.0 * math.Pi, dofs2: 1.2, exp: 3.8},
		testDriveDistStrtVec{dofs1: 1.1, dangle1: 1.0 * math.Pi, dofs2: 1.4, exp: 3.7},

		testDriveDistStrtVec{dofs1: 2.0, dangle1: 0.0 * math.Pi, dofs2: 1.9, exp: 3.9},
		testDriveDistStrtVec{dofs1: 2.2, dangle1: 0.0 * math.Pi, dofs2: 2.0, exp: 3.8},
		testDriveDistStrtVec{dofs1: 2.4, dangle1: 0.0 * math.Pi, dofs2: 2.1, exp: 3.7},
		testDriveDistStrtVec{dofs1: 1.9, dangle1: 1.0 * math.Pi, dofs2: 2.0, exp: 3.9},
		testDriveDistStrtVec{dofs1: 2.0, dangle1: 1.0 * math.Pi, dofs2: 2.2, exp: 3.8},
		testDriveDistStrtVec{dofs1: 2.1, dangle1: 1.0 * math.Pi, dofs2: 2.4, exp: 3.7},

		testDriveDistStrtVec{dofs1: 3.0, dangle1: 0.0 * math.Pi, dofs2: 2.9, exp: 3.9},
		testDriveDistStrtVec{dofs1: 3.2, dangle1: 0.0 * math.Pi, dofs2: 3.0, exp: 3.8},
		testDriveDistStrtVec{dofs1: 3.4, dangle1: 0.0 * math.Pi, dofs2: 3.1, exp: 3.7},
		testDriveDistStrtVec{dofs1: 2.9, dangle1: 1.0 * math.Pi, dofs2: 3.0, exp: 3.9},
		testDriveDistStrtVec{dofs1: 3.0, dangle1: 1.0 * math.Pi, dofs2: 3.2, exp: 3.8},
		testDriveDistStrtVec{dofs1: 3.1, dangle1: 1.0 * math.Pi, dofs2: 3.4, exp: 3.7},
	}
	for i, v := range testTable {
		pose := Pose{Point: Point{Dofs: v.dofs1, Cofs: 0}, DAngle: v.dangle1}
		label := fmt.Sprintf("Vec %d: Track.DriveDist(%v, %v)", i, pose, v.dofs2)
		got := trk.DriveDist(pose, v.dofs2)
		testMetersAreNear(t, label, v.exp, got)
	}
}

//////////////////////////////////////////////////////////////////////

type testRegionOffset struct {
	ofs        phys.Meters
	isInRegion bool
}

// regions that DO NOT cross the finish line
func TestTrackRegionsDoNotCrossFinishLine(t *testing.T) {
	topo := "SSRRSSRR" // Right Capsule
	width := phys.Meters(0.3)
	track, err := NewModularTrack(width, width/2, topo)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for i := 0; i < 5; i++ {
		dofs := phys.Meters(0.35 * float64(i))
		cofs := phys.Meters(0.5 - (0.11 * float64(i)))
		len := phys.Meters(0.3 * float64(i+1))
		width := phys.Meters(0.1 * float64(i+1)) // NOTE: can be wider than the track
		tr := NewRegion(track, Point{Dofs: dofs, Cofs: cofs}, len, width)
		testEqual(t, "tr.Len()", len, tr.Len())
		testEqual(t, "tr.Width()", width, tr.Width())
		testEqual(t, "tr.C1().Dofs", dofs /********/, tr.C1().Dofs)
		testEqual(t, "tr.C1().Cofs", cofs /********/, tr.C1().Cofs)
		testEqual(t, "tr.C2().Dofs", dofs+len /****/, tr.C2().Dofs)
		testEqual(t, "tr.C2().Cofs", cofs+width /**/, tr.C2().Cofs)
		testEqual(t, "tr.CrossesFinishLine()", false, tr.CrossesFinishLine())

		// Cherry-pick some offsets of interest that are inside and outside the
		// region. Note that the edges of the regions are inclusive at the start,
		// and exclusive at the end, for both distance and center dimensions.
		dofsVals := []testRegionOffset{
			testRegionOffset{ofs: dofs - 0.1 + 0.0, isInRegion: false},
			testRegionOffset{ofs: dofs + 0.0 + 0.0, isInRegion: true},
			testRegionOffset{ofs: dofs + 0.1 + 0.0, isInRegion: true},
			testRegionOffset{ofs: dofs + len + 0.0, isInRegion: false},
			testRegionOffset{ofs: dofs + len + 0.1, isInRegion: false},
		}
		cofsVals := []testRegionOffset{
			testRegionOffset{ofs: cofs - 0.100 + 0.0, isInRegion: false},
			testRegionOffset{ofs: cofs + 0.000 + 0.0, isInRegion: true},
			testRegionOffset{ofs: cofs + 0.050 + 0.0, isInRegion: true},
			testRegionOffset{ofs: cofs + width + 0.0, isInRegion: false},
			testRegionOffset{ofs: cofs + width + 0.1, isInRegion: false},
		}
		for _, d := range dofsVals {
			for _, c := range cofsVals {
				tp := Point{Dofs: d.ofs, Cofs: c.ofs}
				expContains := d.isInRegion && c.isInRegion
				condStr := fmt.Sprintf("tr=%v ContainsPoint(%v)", tr, tp)
				testEqual(t, condStr, expContains, tr.ContainsPoint(tp))
			}
		}
	}
}

// regions that DO cross the finish line
func TestTrackRegionsDoCrossFinishLine(t *testing.T) {
	topo := "SSRRSSRR" // Right Capsule
	width := phys.Meters(0.3)
	track, err := NewModularTrack(width, width/2, topo)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for i := 0; i < 5; i++ {
		dofs := track.CenLen() - phys.Meters(0.1*float64(i+1))
		cofs := phys.Meters(0.5 - (0.11 * float64(i)))
		len := phys.Meters(0.3 * float64(i+1))
		width := phys.Meters(0.1 * float64(i+1)) // NOTE: can be wider than the track
		tr := NewRegion(track, Point{Dofs: dofs, Cofs: cofs}, len, width)
		testEqual(t, "tr.Len()", len, tr.Len())
		testEqual(t, "tr.Width()", width, tr.Width())
		testEqual(t, "tr.C1().Dofs", dofs /****************/, tr.C1().Dofs)
		testEqual(t, "tr.C1().Cofs", cofs /****************/, tr.C1().Cofs)
		testEqual(t, "tr.C2().Dofs", dofs+len-track.CenLen(), tr.C2().Dofs)
		testEqual(t, "tr.C2().Cofs", cofs+width /**********/, tr.C2().Cofs)
		testEqual(t, "tr.CrossesFinishLine()", true, tr.CrossesFinishLine())

		// Cherry-pick some offsets of interest that are inside and outside the
		// region. Note that the edges of the regions are inclusive at the start,
		// and exclusive at the end, for both distance and center dimensions.
		dofsVals := []testRegionOffset{
			testRegionOffset{ofs: dofs - 0.1 + 0.0, isInRegion: false},
			testRegionOffset{ofs: dofs + 0.0 + 0.0, isInRegion: true},
			testRegionOffset{ofs: dofs + 0.1 + 0.0, isInRegion: true},
			testRegionOffset{ofs: dofs + len + 0.0, isInRegion: false},
			testRegionOffset{ofs: dofs + len + 0.1, isInRegion: false},
		}
		cofsVals := []testRegionOffset{
			testRegionOffset{ofs: cofs - 0.100 + 0.0, isInRegion: false},
			testRegionOffset{ofs: cofs + 0.000 + 0.0, isInRegion: true},
			testRegionOffset{ofs: cofs + 0.050 + 0.0, isInRegion: true},
			testRegionOffset{ofs: cofs + width + 0.0, isInRegion: false},
			testRegionOffset{ofs: cofs + width + 0.1, isInRegion: false},
		}
		for _, d := range dofsVals {
			for _, c := range cofsVals {
				tp := Point{Dofs: d.ofs, Cofs: c.ofs}
				for j := 0; j < 2; j++ {
					if (j == 1) && (tp.Dofs >= track.CenLen()) {
						tp.Dofs -= track.CenLen() // normalize
					}
					expContains := d.isInRegion && c.isInRegion
					condStr := fmt.Sprintf("tr=%v ContainsPoint(%v)", tr, tp)
					testEqual(t, condStr, expContains, tr.ContainsPoint(tp))
				}
			}
		}
	}
}

// test regions with (tr.Len >= track.CenLen())
func TestTrackRegionsFullLength(t *testing.T) {
	topo := "SSRRSSRR" // Right Capsule
	width := phys.Meters(0.4)
	track, err := NewModularTrack(width, width/2, topo)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for i := 0; i < 5; i++ {
		dofs := phys.Meters(0.2 * float64(i))
		cofs := phys.Meters(-0.1)
		len := track.CenLen() + phys.Meters(0.1*float64(i))
		width := phys.Meters(0.2)
		tr := NewRegion(track, Point{Dofs: dofs, Cofs: cofs}, len, width)
		testEqual(t, "tr.Len()", len, tr.Len())
		testEqual(t, "tr.Width()", width, tr.Width())
		testEqual(t, "tr.C1().Dofs", dofs, tr.C1().Dofs)
		testEqual(t, "tr.C1().Cofs", cofs, tr.C1().Cofs)
		testEqual(t, "tr.CrossesFinishLine()", true, tr.CrossesFinishLine())

		// Cherry-pick some offsets of interest that are inside and outside the
		// region. Note that the edges of the regions are inclusive at the start,
		// and exclusive at the end, for both distance and center dimensions.
		cofsVals := []testRegionOffset{
			testRegionOffset{ofs: cofs - 0.100 + 0.0, isInRegion: false},
			testRegionOffset{ofs: cofs + 0.000 + 0.0, isInRegion: true},
			testRegionOffset{ofs: cofs + 0.050 + 0.0, isInRegion: true},
			testRegionOffset{ofs: cofs + width + 0.0, isInRegion: false},
			testRegionOffset{ofs: cofs + width + 0.1, isInRegion: false},
		}
		for d := phys.Meters(0); d < (track.CenLen() + phys.Meters(0.4)); d += phys.Meters(0.1) {
			for _, c := range cofsVals {
				tp := Point{Dofs: d, Cofs: c.ofs}
				expContains := c.isInRegion
				condStr := fmt.Sprintf("tr=%v ContainsPoint(%v)", tr, tp)
				testEqual(t, condStr, expContains, tr.ContainsPoint(tp))
			}
		}
	}
}
