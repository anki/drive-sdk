// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package viz renders graphics, such as tracks, vehicles, and weapon effects,
// onto a canvas for visualization. It does not actually handle scaling or
// displaying the canvas in a window.
//
// Visualization features are fairly limited. Tracks, track regions, and
// vehicles are natively supported. Anything beyond this is limited to a few
// primitive geometric shapes, such as lines and circles.
package viz

import (
	"fmt"
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/track"
)

//////////////////////////////////////////////////////////////////////
// CONSTANTS
//////////////////////////////////////////////////////////////////////

const (
	KTrackRegionThickness phys.Meters = 0.005
	KFinishLineThickness  phys.Meters = 0.010
)

var (
	KTrackOutlineColor    color.Color = colornames.White
	KTrackCenterColor     color.Color = colornames.Yellow
	KTrackFinishLineColor color.Color = colornames.Lawngreen
)

//////////////////////////////////////////////////////////////////////
// TYPES AND INTERFACES
//////////////////////////////////////////////////////////////////////

const (
	WorldVizPadding = phys.Meters(0.08)
)

// TrackRegion for vizualization needs a color
type TrackRegion struct {
	track.Region
	Color color.Color
}

// WorldViz visualizes the objects in the goverdrive "world", such as tracks and
// vehicles.
type WorldViz interface {
	// MinCorner returns the coordinate the smallest corner of the visible world
	MinCorner() phys.Point

	// MaxCorner returns the coordinate the largest corner of the visible world
	MaxCorner() phys.Point

	// RenderAll() renders each set of game objects onto a canvas.
	//   - The object sets are rendered in the order they are passed in. Ie the
	//     track regions are rendered before the vehicles.
	//   - Within an object set, objects are rendered in the order they occur
	//     within the slice.
	RenderAll(track *track.Track, regions *[]*TrackRegion, vehs *[]robo.Vehicle, shapes *[]*GameShape) *pixelgl.Canvas
}

//////////////////////////////////////////////////////////////////////

// PixelWorldViz satisfies WorldViz interface, using the package
// github.com/faiface/pixel.
type PixelWorldViz struct {
	pv        PrimitiveVisualizer
	minCorner phys.Point // minimum corner of the visible world
	maxCorner phys.Point // maximum corner of the visible world
	canvas    *pixelgl.Canvas
}

func NewPixelWorldViz(pv PrimitiveVisualizer, track *track.Track) *PixelWorldViz {
	// add padding so track display looks nicer, and there is a little room for
	// game objects, off-track driving, etc.
	minCorner := track.MinCorner()
	maxCorner := track.MaxCorner()
	minCorner.X -= WorldVizPadding
	minCorner.Y -= WorldVizPadding
	maxCorner.X += WorldVizPadding
	maxCorner.Y += WorldVizPadding

	return &PixelWorldViz{
		pv:        pv,
		minCorner: minCorner,
		maxCorner: maxCorner,
		canvas:    nil,
	}
}

func (wv *PixelWorldViz) MinCorner() phys.Point {
	return wv.minCorner
}

func (wv *PixelWorldViz) MaxCorner() phys.Point {
	return wv.maxCorner
}

func (wv *PixelWorldViz) RenderAll(trk *track.Track, regions *[]*TrackRegion, vehs *[]robo.Vehicle, shapes *[]*GameShape) *pixelgl.Canvas {
	if wv.canvas == nil {
		bounds := pixel.R(
			PixPerMeter*float64(wv.minCorner.X),
			PixPerMeter*float64(wv.minCorner.Y),
			PixPerMeter*float64(wv.maxCorner.X),
			PixPerMeter*float64(wv.maxCorner.Y))
		wv.canvas = pixelgl.NewCanvas(bounds)
		fmt.Printf("canvas.Bounds()=%v\n", wv.canvas.Bounds())
	}

	wv.canvas.Clear(colornames.Black)
	wv.pv.ClearAndReset()

	// Track
	wv.addTrack(trk)

	// Track Regions
	for _, tr := range *regions {
		wv.addTrackRegion(trk, tr)
	}

	// Vehicles
	for i, _ := range *vehs {
		wv.addVehicle(i, trk, vehs)
	}

	// Game Shapes
	for _, shape := range *shapes {
		wv.addGameShape(shape, trk, vehs)
	}

	wv.pv.RenderAll(wv.canvas)
	return wv.canvas
}

