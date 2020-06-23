package main

import (
	"github.com/t-RED-69/games-with-go/RPG/UI2d"
	"github.com/t-RED-69/games-with-go/RPG/game"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	ui := &ui2d.UI2d{}
	game.Run(ui)
	sdl.Delay(10000)
}
