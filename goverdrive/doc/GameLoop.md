This document presents a high-level view of the seqeuence of game and
simulation "ticks" or "updates" that make goverdrive go.

```
RunGameLoop(gamePhase) {
  // No sim time passes in gamePhase.Start(). Game state can be
  // initialized, and vehicles can be repositioned, eg to simulate a
  // race lineup.
  gamePhase.Start(rsys)

  done = false
  while !done {
    // robo.System.Tick() updates the position and state of all vehicles,
    // based on their current state and commands from the game phase.
    // This is is the part of the program where sim time advances.
    // Auxiliary robotics tasks, such as updating the vehicle lights and
    // collision detection, also happen as part of the robotics simulation
    // tick.
    rsys.Tick()

    // gamePhase.Update() may do any of the following:
    //   - Process user input
    //   - Issue new vehicle commands
    //   - Update the game state
    // Update() returns all of the extra objects that need to visualized,
    // such as game shapes, track regions, and message board text.
    //
    // Worth noting:
    // - GamePhase.Update() may issue commands to vehicles, but it does 
    //   NOT directly move or change the state of the vehicles. Vehicle
    //   movement and state updates are performed by the robotics
    //   simulator.
    // - EXCEPTION: Vehicle.Reposition() DOES directly change the
    //   position of the vehicle.
    done = gamePhase.Update(...)

    // wait for remainder of real-time tick to finish

    // Window.Update() draws the track, vehicles, and other game 
    // objects to the display. It also gathers new input (eg from
    // the keyboard) for the next gaemPhase.Update().
    Window.Update()

    // Momentary Pause
    while IsPressed(Backspace (ie delete) key) {
      // no sim or game updates; only monitor key presses
    }
  }

  // No sim time passes in gamePhase.Stop(). Final game state and ranking
  // can be determined.
  gamePhase.Stop(rsys)
}
```
