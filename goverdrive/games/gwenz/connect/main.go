// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

package main

import (
	"github.com/faiface/pixel/pixelgl"

	"github.com/anki/goverdrive/engine"
	"github.com/anki/goverdrive/robo"
	"github.com/anki/goverdrive/robo/light"
	"github.com/anki/goverdrive/viz"
)

func run() {
	// Configure standard parts of the game from command-line args
	gameConfig := engine.NewCLIGameConfig("Connect3 (goverdrive)", light.Gen2Spec)

	// Create the remaining game components
	primViz := viz.NewPixelViz()
	worldViz := viz.NewPixelWorldViz(primViz, gameConfig.Track())
	rsim := robo.NewIdealSimulator()
	rcollide := robo.NewCollisionDetector(gameConfig.Track(), gameConfig.Vehicles())
	roboSys := robo.NewSystem(gameConfig.Track(), gameConfig.Vehicles(), rsim, rcollide)

	// Run the game
	vizCfg := engine.GamePhaseVizConfig{
		ShowInstr:         gameConfig.ShowInstructions(),
		MsgBoardPixHeight: gameConfig.MsgBoardPixHeight(),
		WorldViz:          worldViz,
		Window:            gameConfig.Window(),
	}
	engine.RunGameLoop(vizCfg, roboSys, &ConnectGamePhase{})
}

func main() {
	pixelgl.Run(run)
}