// addLineAtPose adds a line between two points whose locations are relative to
// the position + rotation of a pose.
func (wv *PixelWorldViz) addLineAtPose(p phys.Pose, p1, p2 phys.Point, thickness phys.Meters, clr color.Color) {
	pp1 := p1.ToPolarPoint()
	pp2 := p2.ToPolarPoint()

	// p.A == 0 means facing +X direction
	pp1.A += p.Theta
	pp2.A += p.Theta

	p1 = pp1.ToPoint()
	p2 = pp2.ToPoint()

	p1.X += p.X
	p1.Y += p.Y
	p2.X += p.X
	p2.Y += p.Y

	wv.pv.AddLine(p1, p2, thickness, clr)
}

//////////////////////////////////////////////////////////////////////

// addRoadPieceDLine renders a "distance" line at a specified center offset, on
// a particular road piece. It is required that:
//   1. begDofs >= 0
//   2. endDofs > begDofs
//   3. endDofs <= rp.CenLen()
func (wv *PixelWorldViz) addRoadPieceDLine(trk *track.Track, rpi track.Rpi, cofs, begDofs, endDofs, thickness phys.Meters, clr color.Color) {
	rp := trk.Rp(rpi)
	if begDofs < 0 {
		panic(fmt.Sprintf("addRoadPieceDLine requires (begDofs >= 0), but begVof =%v", begDofs))
	}
	if begDofs >= endDofs {
		panic(fmt.Sprintf("addRoadPieceDLine requires (endDofs > begDofs), but begDofs=%v, endDofs=%v", begDofs, endDofs))
	}
	if endDofs > rp.CenLen() {
		panic(fmt.Sprintf("addRoadPieceDLine requires (endDofs <= rp.CenLen()), but endDofs=%v, rp.CenLen()=%v", endDofs, rp.CenLen()))
	}

	if rp.IsStraight() {
		p1 := phys.Point{X: begDofs, Y: cofs}
		p2 := phys.Point{X: endDofs, Y: cofs}
		wv.addLineAtPose(trk.RpEntryPose(rpi), p1, p2, thickness, clr)
	} else { // curved road piece
		// compute absolute beg/end angles of the circle arc
		rad := rp.CurveRadius(cofs)
		ctr := trk.RpCurveCenter(rpi)
		angles := make([]phys.Radians, 2, 2)
		for i := 0; i < 2; i++ {
			pose := trk.RpEntryPose(rpi + track.Rpi(i))
			pp := phys.Point{X: pose.X - ctr.X, Y: pose.Y - ctr.Y}.ToPolarPoint()
			angles[i] = pp.A
		}
		if (rp.DAngle() > 0) && (angles[1] < angles[0]) {
			angles[1] += (2 * math.Pi)
		}
		if (rp.DAngle() < 0) && (angles[1] > angles[0]) {
			angles[1] -= (2 * math.Pi)
		}
		// angles[] now has the beg/end angles for drawing a vert line for the WHOLE
		// piece. Need to adjust for begCofs/endCofs.
		dAngle := angles[1] - angles[0]
		angles[1] = angles[0] + (dAngle * phys.Radians(endDofs/rp.CenLen()))
		angles[0] = angles[0] + (dAngle * phys.Radians(begDofs/rp.CenLen()))
		//fmt.Printf("rpi=%v: angles[0]=%v, angles[1]=%v\n", rpi, angles[0], angles[1])
		wv.pv.AddCircleArc(ctr, rad, angles[0], angles[1], thickness, clr)
	}
}

// addTrackDLine renders a "distance" line at a specified center offset on the
// track. The line follows the curvature of the track. When (dofs1 > dofs2), the
// distance line will cross the finish line.
func (wv *PixelWorldViz) addTrackDLine(trk *track.Track, cofs, dofs1, dofs2, thickness phys.Meters, clr color.Color) {
	rpi1, rpDofs1 := trk.RpiAndRpDofs(dofs1)
	rpi2, rpDofs2 := trk.RpiAndRpDofs(dofs2)
	if rpi1 == rpi2 {
		wv.addRoadPieceDLine(trk, rpi1, cofs, rpDofs1, rpDofs2, thickness, clr)
		return
	}
	rpCount := int(rpi2 - rpi1)
	if rpi2 < rpi1 {
		rpCount += trk.NumRp()
	}
	for i := 0; i < rpCount; i++ {
		rpi := rpi1 + track.Rpi(i)
		if int(rpi) >= trk.NumRp() {
			rpi -= track.Rpi(trk.NumRp())
		}
		rp := trk.Rp(rpi)
		//fmt.Printf("addTrackDLine(rpi=%v): cofs=%v, dofs1=%v, vof2=%v, rpDofs1=%v, rpDofs2=%v, rp.CenLen()=%v\n", rpi, cofs, dofs1, dofs2, rpDofs1, rpDofs2, rp.CenLen())
		if rpi == rpi1 {
			wv.addRoadPieceDLine(trk, rpi, cofs, rpDofs1, rp.CenLen(), thickness, clr)
		} else {
			wv.addRoadPieceDLine(trk, rpi, cofs, 0, rp.CenLen(), thickness, clr)
		}
	}
	if rpDofs2 > 1.0e-6 {
		wv.addRoadPieceDLine(trk, rpi2, cofs, 0, rpDofs2, thickness, clr)
	}
}

