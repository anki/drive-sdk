// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com
//
// Detect vehicle collisions. There may or may not be a reaction.
//
// TODO(gwenz): Should the robotics system natively support collisions between a
// vehicle and a non-vehicle object, eg a road obstacle that is part of the
// game?

package robo

import (
	_ "fmt"
	"math"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo/track"
)

// VehicleCollider monitors vehicles to detect collisions, and possibly change
// vehicle state when a collision happens.
type VehicleCollider interface {
	// NewCollisions returns all of the vehicle collisions that have happened
	// since the last call to NewCollisions. In most cases, this will be a very
	// small list (or empty list). There is no order guarantee for returned
	// CollisionEvents. NOTE: If called irregularly, some of the older "new"
	// collisions may be forgotten
	NewCollisions() []CollisionEvent

	// CurCollisions returns all collision events that are ongoing. There is no
	// order guarantee for returned CollisionEvents.
	CurCollisions() []CollisionEvent

	// Update updates the collision and vehicle states, based on position of each
	// vehicle. Should be called by the robotics system.
	update(now phys.SimTime, trk *track.Track, vehs *[]Vehicle)
}

// VehicleCollisionInfo captures the collision info for one of the two vehicles
// involved. The POI (point-of-impact) is in that vehicle's frame of reference.
type VehicleCollisionInfo struct {
	Id  int
	POI phys.Point
}

// CollisionEvent captures all of the information about a vehicle's collision
// with another vehicle, at the moment of impact.
type CollisionEvent struct {
	ImpactTime phys.SimTime
	VehInfo    [2]VehicleCollisionInfo
}

// XXX(gwenz): Angle boundaries for high-level colllision direction are
// simplistic. May want to take each vehicle's actual length and width into
// account.

func (vci VehicleCollisionInfo) IsFrontCollision() bool {
	angle := vci.POI.ToPolarPoint().A
	return (angle > (-math.Pi / 4)) && (angle < (+math.Pi / 4))
}

func (vci VehicleCollisionInfo) IsRearCollision() bool {
	angle := vci.POI.ToPolarPoint().A
	return (angle > (3 * math.Pi / 4)) || (angle < (-4 * math.Pi / 4))
}

func (vci VehicleCollisionInfo) IsLeftSideCollision() bool {
	angle := vci.POI.ToPolarPoint().A
	return (angle >= (math.Pi / 4)) && (angle <= (3 * math.Pi / 4))
}

func (vci VehicleCollisionInfo) IsRightSideCollision() bool {
	angle := vci.POI.ToPolarPoint().A
	return (angle >= (-3 * math.Pi / 4)) && (angle <= (-math.Pi / 4))
}

//////////////////////////////////////////////////////////////////////

// CollisionDetector is a simple detector of vehicle collisions, based on
// rectangular vehicle gemoetry. It does not modify vehicle state when a
// collision happens.
type CollisionDetector struct {
	maxDimension  map[vehPair]phys.Meters
	curCollisions map[vehPair]CollisionEvent
	newCollisions map[vehPair]CollisionEvent
}

type vehPair struct {
	Veh1, Veh2 int
}

// NewCollisionDetector creates a new detector suited for the specific trk and
// set of vehicles.
func NewCollisionDetector(trk *track.Track, vehs *[]Vehicle) *CollisionDetector {
	maxDimension := make(map[vehPair]phys.Meters)
	for v1 := range *vehs {
		for v2 := v1 + 1; v2 < len(*vehs); v2++ {
			veh1 := (*vehs)[v1]
			veh2 := (*vehs)[v2]
			lmax := math.Max(float64(veh1.Length()), float64(veh2.Length()))
			wmax := math.Max(float64(veh1.Width()), float64(veh2.Width()))
			max := math.Max(lmax, wmax)
			maxDimension[vehPair{v1, v2}] = phys.Meters(max)
		}
	}

	return &CollisionDetector{
		maxDimension:  maxDimension,
		curCollisions: make(map[vehPair]CollisionEvent),
		newCollisions: make(map[vehPair]CollisionEvent),
	}
}

