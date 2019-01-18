// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package vehlights provides additional functionality for vehicle light
// effects.
package vehlights

import (
	_ "fmt"
	"golang.org/x/image/colornames"
	"image/color"

	"github.com/anki/goverdrive/phys"
)

type SpeedColorPair struct {
	Speed phys.MetersPerSec
	Color color.Color
}

// DefSpeedometerColors is a nice "default" color map for using a vehicle light
// as a speedometer.
var DefSpeedometerColors = []SpeedColorPair{
	SpeedColorPair{0.3, colornames.Black},
	SpeedColorPair{0.7, colornames.Darkkhaki},
	SpeedColorPair{1.0, colornames.Lime},
	SpeedColorPair{1.4, colornames.White},
}

// SpeedometerColor chooses a color by linearly interpolating between N many
// user-defined points in color space.
func SpeedometerColor(clrMap []SpeedColorPair, speed phys.MetersPerSec) color.Color {
	n := len(clrMap)
	if n == 0 {
		return color.Black
	}

	// edge-case behavior
	if speed < clrMap[0].Speed {
		return clrMap[0].Color
	}
	if speed >= clrMap[n-1].Speed {
		return clrMap[n-1].Color
	}

	// interpolate
	for i := 0; i < (n - 1); i++ {
		if clrMap[i+1].Speed > speed {
			percent := (float64(speed) - float64(clrMap[i].Speed)) / (float64(clrMap[i+1].Speed) - float64(clrMap[i].Speed))
			var a uint32
			c1 := make([]uint32, 3)
			c2 := make([]uint32, 3)
			c1[0], c1[1], c1[2], a = clrMap[i+0].Color.RGBA()
			c2[0], c2[1], c2[2], _ = clrMap[i+1].Color.RGBA()
			c := make([]uint8, 3)
			for i := 0; i < 3; i++ {
				if c2[i] > c1[i] {
					c[i] = uint8(c1[i]) + uint8(percent*float64(uint8(c2[i])-uint8(c1[i])))
				} else {
					c[i] = uint8(c1[i]) - uint8(percent*float64(uint8(c1[i])-uint8(c2[i])))
				}
			}
			//fmt.Printf("i=%v percent=%v (%v %v %v) (%v %v %v) => %v %v %v\n", i, percent, c1[0], c1[1], c1[2], c2[0], c2[1], c2[2], c[0], c[1], c[2])
			return color.RGBA{R: c[0], G: c[1], B: c[2], A: uint8(a)}
		}
	}
	panic("CalcSpeedometerColor reached end of function")
	return color.RGBA{0, 0, 0, 0}
}
