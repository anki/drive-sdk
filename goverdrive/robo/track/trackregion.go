// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package track

import (
	"fmt"

	"github.com/anki/goverdrive/phys"
)

// Region defines a "rectangular" sub-region of the track. If the
// definition includes areas of curvature, the region bends to match the shape
// of the track.
//   - The start corner's Dofs must satisfy (0 <= Dofs <= track.CenLen())
//   - The length of the region can be >track.CenLen() => covers whole track
//   - Cofs can extend beyond the width of the track
//   - Regions always extend in the forward driving direction from first corner,
//     and always in the +Cofs direction
type Region struct {
	c1    Point // start corner
	len   phys.Meters
	width phys.Meters
	track *Track
}

func (tr *Region) String() string {
	return fmt.Sprintf("Region{c1: %v, len: %v, width: %v}", tr.c1, tr.len, tr.width)
}

// NewRegion creates a track region by specifying its start corner, length,
// and width.
func NewRegion(track *Track, c1 Point, len, width phys.Meters) *Region {
	// check input values
	if c1.Dofs < 0 {
		panic(fmt.Sprintf("NewRegion: c1.Dofs=%v invalid; must be >= 0", c1.Dofs))
	}
	if c1.Dofs >= track.CenLen() {
		panic(fmt.Sprintf("NewRegion: c1.Dofs=%v invalid; must be < track.CenLen()=%v", c1.Dofs, track.CenLen()))
	}
	if width <= 0 {
		panic(fmt.Sprintf("NewRegion: width=%v invalid; must be >0", width))
	}
	if len <= 0 {
		panic(fmt.Sprintf("NewRegion: len=%v invalid; must be >0", len))
	}

	return &Region{
		c1:    c1,
		len:   len,
		width: width,
		track: track,
	}
}

// C1 returns the "start" corner of the track region.
func (tr *Region) C1() Point {
	return tr.c1
}

// C2 returns the "end" corner of the track region.
func (tr *Region) C2() Point {
	c2 := tr.c1
	c2.Cofs += tr.width
	c2.Dofs = tr.track.NormalizeDofs(c2.Dofs + tr.len)
	return c2
}

// Width returns the width of the track region.
func (tr *Region) Width() phys.Meters {
	return tr.width
}

// Len returns the length of the track region.
func (tr *Region) Len() phys.Meters {
	return tr.len
}

// CrossesFinishLine returns true if the track region crosses the finish line.
func (tr *Region) CrossesFinishLine() bool {
	return (tr.c1.Dofs + tr.len) >= tr.track.CenLen()
}

// ContainsPoint returns true if a track point is contained inside the track
// region. Note that corner C1 is included in the region, but C2 is not. In
// other words, the rectangular track region is [C1, C2).
func (tr *Region) ContainsPoint(p Point) bool {
	// center offset
	//        vvv Inclusive            vvvv Exclusive
	if (p.Cofs < tr.c1.Cofs) || (p.Cofs >= (tr.c1.Cofs + tr.width)) {
		return false
	}

	// distance offset
	p.Dofs = tr.track.NormalizeDofs(p.Dofs)
	if tr.CrossesFinishLine() {
		if p.Dofs >= tr.c1.Dofs {
			return true
		}
		c2Dofs := tr.track.NormalizeDofs(tr.c1.Dofs + tr.len)
		if p.Dofs < c2Dofs {
			return true
		}
		return false
	} else {
		// region does not cross finish line
		//        vvv Inclusive            vvvv Exclusive
		if (p.Dofs < tr.c1.Dofs) || (p.Dofs >= (tr.c1.Dofs + tr.len)) {
			return false
		}
	}

	return true
}