func (cd *CollisionDetector) NewCollisions() []CollisionEvent {
	events := make([]CollisionEvent, 0)
	for _, ce := range cd.newCollisions {
		events = append(events, ce)
	}
	// once reported, the collisions aren't "new" anymore
	cd.newCollisions = make(map[vehPair]CollisionEvent)
	return events
}

func (cd *CollisionDetector) CurCollisions() []CollisionEvent {
	events := make([]CollisionEvent, 0)
	for _, ce := range cd.curCollisions {
		events = append(events, ce)
	}
	return events
}

func (cd *CollisionDetector) update(now phys.SimTime, trk *track.Track, vehs *[]Vehicle) {
	// populate collision inputs, for helper function
	inputs := make([]vehCollisionInputs, len(*vehs))
	for i, veh := range *vehs {
		inputs[i] = vehCollisionInputs{
			dofs:  veh.CurTrackPose().Dofs,
			pose:  trk.ToPose(veh.CurTrackPose()),
			len:   veh.Length(),
			width: veh.Width(),
		}
	}

	cd.updateHelper(now, trk, inputs)
}

//////////////////////////////////////////////////////////////////////

// vehCollisionInputs is an intermediate type that holds all of the per-vehicle
// information needed to detect vehicles collisions. This is intended to help
// unit test the collision indexing and math without having to create a track
// and set of vehicles and then carefully manipulate their state.
type vehCollisionInputs struct {
	dofs  phys.Meters // Track
	pose  phys.Pose   // Cartesian
	len   phys.Meters
	width phys.Meters
}

func (cd *CollisionDetector) updateHelper(now phys.SimTime, trk *track.Track, allInputs []vehCollisionInputs) {
	for v0 := range allInputs {
		for v1 := v0 + 1; v1 < len(allInputs); v1++ {
			pair := vehPair{v0, v1}

			// Track pieces can overlap in 2D space, ie very different Dofs values can
			// map to same Cartesian coordinates, such as an overpass. In this case,
			// the vehicles are NOT colliding.
			maxDim := cd.maxDimension[pair]
			if trk.DofsDist(allInputs[v0].dofs, allInputs[v1].dofs) > maxDim {
				delete(cd.curCollisions, pair)
				continue
			}

			// vehicles are close => need to do the collision math
			poiInputs := [2]vehCollisionInputs{allInputs[v0], allInputs[v1]}
			isCollision, absPOI := calcPointOfImpact(poiInputs)
			if isCollision {
				if _, ok := cd.curCollisions[pair]; !ok {
					// Convert absolute Cartesian point into vehicle-relative point for
					// each vehicle
					var vehInfo [2]VehicleCollisionInfo
					impactPose := phys.Pose{Point: absPOI, Theta: 0}
					vehInfo[0].POI = impactPose.RelativeTo(allInputs[v0].pose).Point
					vehInfo[1].POI = impactPose.RelativeTo(allInputs[v1].pose).Point
					vehInfo[0].Id = v0
					vehInfo[1].Id = v1

					newEvent := CollisionEvent{
						ImpactTime: now,
						VehInfo:    vehInfo,
					}
					cd.curCollisions[pair] = newEvent
					cd.newCollisions[pair] = newEvent
					// NOTE: ^^^ will quietly replace any existing "newCollision" for the pair
				}
				// For non-new collisions, do NOT update curCollisions, to preserve the
				// initial time of impact.
			} else {
				// not colliding at this moment
				delete(cd.curCollisions, pair)
				continue
			}
		}
	}
}

