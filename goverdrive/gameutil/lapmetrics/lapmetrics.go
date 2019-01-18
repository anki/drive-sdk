// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package lapmetrics provides track lap count, duration, path length, and other
// lap metrics for all vehicles involved in the game.
package lapmetrics

import (
	"fmt"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/track"
)

// CompletedLapInfo stores information about a completed lap.
type CompletedLapInfo struct {
	LapNumber   int
	LapTime     phys.SimTime
	IsTrackwise bool        // false => counter-trackwise
	PathLen     phys.Meters // actual driving path length
	MinDspd     phys.MetersPerSec
	MaxDspd     phys.MetersPerSec
}

// VehLapInfo stores completed laps and current lap info for one vehicle.
type VehLapInfo struct {
	curLapStartOdom    phys.Meters
	curLapStartTime    phys.SimTime
	curLapMinDspd      phys.MetersPerSec
	curLapMaxDspd      phys.MetersPerSec
	doneLaps           []CompletedLapInfo
	numNewReportedLaps int
}

func (cli *CompletedLapInfo) String() string {
	durSeconds := float64(cli.LapTime) / float64(phys.SimSecond)
	return fmt.Sprintf("LapNumber=%v, IsTrackwise=%v, LapTime=%.3f sec, PathLen=%.3f, MinDspd=%.3f, MaxDspd=%.3f",
		cli.LapNumber, cli.IsTrackwise, durSeconds, cli.PathLen, cli.MinDspd, cli.MaxDspd)
}

//////////////////////////////////////////////////////////////////////

// LapMetrics stores track VehLapInfo for all vehicles
type LapMetrics struct {
	recordTrackwiseLaps        bool
	recordCounterTrackwiseLaps bool
	info                       []VehLapInfo
}

// New returns a fresh LapMetrics object, which starts measuring from the
// current speed, odom, etc of the vehicles.
func New(now phys.SimTime, vehs *[]robo.Vehicle, recordTrackwiseLaps, recordCounterTrackwiseLaps bool) *LapMetrics {
	lm := LapMetrics{
		recordTrackwiseLaps:        recordTrackwiseLaps,
		recordCounterTrackwiseLaps: recordCounterTrackwiseLaps,
		info: make([]VehLapInfo, len(*vehs)),
	}
	for v, veh := range *vehs {
		lm.info[v] = VehLapInfo{
			curLapStartOdom: veh.Odom(),
			curLapStartTime: now,
			curLapMinDspd:   veh.CurDriveDspd(),
			curLapMaxDspd:   veh.CurDriveDspd(),
			doneLaps:        make([]CompletedLapInfo, 0),
		}
	}
	return &lm
}

// NumLapsCompleted returns the number laps that a vehicle has completed.
func (lm *LapMetrics) NumLapsCompleted(v int) int {
	return len(lm.info[v].doneLaps)
}

// AllCompletedLapInfo returns all completed lap info for a particular vehicle.
func (lm *LapMetrics) AllCompletedLapInfo(v int) []CompletedLapInfo {
	return lm.info[v].doneLaps
}

// NewCompletedLapInfo returns info about all newly completed laps, ie since the
// last call to NewCompletedLapInfo.
func (lm *LapMetrics) NewCompletedLapInfo(v int) []CompletedLapInfo {
	newLapInfo := make([]CompletedLapInfo, 0) // empty
	numCompl := lm.NumLapsCompleted(v)
	if lm.info[v].numNewReportedLaps < numCompl {
		newLapInfo = lm.info[v].doneLaps[lm.info[v].numNewReportedLaps:numCompl]
		lm.info[v].numNewReportedLaps = numCompl
	}
	return newLapInfo
}

// Update is the "tick" that should be called from the game phase's Update().
func (lm *LapMetrics) Update(now phys.SimTime, trk *track.Track, vehs *[]robo.Vehicle) {
	for v, veh := range *vehs {
		// Update current lap's min/max values
		curDspd := veh.CurDriveDspd()
		if curDspd < lm.info[v].curLapMinDspd {
			lm.info[v].curLapMinDspd = curDspd
		}
		if curDspd > lm.info[v].curLapMaxDspd {
			lm.info[v].curLapMaxDspd = curDspd
		}

		lapDist := veh.Odom() - lm.info[v].curLapStartOdom
		// TODO(gwenz): Review and tune lap thresholds
		if veh.CurDriveDofs() < 0.10 {
			// restart lap tracking
			if lapDist >= (0.7 * trk.CenLen()) {
				// succesfully completed a lap
				lapDist -= veh.CurDriveDofs()
				isTrackwise := veh.IsFacingTrackwise()
				if (isTrackwise && lm.recordTrackwiseLaps) ||
					(!isTrackwise && lm.recordCounterTrackwiseLaps) {
					newLap := CompletedLapInfo{
						LapNumber:   len(lm.info[v].doneLaps) + 1,
						LapTime:     now - lm.info[v].curLapStartTime,
						IsTrackwise: isTrackwise,
						PathLen:     lapDist,
						MinDspd:     lm.info[v].curLapMinDspd,
						MaxDspd:     lm.info[v].curLapMaxDspd,
					}
					lm.info[v].doneLaps = append(lm.info[v].doneLaps, newLap)
				}
			}
			lm.info[v].curLapStartOdom = veh.Odom() - veh.CurDriveDofs()
			lm.info[v].curLapStartTime = now
			lm.info[v].curLapMinDspd = curDspd
			lm.info[v].curLapMaxDspd = curDspd
		}
	}
}
