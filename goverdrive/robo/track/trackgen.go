// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com
//
// trackgen provides convenience functions to generate common tracks, such as
// the standard tracks in the OverDrive starter kit.

package track

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/anki/goverdrive/phys"
)

// NewModularTrack constructs a track using standard modular track pieces:
//   S = straight
//   R = right turn (90 degrees)
//   L = left  turn (90 degrees)

// Examples:
//   topo="SRRSSRRS"   => Right Capsule
//   topo="SLSRRRSSLL" => Left  Loopback
//
// The first letter of the topo string must be `S`, and this will become the
// standard Short/Long start piece; the first track piece is Start Short and the
// last track piece is Start Long.
func NewModularTrack(width phys.Meters, maxCofs phys.Meters, topo string) (*Track, error) {
	if topo[0] != 'S' {
		return nil, fmt.Errorf("NewModularTrack topo string must start with 'S'. topo=%s", topo)
	}
	numRp := len(topo) + 1 // 1st straight is two road pieces
	pieces := make([]RoadPiece, numRp, numRp)

	for i, tc := range topo {
		if i == 0 {
			pieces[0] = *NewRoadPiece(TrackLenModStartShort, 0)
			continue
		}
		switch tc {
		case 'S':
			pieces[i] = *NewRoadPiece(TrackLenModStraight, 0)
		case 'L':
			pieces[i] = *NewRoadPiece(TrackLenModCurve, phys.Radians90DegreeTurnL)
		case 'R':
			pieces[i] = *NewRoadPiece(TrackLenModCurve, phys.Radians90DegreeTurnR)
		default:
			return nil, fmt.Errorf("Unsupported character in track topology string: %v", tc)
		}
	}
	pieces[numRp-1] = *NewRoadPiece(TrackLenModStartLong, 0)

	return NewTrack(width, maxCofs, pieces)
}

// kStarterKitTracks defines topology strings for starter kit tracks. Do not
// modify this variable!
var kStarterKitTracks = map[string]string{
	"cap":        "SLLSLL", // XXX: "Cap" is a shorthand for "Microloop"
	"lcap":       "SLLSLL",
	"rcap":       "SRRSRR",
	"microloop":  "SLLSLL",
	"lmicroloop": "SLLSLL",
	"rmicroloop": "SRRSRR",
	"capsule":    "SLLSSLLS",
	"lcapsule":   "SLLSSLLS",
	"rcapsule":   "SRRSSRRS",
	"quadra":     "SLSLSLSL",
	"lquadra":    "SLSLSLSL",
	"rquadra":    "SRSRSRSR",
	"point":      "SLSLSLRLLS",
	"lpoint":     "SLSLSLRLLS",
	"rpoint":     "SRSRSRLRRS",
	"wedge":      "SLSLLRLL",
	"lwedge":     "SLSLLRLL",
	"rwedge":     "SRSRRLRR",
	"hook":       "SSLSLLRSLL",
	"lhook":      "SSLSLLRSLL",
	"rhook":      "SSRSRRLSRR",
	"overpass":   "SLLLSRRR",
	"loverpass":  "SLLLSRRR",
	"roverpass":  "SRRRSLLL",
	"loopback":   "SLSRRRSSLL",
	"lloopback":  "SLSRRRSSLL",
	"rloopback":  "SRSLLLSSRR",
}

// StarterKitTrackNames returns a string with all of the supported starter kit
// track names.
func StarterKitTrackNames(seperator string) string {
	s := ""
	for name, _ := range kStarterKitTracks {
		s += name + seperator
	}
	return s
}

// NewStarterKitTrack constructs a variant of one of the OverDrive starter kit
// modular tracks.
//
// The track string has the form "Name_N", where:
//   Name = {microloop, capsule, quadra, point, wedge, hook, overpass, loopback}
//          Can add 'l' or 'r' prefix to any name to dictate the direction of
//          the first curve.
//   _N   = (optional) replace each straight piece with N straight pieces
//          This only creates a valid track for a subset of the tracks.
func NewStarterKitTrack(width phys.Meters, maxCofs phys.Meters, trackStr string) (*Track, error) {
	trackStr = strings.ToLower(trackStr)
	trackName := trackStr
	paramIdx := strings.Index(trackStr, "_")
	straightRep := 1
	if paramIdx != -1 {
		trackName = trackStr[0:paramIdx]
		val, err := strconv.Atoi(trackStr[paramIdx+1:])
		if err == nil {
			straightRep = val
		}
	}

	topo, ok := kStarterKitTracks[trackName]
	if !ok {
		return nil, fmt.Errorf("trackName=%s is not recognized", trackName)
	}
	topo = strings.Replace(topo, "S", strings.Repeat("S", straightRep), -1)

	return NewModularTrack(width, maxCofs, topo)
}

