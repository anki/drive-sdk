// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package phys

import (
	"fmt"
	"math"
)

// Point represents a point in traditional Cartesian space.
// X is the primary axis, Y is the secondary axis.
//   X>0 => right
//   X<0 => left
//   Y>0 => up
//   Y<0 => down
type Point struct {
	X Meters
	Y Meters
}

func (p Point) String() string {
	return fmt.Sprintf("Point{X: %v, Y: %v}", p.X, p.Y)
}

// Pose represents a location and an orientation.
//  Theta==0      => facing primary axis (eg +X, or right)
//  Theta==pi/2   => facing secondary axis (eg +Y, or up)
//  Theta==pi     => facing away from primary axis (eg -X, or left)
//  Theta==3*pi/2 => facing away from secondary axis (eg -Y, or down)
//  Theta==-pi/2  => same as (3*pi/2)
type Pose struct {
	Point
	Theta Radians
}

func (p Pose) String() string {
	return fmt.Sprintf("Pose{X: %v, Y: %v, Theta: %v}", p.X, p.Y, p.Theta)
}

// Dist returns the Cartesian distance between two points.
func Dist(p1, p2 Point) Meters {
	dx := (p1.X - p2.X)
	dy := (p1.Y - p2.Y)
	dist := math.Sqrt(float64((dx * dx) + (dy * dy)))
	return Meters(dist)
}

// AdvancePose advances by p2, starting from Pose p1.
// Specifically:
//   1. X/Y translation in the direction of p1.Theta
//   2. Rotation by p2.Theta
func (p1 Pose) AdvancePose(p2 Pose) Pose {
	// Assume p1 is at the origin, and move by p2's X,Y in p1.Theta direction
	pp := Point{X: p2.X, Y: p2.Y}.ToPolarPoint()
	pp.A = NormalizeRadians(p1.Theta + pp.A)
	p := pp.ToPoint()

	// Add original displacemet of p1
	// DEBUG: fmt.Printf("pp=%s\n p=%s\n p1=%s\n\n", pp.String(), p.String(), p1.String())
	p.X += p1.X
	p.Y += p1.Y

	// Compute orientation of final pose
	pose := Pose{Point: Point{X: p.X, Y: p.Y}, Theta: p1.Theta + p2.Theta}
	pose.Theta = NormalizeRadians(pose.Theta)
	return pose
}

// RelativeTo expresses pose p1 relative to pose p2 frame-of-reference.
func (p1 Pose) RelativeTo(p2 Pose) Pose {
	// translate point to origin
	xlatePoint := Point{X: p1.X - p2.X, Y: p1.Y - p2.Y}
	pp := xlatePoint.ToPolarPoint()

	// rotate about the origin
	pp.A = NormalizeRadians(pp.A - p2.Theta)
	p := pp.ToPoint()

	// correct the new pose angle
	pose := Pose{Point: p, Theta: p1.Theta - p2.Theta}
	pose.Theta = NormalizeRadians(pose.Theta)

	return pose
}

// PolarPoint is a polar representation of a point, ie radius + angle.
type PolarPoint struct {
	R Meters
	A Radians
}

func (pp PolarPoint) String() string {
	return fmt.Sprintf("PolarPoint{R: %v, A: %v}", pp.R, pp.A)
}

// ToPolarPoint converts a Cartesian Point to its PolarPoint representation.
func (p Point) ToPolarPoint() PolarPoint {
	r := Meters(math.Sqrt(float64((p.X * p.X) + (p.Y * p.Y))))
	a := Radians(math.Atan2(float64(p.Y), float64(p.X)))
	a = NormalizeRadians(a) // TODO: Is this the right thing to do?
	return PolarPoint{R: r, A: a}
}

// ToPoint converts a PolarPoint to its Cartesian Point representation.
func (pp PolarPoint) ToPoint() Point {
	// TODO: Is noramlizing the right thing to do?
	pp.A = NormalizeRadians(pp.A) // range [-Pi,+Pi]
	return Point{
		X: pp.R * Meters(math.Cos(float64(pp.A))),
		Y: pp.R * Meters(math.Sin(float64(pp.A))),
	}
}
