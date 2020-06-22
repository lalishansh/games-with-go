package game

import (
	"bufio"
	"os"
)

type GameUI interface {
	Draw(*Level)
}

type Tile rune

const (
	StoneWall Tile = '#'
	DirtFloor Tile = '.'
	Door      Tile = '|'
)

//Level infor. for game
type Level struct {
	Map [][]Tile
}

//loadLvlFromFile will try to load level provided from "game/maps/" folder
func loadLvlFromFile(filNam string) (lvl *Level) {
	file, err := os.Open("game/maps/" + filNam)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	levelLines := make([]string, 0)
	longestRow := 0
	index := 0
	for scanner.Scan() {
		levelLines = append(levelLines, scanner.Text())
		if len(levelLines[index]) > longestRow {
			longestRow = len(levelLines[index])
		}
		index++
	}
	level := &Level{}
	level.Map = make([][]Tile, len(levelLines))
	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow)
	}
	return level
}
func Run(ui GameUI) {
	level := loadLvlFromFile("level1.map")
	ui.Draw(level)
}
