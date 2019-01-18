// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package viz

import (
	"image/color"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo/track"
)

const (
	// Supported GameShape types
	// Note that a thick line works as a rectangle. Boo-yeah!
	shapeLine = 0
	shapeCirc = 1
	numShapes = 2
)

// GameShape defines a flexible container for specifying primitive shapes used
// by the game. The meaning of some fields depends on specific shape.
//   - Shape coordinates can be absolute or relative to a particular vehicle
//   - Shape coordinates can be in Track or Cartesian coordinate space
type GameShape struct {
	vehId     int         // >= 0 means relative to that vehicle
	shape     uint        // eg ShapeLine
	isCartes  bool        // coordinate space: true => cartesian; false => track
	x1        phys.Meters // X or Dofs of point 1
	y1        phys.Meters // Y or Cofs of point 1
	x2        phys.Meters // X or Dofs of point 2 (may be unused, depending on Shape)
	y2        phys.Meters // Y or Cofs of point 2 (may be unused, depending on Shape)
	color     color.Color
	thickness phys.Meters // line thickness (0 => filled)
}

func NewCartesGameLine(vehId int, p1, p2 phys.Point, color color.Color, thickness phys.Meters) *GameShape {
	return &GameShape{
		vehId:     vehId,
		shape:     shapeLine,
		isCartes:  true,
		x1:        p1.X,
		y1:        p1.Y,
		x2:        p2.X,
		y2:        p2.Y,
		color:     color,
		thickness: thickness,
	}
}

func NewTrackGameLine(vehId int, tp1, tp2 track.Point, color color.Color, thickness phys.Meters) *GameShape {
	return &GameShape{
		vehId:     vehId,
		shape:     shapeLine,
		isCartes:  false,
		x1:        tp1.Dofs,
		y1:        tp1.Cofs,
		x2:        tp2.Dofs,
		y2:        tp2.Cofs,
		color:     color,
		thickness: thickness,
	}
}

func NewCartesGameCirc(vehId int, ctr phys.Point, rad phys.Meters, color color.Color, thickness phys.Meters) *GameShape {
	return &GameShape{
		vehId:     vehId,
		shape:     shapeCirc,
		isCartes:  true,
		x1:        ctr.X,
		y1:        ctr.Y,
		x2:        ctr.X + rad,
		y2:        ctr.Y,
		color:     color,
		thickness: thickness,
	}
}

func NewTrackGameCirc(vehId int, ctr track.Point, rad phys.Meters, color color.Color, thickness phys.Meters) *GameShape {
	return &GameShape{
		vehId:     vehId,
		shape:     shapeCirc,
		isCartes:  false,
		x1:        ctr.Dofs,
		y1:        ctr.Cofs,
		x2:        ctr.Dofs + rad,
		y2:        ctr.Cofs,
		color:     color,
		thickness: thickness,
	}
}

func (gs GameShape) VehId() int {
	return gs.vehId
}

func (gs GameShape) IsCartesian() bool {
	return gs.isCartes
}

func (gs GameShape) Color() color.Color {
	return gs.color
}

func (gs GameShape) Thickness() phys.Meters {
	return gs.thickness
}
