package game

import (
	//"github.com/veandco/go-sdl2/sdl"
	"bufio"
	"os"
)

type GameUI interface {
	Draw(*Level)
	GetInput() *Input
}
type InputType int

const (
	Quit InputType = iota
	Up
	Down
	Left
	Right
	Open
	Blank
)

type Input struct {
	Typ InputType
}

type Tile rune

const (
	StoneWall Tile = '#'
	DirtFloor Tile = '.'
	CloseDoor Tile = '|'
	OpenDoor  Tile = '/'
	Void      Tile = ' '
	Player    Tile = 'P'
)

//Level infor. for game
type Level struct {
	Map    [][]Tile
	Player Entity
}
type Entity struct {
	Symbol Tile
	X, Y   int32
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
				t = CloseDoor
			case '/':
				t = CloseDoor
			default:
				panic("undefined charecter in map")
			}
			level.Map[y][x] = t
		}
	}
	level.Player.X, level.Player.Y, level.Player.Symbol = 84, 84, 'P'
	return level
}
func Run(ui GameUI) {
	level := loadLvlFromFile("level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()
		p := level.Player
		if (*input).Typ == Quit {
			return
		} else if (*input).Typ == Up && (level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == DirtFloor || level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == OpenDoor) {
			level.Player.Y--
		} else if (*input).Typ == Down && (level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == DirtFloor || level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == OpenDoor) {
			level.Player.Y++
		} else if (*input).Typ == Right && (level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == DirtFloor || level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == OpenDoor) {
			level.Player.X++
		} else if (*input).Typ == Left && (level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == DirtFloor || level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == OpenDoor) {
			level.Player.X--
		} else {
			LevelManager(level, input)
		}
	}
}
func LevelManager(level *Level, input *Input) {
	p := level.Player
	if input.Typ == Open {
		if level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == CloseDoor {
			level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] = OpenDoor
		} else if level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == CloseDoor {
			level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] = OpenDoor
		} else if level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == CloseDoor {
			level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] = OpenDoor
		} else if level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == CloseDoor {
			level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] = OpenDoor
		}
	}
}