//////////////////////////////////////////////////////////////////////

// kCustomTrackNames keeps the list of names, for documentation strings.
var kCustomTrackNames = map[string]bool{
	"miniocto":   true,
	"miniquadra": true,
	"minicap":    true,
	"minirhom":   true,
	"minitrap":   true,
	"triangle":   true,
	"go":         true,
	"oval":       true,
}

// CustomTrackNames returns a string with all of the supported custom track
// names.
func CustomTrackNames(seperator string) string {
	s := ""
	for name, _ := range kCustomTrackNames {
		s += name + seperator
	}
	return s
}

// NewCustomTrack constructs a specific named custom track.
func NewCustomTrack(width phys.Meters, maxCofs phys.Meters, name string) (*Track, error) {
	if !kCustomTrackNames[name] {
		return nil, fmt.Errorf("Custom track name=%v is not recognized", name)
	}

	// "mini" tracks are built with short straights and 45-degree turns
	miniStraight := *NewRoadPiece(0.3, 0.0)
	miniCurve45 := *NewRoadPiece(0.3, math.Pi/4)

	switch strings.ToLower(name) {
	case "miniocto": // octogon
		pieces := []RoadPiece{
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniCurve45,
		}
		return NewTrack(width, maxCofs, pieces)

	case "miniquadra":
		pieces := []RoadPiece{
			miniStraight, miniStraight, miniCurve45, miniCurve45,
			miniStraight, miniStraight, miniCurve45, miniCurve45,
			miniStraight, miniStraight, miniCurve45, miniCurve45,
			miniStraight, miniStraight, miniCurve45, miniCurve45,
		}
		return NewTrack(width, maxCofs, pieces)

	case "minicap": // capsule
		pieces := []RoadPiece{
			miniStraight, miniStraight, miniStraight, miniStraight, miniCurve45, miniCurve45, miniCurve45, miniCurve45,
			miniStraight, miniStraight, miniStraight, miniStraight, miniCurve45, miniCurve45, miniCurve45, miniCurve45,
		}
		return NewTrack(width, maxCofs, pieces)

	case "minirhom": // rhombus
		pieces := []RoadPiece{
			miniStraight, miniStraight, miniCurve45, miniCurve45, miniCurve45,
			miniStraight, miniStraight, miniCurve45,
			miniStraight, miniStraight, miniCurve45, miniCurve45, miniCurve45,
			miniStraight, miniStraight, miniCurve45,
		}
		return NewTrack(width, maxCofs, pieces)

	case "minitrap": // trapezoid
		pieces := []RoadPiece{
			miniStraight, miniStraight, miniCurve45, miniCurve45, miniCurve45,
			miniStraight, miniCurve45,
			miniStraight, miniStraight, miniCurve45, miniCurve45, miniCurve45,
			miniStraight, miniCurve45,
		}
		return NewTrack(width, maxCofs, pieces)

	case "triangle":
		pieces := []RoadPiece{
			*NewRoadPiece(1.0, 0.0),
			*NewRoadPiece(0.3, 2*math.Pi/6),
			*NewRoadPiece(0.3, 2*math.Pi/6),
			*NewRoadPiece(1.0, 0.0),
			*NewRoadPiece(0.3, 2*math.Pi/6),
			*NewRoadPiece(0.3, 2*math.Pi/6),
			*NewRoadPiece(1.0, 0.0),
			*NewRoadPiece(0.3, 2*math.Pi/6),
			*NewRoadPiece(0.3, 2*math.Pi/6),
		}
		return NewTrack(width, maxCofs, pieces)

	case "go":
		return NewModularTrack(width, maxCofs, "SSSSLSLSLSRLSRSRRRLLSSSLSLSL")

	case "oval":
		pieces := []RoadPiece{
			*NewRoadPiece(0.30, 0.0),
			*NewRoadPiece(1.20, 0.0),
			*NewRoadPiece(0.45, math.Pi/2.67),
			*NewRoadPiece(0.45, math.Pi/2.67),
			*NewRoadPiece(0.45, math.Pi/2.67),
			*NewRoadPiece(1.25, -2*math.Pi/2.7/3),
			*NewRoadPiece(0.45, math.Pi/2.67),
			*NewRoadPiece(0.45, math.Pi/2.67),
			*NewRoadPiece(0.45, math.Pi/2.67),
		}
		return NewTrack(width, maxCofs, pieces)

	default:
		return nil, fmt.Errorf("Custom track name=%v is not recognized", name)
	}
}
