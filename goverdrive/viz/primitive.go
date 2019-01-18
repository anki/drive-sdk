// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package viz

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/phys"
)

// PrimitiveVisualizer provides drawing primitives whose values are in absolute
// cartesian space, using Meters. Primitives are rendered onto a canvas.
//
// Intended usage pattern:
//   pv.ClearAndReset()
//   pv.AddLine()
//   pv.AddRectangle()
//   ...  // remaining shapes
//   pv.RenderAll(canvas)
//   // Display canvas in a window
type PrimitiveVisualizer interface {
	// ClearAndReset clears all drawn shapes and resets the internal state for a
	// "clean slate".
	ClearAndReset()

	// RenderAll renders all of the shapes that have been added since the last
	// call to ClearAndReset(). Shapes are rendered onto the passed-in canvas.
	RenderAll(canvas *pixelgl.Canvas)

	// AddLine adds a line between two points
	AddLine(p1, p2 phys.Point, thickness phys.Meters, clr color.Color)

	// AddRectangle adds a rectangle based on the opposite corners. When
	// thickness==0, the rectangle is filled in.
	AddRectangle(v1, v2 phys.Point, thickness phys.Meters, clr color.Color)

	// AddCircle adds a circle based on center point and radius. When
	// thickness==0, the circle is filled in.
	AddCircle(ctr phys.Point, rad phys.Meters, thickness phys.Meters, clr color.Color)

	// AddCircleArc adds circle arc based on center point and radius, and the
	// beginning and end angles. When thickness==0, the circle arc is filled in.
	AddCircleArc(ctr phys.Point, rad phys.Meters, begAngle phys.Radians, endAngle phys.Radians, thickness phys.Meters, clr color.Color)
}

// XXX: The window and canvas modules think in terms of pixels, while most
// individual track and game objects are < 1.0 Meters. The primitive visualizer
// scales meters into a more usable pixel space.
const PixPerMeter float64 = 1000.0

//////////////////////////////////////////////////////////////////////

// PixelViz satisfies PrimitiveVisualizer interface, using the package
// github.com/faiface/pixel.
type PixelViz struct {
	imd *imdraw.IMDraw
}

func NewPixelViz() *PixelViz {
	imd := imdraw.New(nil)
	return &PixelViz{imd: imd}
}

func (pv *PixelViz) ClearAndReset() {
	pv.imd.Clear()
	pv.imd.Reset()
}

func (pv *PixelViz) RenderAll(canvas *pixelgl.Canvas) {
	pv.imd.Draw(canvas)
}

// metersToPix handles conversion of units and data type casting
func metersToPix(m phys.Meters) float64 {
	return PixPerMeter * float64(m)
}

func (pv *PixelViz) AddLine(p1, p2 phys.Point, thickness phys.Meters, clr color.Color) {
	pv.imd.Color = clr
	pv.imd.Push(pixel.Vec{X: metersToPix(p1.X), Y: metersToPix(p1.Y)})
	pv.imd.Push(pixel.Vec{X: metersToPix(p2.X), Y: metersToPix(p2.Y)})
	pv.imd.Line(metersToPix(thickness))
}

func (pv *PixelViz) AddRectangle(v1, v2 phys.Point, thickness phys.Meters, clr color.Color) {
	pv.imd.Color = clr
	pv.imd.Push(pixel.Vec{X: metersToPix(v1.X), Y: metersToPix(v1.Y)})
	pv.imd.Push(pixel.Vec{X: metersToPix(v2.X), Y: metersToPix(v2.Y)})
	pv.imd.Rectangle(metersToPix(thickness))
}

func (pv *PixelViz) AddCircle(ctr phys.Point, rad phys.Meters, thickness phys.Meters, clr color.Color) {
	pv.imd.Color = clr
	pv.imd.Push(pixel.Vec{X: metersToPix(ctr.X), Y: metersToPix(ctr.Y)})
	pv.imd.Circle(metersToPix(rad), metersToPix(thickness))
}

func (pv *PixelViz) AddCircleArc(ctr phys.Point, rad phys.Meters, begAngle phys.Radians, endAngle phys.Radians, thickness phys.Meters, clr color.Color) {
	pv.imd.Color = clr
	pv.imd.Push(pixel.Vec{X: metersToPix(ctr.X), Y: metersToPix(ctr.Y)})
	pv.imd.CircleArc(metersToPix(rad), float64(begAngle), float64(endAngle), metersToPix(thickness))
}