// addTrackCLine renders a "center offset" line at a specified distance offset
// on the track.
func (wv *PixelWorldViz) addTrackCLine(trk *track.Track, dofs, cofs1, cofs2, thickness phys.Meters, clr color.Color) {
	pose := trk.ToPose(track.Pose{Point: track.Point{Dofs: dofs, Cofs: 0}, DAngle: 0})

	// line endpoints are relative to the pose of the track
	p1 := phys.Point{X: 0, Y: cofs1}
	p2 := phys.Point{X: 0, Y: cofs2}

	wv.addLineAtPose(pose, p1, p2, thickness, clr)
}

// addTrackRegion renders an unfilled track region which bends to the shape of
// the track.
func (wv *PixelWorldViz) addTrackRegion(track *track.Track, tr *TrackRegion) {
	if tr.Len() >= track.CenLen() {
		// XXX(gwenz): This avoids crashes and incorrectly rendered track regions.
		// Probably it should not be necessary, with proper rendering algorithms.
		wv.addTrackDLine(track, tr.C1().Cofs, 0, track.CenLen(), KTrackRegionThickness, tr.Color)
		wv.addTrackDLine(track, tr.C2().Cofs, 0, track.CenLen(), KTrackRegionThickness, tr.Color)
		return
	}
	wv.addTrackCLine(track, tr.C1().Dofs, tr.C1().Cofs, tr.C2().Cofs, KTrackRegionThickness, tr.Color)
	wv.addTrackCLine(track, tr.C2().Dofs, tr.C1().Cofs, tr.C2().Cofs, KTrackRegionThickness, tr.Color)
	wv.addTrackDLine(track, tr.C1().Cofs, tr.C1().Dofs, tr.C2().Dofs, KTrackRegionThickness, tr.Color)
	wv.addTrackDLine(track, tr.C2().Cofs, tr.C1().Dofs, tr.C2().Dofs, KTrackRegionThickness, tr.Color)
}

// addTrack performs the individual commands to render an entire track.
func (wv *PixelWorldViz) addTrack(trk *track.Track) {
	// finish line
	flL := phys.Point{X: 0, Y: +trk.Width() / 2}
	flR := phys.Point{X: 0, Y: -trk.Width() / 2}
	flTrackPose := track.Pose{Point: track.Point{Dofs: track.TrackLenModStartShort, Cofs: 0}, DAngle: 0}
	flPoint := trk.ToPose(flTrackPose).Point
	wv.pv.AddLine(flL, flPoint, KFinishLineThickness/3, KTrackFinishLineColor)
	wv.pv.AddLine(flR, flPoint, KFinishLineThickness/3, KTrackFinishLineColor)
	wv.addTrackCLine(trk, 0, -trk.Width()/2, +trk.Width()/2, KFinishLineThickness, KTrackFinishLineColor)

	// Easiest way to render is to make each road piece a single track region.
	for rpi := track.Rpi(0); rpi < track.Rpi(trk.NumRp()); rpi++ {
		rp := trk.Rp(rpi)
		cenLen := rp.CenLen()

		centerTrC1 := track.Point{Dofs: trk.RpEntryDofs(rpi), Cofs: 0}
		centerTr := track.NewRegion(trk, centerTrC1, cenLen, 0.0001)
		centerRegion := TrackRegion{
			Region: *centerTr,
			Color:  KTrackCenterColor,
		}
		wv.addTrackRegion(trk, &centerRegion)

		outlineTrC1 := track.Point{Dofs: trk.RpEntryDofs(rpi), Cofs: -trk.Width() / 2}
		outlineTr := track.NewRegion(trk, outlineTrC1, cenLen, trk.Width())
		outlineRegion := TrackRegion{
			Region: *outlineTr,
			Color:  KTrackOutlineColor,
		}
		wv.addTrackRegion(trk, &outlineRegion)
	}
}