// calcPointOfImpact determines if two vehicles are colliding, based on their
// physical position and dimensions. If they are colliding, a point-of-impact is
// calculated (absolute Cartesian coordinate space).
//   - Not colliding => returns false with invalid phys.Point
//   -     Colliding => returns true  with   valid phys.Point
func calcPointOfImpact(inputs [2]vehCollisionInputs) (bool, phys.Point) {
	// Collision detect algorithm:
	// - A vehicles is modeled as a rectangle
	// - Check if any of the four corners of one vehicle is inside the other vehicle

	collisionPoints := make([]phys.Point, 0)
	for rv := 0; rv < 2; rv++ { // rv = index of the "Reference" vehicle
		ov := (rv + 1) % 2 //        ov = index of the "Other"     vehicle

		// Abs = calculate the Other vehicle's four corner points, in absolute
		// Cartesian frame of reference
		ovHalfLen := inputs[ov].len / 2
		ovHalfWid := inputs[ov].width / 2
		ovCornersAbs := []phys.Point{
			inputs[ov].pose.AdvancePose(phys.Pose{Point: phys.Point{X: +ovHalfLen, Y: +ovHalfWid}, Theta: 0}).Point, // front L
			inputs[ov].pose.AdvancePose(phys.Pose{Point: phys.Point{X: +ovHalfLen, Y: -ovHalfWid}, Theta: 0}).Point, // front R
			inputs[ov].pose.AdvancePose(phys.Pose{Point: phys.Point{X: -ovHalfLen, Y: +ovHalfWid}, Theta: 0}).Point, // back  L
			inputs[ov].pose.AdvancePose(phys.Pose{Point: phys.Point{X: -ovHalfLen, Y: -ovHalfWid}, Theta: 0}).Point, // back  R
		}
		// for _, corner := range ovCornersAbs {
		// 	fmt.Printf("ov=%v => ovCornersAbs=%s\n", ov, corner.String())
		// }

		// Rel = calculate the Other vehicle's four corner points, in Cartesian
		// frame of reference relative to the Reference vehicle
		ovCornersRel := make([]phys.Point, 4)
		for i, cp := range ovCornersAbs {
			cpose := phys.Pose{Point: cp, Theta: 0}
			ovCornersRel[i] = cpose.RelativeTo(inputs[rv].pose).Point
		}
		// for _, corner := range ovCornersRel {
		// 	fmt.Printf("ov=%v => ovCornersRel=%s\n", ov, corner.String())
		// }

		// Determine which of the Other vehicle's four corners are inside the
		// Reference vehicle's rectangle
		rvHalfLen := inputs[rv].len / 2
		rvHalfWid := inputs[rv].width / 2
		// xstr := fmt.Sprintf("x = [%v %v %v %v %v %v %v %v]", rvHalfLen, rvHalfLen, -rvHalfLen, -rvHalfLen, ovCornersRel[0].X, ovCornersRel[1].X, ovCornersRel[2].X, ovCornersRel[3].X)
		// ystr := fmt.Sprintf("y = [%v %v %v %v %v %v %v %v]", rvHalfWid, -rvHalfWid, rvHalfWid, -rvHalfWid, ovCornersRel[0].Y, ovCornersRel[1].Y, ovCornersRel[2].Y, ovCornersRel[3].Y)
		// fmt.Printf("%s\n%s\n", xstr, ystr)  // XXX: quick-and-dirty for Matlab display
		for i, point := range ovCornersRel {
			if (point.X > rvHalfLen) || (point.X < -rvHalfLen) ||
				(point.Y > +rvHalfWid) || (point.Y < -rvHalfWid) {
				continue
			}
			// Note: record the Abs collision point, not Rel
			collisionPoints = append(collisionPoints, ovCornersAbs[i])
			//fmt.Printf("ov=%v, corner=%v, collisionPoint=%v\n", ov, i, ovCornersAbs[i])
		}
	}

	if len(collisionPoints) == 0 {
		return false, phys.Point{X: 0, Y: 0}
	}

	// There may be >1 collision point. If so, the "net" collision point (absolute
	// Cartesian space) applies to both vehicles and is simply the average of all
	// detected collision points. This is not a perfect answer, but is
	// straightforward and should be good enough.
	collisionPoint := phys.Point{X: 0, Y: 0}
	for _, cp := range collisionPoints {
		collisionPoint.X += cp.X
		collisionPoint.Y += cp.Y
	}
	collisionPoint.X /= phys.Meters(len(collisionPoints))
	collisionPoint.Y /= phys.Meters(len(collisionPoints))

	return true, collisionPoint
}
