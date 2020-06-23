package game

import (
	//"github.com/veandco/go-sdl2/sdl"
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
	Void      Tile = ' '
	Player    Tile = 'P'
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
	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]
		var t Tile
		for x, e := range line {
			switch e {
			case ' ', '\t', '\n', '\r':
				t = Void
			case '#':
				t = StoneWall
			case '.':
				t = DirtFloor
			case '|':
				t = Door
			default:
				panic("undefined charecter in map")
			}
			level.Map[y][x] = t
		}

	}
	return level
}
func Run(ui GameUI) {
	level := loadLvlFromFile("level1.map")
	ui.Draw(level)
}