// addVehicle renders a vehicle at its position on the track
func (wv *PixelWorldViz) addVehicle(vehId int, track *track.Track, vehs *[]robo.Vehicle) {
	v := &(*vehs)[vehId]
	// car body = colored rectangle
	wv.addLineAtPose(track.ToPose(v.CurTrackPose()),
		phys.Point{X: -(v.Length() / 2), Y: 0},
		phys.Point{X: +(v.Length() / 2), Y: 0},
		v.Width(), v.Color())
	// lights = filled circles
	for _, lvi := range v.Lights().VizInfo() {
		gs := NewCartesGameCirc(vehId, phys.Point{X: lvi.X, Y: lvi.Y}, lvi.R, lvi.Color, 0)
		wv.addGameShape(gs, track, vehs)
	}
}

// addGameShape renders the appropriate game shape
func (wv *PixelWorldViz) addGameShape(gs *GameShape, trk *track.Track, vehs *[]robo.Vehicle) {
	if (gs.VehId() >= 0) && (gs.VehId() >= len(*vehs)) {
		panic(fmt.Sprintf("PixelWorldViz.addGameShape() with VehdId=%d is invalid; game only has %d vehicles!", gs.VehId(), len(*vehs)))
	}
	if gs.shape >= numShapes {
		panic(fmt.Sprintf("PixelWorldViz.addGameShape: gs.shape=%v is invalid", gs.shape))
	}

	if (gs.VehId() >= 0) && gs.IsCartesian() {
		// shape's position is relative to vehicle's pose, in Cartesian coordinate space
		pose := trk.ToPose((*vehs)[gs.VehId()].CurTrackPose())
		pose1 := pose.AdvancePose(phys.Pose{Point: phys.Point{X: gs.x1, Y: gs.y1}, Theta: 0})
		pose2 := pose.AdvancePose(phys.Pose{Point: phys.Point{X: gs.x2, Y: gs.y2}, Theta: 0})

		p1 := phys.Point{X: pose1.X, Y: pose1.Y}
		p2 := phys.Point{X: pose2.X, Y: pose2.Y}
		switch gs.shape {
		case shapeLine:
			wv.pv.AddLine(p1, p2, gs.Thickness(), gs.Color())
		case shapeCirc:
			radius := phys.Dist(p1, p2)
			wv.pv.AddCircle(p1, radius, gs.Thickness(), gs.Color())
		}
	} else if (gs.VehId() >= 0) && !gs.IsCartesian() {
		// shape's position is relative to vehicle's pose, in Track coordinate space
		vtp := (*vehs)[gs.VehId()].CurTrackPose()
		tp1 := track.Pose{Point: track.Point{Dofs: gs.x1 + vtp.Dofs, Cofs: gs.y1 + vtp.Cofs}, DAngle: 0}
		tp2 := track.Pose{Point: track.Point{Dofs: gs.x2 + vtp.Dofs, Cofs: gs.y2 + vtp.Cofs}, DAngle: 0}
		tp1.Dofs = trk.NormalizeDofs(tp1.Dofs)
		tp2.Dofs = trk.NormalizeDofs(tp2.Dofs)
		pose1 := trk.ToPose(tp1)
		pose2 := trk.ToPose(tp2)
		p1 := phys.Point{X: pose1.X, Y: pose1.Y}
		p2 := phys.Point{X: pose2.X, Y: pose2.Y}
		switch gs.shape {
		case shapeLine:
			wv.addTrackDLine(trk, tp1.Cofs, tp1.Dofs, tp2.Dofs, gs.Thickness(), gs.Color())
			//wv.pv.AddLine(p1, p2, gs.Thickness(), gs.Color())
		case shapeCirc:
			radius := phys.Dist(p1, p2)
			wv.pv.AddCircle(p1, radius, gs.Thickness(), gs.Color())
		}
	} else {
		// shape's position is absolute
		var pose1 phys.Pose
		var pose2 phys.Pose
		if gs.IsCartesian() {
			// Cartesian coordinate space
			pose1 = phys.Pose{Point: phys.Point{X: gs.x1, Y: gs.y1}, Theta: 0}
			pose2 = phys.Pose{Point: phys.Point{X: gs.x2, Y: gs.y2}, Theta: 0}
		} else {
			// Track coordinate space
			pose1 = trk.ToPose(track.Pose{Point: track.Point{Dofs: gs.x1, Cofs: gs.y1}, DAngle: 0})
			pose2 = trk.ToPose(track.Pose{Point: track.Point{Dofs: gs.x2, Cofs: gs.y2}, DAngle: 0})
		}
		p1 := phys.Point{X: pose1.X, Y: pose1.Y}
		p2 := phys.Point{X: pose2.X, Y: pose2.Y}
		switch gs.shape {
		case shapeLine:
			wv.pv.AddLine(p1, p2, gs.Thickness(), gs.Color())
		case shapeCirc:
			radius := phys.Dist(p1, p2)
			wv.pv.AddCircle(p1, radius, gs.Thickness(), gs.Color())
		}
	}
}
