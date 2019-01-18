// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package engine has the reusable game engine core, such as the game loop.
package engine

import (
	"fmt"
	"golang.org/x/image/colornames"
	"math"
	"sort"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"

	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/viz"
)

const (
	// TODO(gwenz): It's unclear what range of values works for
	// roboTicksPerGameTick, and whether there is any need for a game developer to
	// change it.
	roboTicksPerGameTick uint = 2

	mbPaddingPixX = 20
	mbPaddingPixY = 40
)

// GamePhaseVizConfig is a wrapper for all of the visualization configuration,
// such as the window, and how much of the window is occupied by the Message
// Board.
type GamePhaseVizConfig struct {
	ShowInstr         bool
	MsgBoardPixHeight uint // pixels
	WorldViz          viz.WorldViz
	Window            *pixelgl.Window
	atlas             *text.Atlas
}

// RunGameLoop is the core loop that drives the game. It runs one game phase
// from start to finish, with the supplied visualization config and robotics
// system. To run without vizualization or UI, set the window to nil.
//
// RunGameLoop includes:
//   - Robotics simulation
//   - User input
//   - Rendering the world, and displaying it to a window
func RunGameLoop(vizCfg GamePhaseVizConfig, rsys *robo.System, phase GamePhase) {
	fmt.Printf("track.CenLen()=%v, track.MinCorner()=%v, track.MaxCorner=%v\n", rsys.Track.CenLen(), rsys.Track.MinCorner(), rsys.Track.MaxCorner())
	//fmt.Printf("winBounds.Min=%v, winBounds.Max=%v\n", vizCfg.Window.Bounds().Min, vizCfg.Window.Bounds().Max)

	vizCfg.atlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
	vizCfg.Window.SetSmooth(true) // less pixelated rendering

	phase.Start(rsys)

	if vizCfg.ShowInstr {
		// before starting the game, display instructions on the message board
		vizObj := EmptyGamePhaseVizObjects()
		vizObj.MBText = phase.InstructionText(rsys) + "\n<<<Press SPACE BAR to continue>>>"
		drawToWindow(vizCfg, rsys, vizObj)
		fps := time.Tick(time.Second / 20)
		for !vizCfg.Window.JustReleased(pixelgl.KeySpace) {
			vizCfg.Window.Update()
			<-fps
		}
	}

	var gameDeltaT time.Duration = time.Duration(uint64(roboTicksPerGameTick)*uint64(rsys.SimDeltaT())) * time.Nanosecond
	gameDelay := time.After(gameDeltaT)
	done := false
	for !done && !vizCfg.Window.Closed() {
		// Robotics simulation
		for i := uint(0); i < roboTicksPerGameTick; i++ {
			rsys.Tick()
		}

		// Game logic
		isDone, vizObj := phase.Update(rsys, vizCfg.Window)
		done = isDone

		// Display and inputs
		if vizCfg.Window != nil {
			<-gameDelay
			gameDelay = time.After(gameDeltaT)
			drawToWindow(vizCfg, rsys, vizObj)
			vizCfg.Window.Update() // display and inputs

			// momentary pause (while key is pressed)
			if vizCfg.Window.JustPressed(pixelgl.KeyBackspace) {
				fps := time.Tick(time.Second / 20)
				for !vizCfg.Window.JustReleased(pixelgl.KeyBackspace) {
					vizCfg.Window.Update()
					<-fps
				}
			}
		}
	}

	phase.Stop(rsys)

	if done {
		// Show the final vehicle ranking on the Message Board

		vizObj := EmptyGamePhaseVizObjects()
		rstr := ""
		rankings := VehRankingSorter{phase.VehRankings()}
		sort.Sort(&rankings) // in-place
		for _, r := range rankings.Rankings {
			rstr += fmt.Sprintf("[%s] %s\n", rsys.Vehicles[r.VehId].Type(), r.String())
		}
		vizObj.MBText = rstr + "\nDONE. Press SPACE BAR to continue.."
		drawToWindow(vizCfg, rsys, vizObj)
		fps := time.Tick(time.Second / 20)
		for !vizCfg.Window.JustReleased(pixelgl.KeySpace) {
			vizCfg.Window.Update()
			<-fps
		}
	}
}

func drawToWindow(vizCfg GamePhaseVizConfig, rsys *robo.System, vizObj GamePhaseVizObjects) {
	canvas := vizCfg.WorldViz.RenderAll(&rsys.Track, vizObj.Regions, &rsys.Vehicles, vizObj.Shapes)

	// TODO(gwenz): Encapsulate window/canvas/text/etc into package viz, so
	// that gameloop does not directly depend on visualization implementation?

	// stretch the canvas to fit the window
	// (leave room at the bottom for the message board)
	scaleFactor := math.Min(
		vizCfg.Window.Bounds().W()/canvas.Bounds().W(),
		(vizCfg.Window.Bounds().H()-float64(vizCfg.MsgBoardPixHeight))/canvas.Bounds().H())

	winBounds := vizCfg.Window.Bounds()
	winBounds.Max.Y += float64(vizCfg.MsgBoardPixHeight)
	vizCfg.Window.Clear(colornames.Black)
	vizCfg.Window.SetMatrix(pixel.IM.Scaled(pixel.ZV, scaleFactor).Moved(winBounds.Center()))
	canvas.Draw(vizCfg.Window, pixel.IM)

	mbPos := pixel.Vec{
		X: mbPaddingPixX - (winBounds.Center().X / scaleFactor),
		Y: (-winBounds.Center().Y+float64(vizCfg.MsgBoardPixHeight))/scaleFactor - mbPaddingPixY,
	}
	txt := text.New(pixel.V(0, 0), vizCfg.atlas)
	txt.Color = colornames.Lightgrey
	txt.WriteString(vizObj.MBText)
	txt.Draw(vizCfg.Window, pixel.IM.Scaled(pixel.ZV, 1.4/scaleFactor).Moved(mbPos))
}
