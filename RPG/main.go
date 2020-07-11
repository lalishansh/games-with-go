package main

import (
	ui2d "github.com/t-RED-69/games-with-go/RPG/UI2d"
	"github.com/t-RED-69/games-with-go/RPG/game"
)

func main() {
	//TODO when we need multiple UI support - refctor event-poll into its own component
	//and run it only on the main thread
	game := game.NewGame(1, "game/maps/level1.map")

	go func() { game.Run() }()

	ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	ui.Run()
}
