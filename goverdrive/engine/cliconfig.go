// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com
//
// cliconfig.go provides a means to configure many parts of the game (track,
// vehicles, etc) in using standard command-line arguments. Unless there is a
// very good reason not to, this is THE way to configure these parts of the
// game.

package engine

import (
	"flag"
	"fmt"
	"strings"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/robo/track"
)

// CLIGameConfig is the game's configuration, based on command-line values
type CLIGameConfig struct {
	trk       *track.Track
	vehs      []robo.Vehicle
	win       *pixelgl.Window
	mbHeight  uint
	showInstr bool
}

// NewCLIGameConfig parses command-line arguments and creates a game
// configuration based on their values.
func NewCLIGameConfig(title string, lightSpec light.Spec) *CLIGameConfig {
	var gc CLIGameConfig

	winFlag /*******/ := flag.String("w", "1200x850", "Window size, expressed as integer pixels WIDTHxHEIGHT")
	mbFlag /********/ := flag.Uint("mb", 200, "Message board height, expressed as integer number of pixels. Can be 0.")
	tWidthFlag /****/ := flag.Float64("twidth", 0.20, "Track width, in Meters")
	tMaxCofsFlag /**/ := flag.Float64("tmaxcofs", 0.0, "Track max center offset, from road center")
	trackFlag /*****/ := flag.String("t", "Capsule", "Track name or modular track string")
	vehsFlag /******/ := flag.String("v", "gs", "List of vehicles, using two-letter abberviations; eg \"gs sk\" for Groundshock and Skull")
	insFlag /*******/ := flag.Bool("ins", false, "Display instructions at the start of each game phase")
	flag.Parse()

	// parse the window size
	var winWidth, winHeight uint
	if n, werr := fmt.Sscanf(*winFlag, "%dx%d", &winWidth, &winHeight); (werr != nil) || (n != 2) {
		panic(fmt.Sprintf("win=\"%s\" could not be parsed as WxH pixels", *winFlag))
	}

	// message board height
	gc.mbHeight = *mbFlag
	if gc.mbHeight > (winHeight / 2) {
		panic(fmt.Sprintf("Message board height=%v is too big, relative to window height=%v", gc.mbHeight, winHeight))
	}

	// track width
	twidth := phys.Meters(*tWidthFlag)
	if (twidth < 0.001) || (twidth > 2.0) {
		panic(fmt.Sprintf("Track width=%v is not reasonable", twidth))
	}

	// game instructions
	gc.showInstr = *insFlag

	// create the track
	tMaxCofs := phys.Meters(*tMaxCofsFlag)
	if *trackFlag != "" {
		gc.trk, _ = track.NewModularTrack(twidth, tMaxCofs, *trackFlag)
		if gc.trk == nil {
			gc.trk, _ = track.NewStarterKitTrack(twidth, tMaxCofs, *trackFlag)
		}
		if gc.trk == nil {
			gc.trk, _ = track.NewCustomTrack(twidth, tMaxCofs, *trackFlag)
		}
	}
	if gc.trk == nil {
		fmt.Printf("Supported starter kit tracks:\n  %s\n", track.StarterKitTrackNames("\n  "))
		fmt.Printf("Supported custom tracks:\n  %s\n", track.CustomTrackNames("\n  "))
		panic("A valid track is required to proceed!")
	}

	// create the vehicles
	vehStrList := strings.Split(*vehsFlag, " ")
	gc.vehs = make([]robo.Vehicle, 0)
	for _, vs := range vehStrList {
		gc.vehs = append(gc.vehs, *robo.NewVehicle(robo.VehType(vs), lightSpec, gc.trk.CenLen()))
	}

	// create the window
	winCfg := pixelgl.WindowConfig{
		Title:  title,
		Bounds: pixel.R(0, 0, float64(winWidth), float64(winHeight)),
		VSync:  true,
	}
	var werr error
	gc.win, werr = pixelgl.NewWindow(winCfg)
	if werr != nil {
		panic(werr)
	}

	return &gc
}

// Track returns a pointer to the track that was created
func (gc *CLIGameConfig) Track() *track.Track {
	return gc.trk
}

// Vehicles returns a pointer to the vehicles that were created
func (gc *CLIGameConfig) Vehicles() *[]robo.Vehicle {
	return &gc.vehs
}

// Window returns a pointer to the window that was created
func (gc *CLIGameConfig) Window() *pixelgl.Window {
	return gc.win
}

// MsgBoardPixHeight returns the number of vertical pixels that should be
// dedicated to the message board.
func (gc *CLIGameConfig) MsgBoardPixHeight() uint {
	return gc.mbHeight
}

// ShowInstructions returns true if instructions should be displayed before the
// start of each game phase.
func (gc *CLIGameConfig) ShowInstructions() bool {
	return gc.showInstr
}
