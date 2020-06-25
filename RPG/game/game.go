package game

import (
	"fmt"
	"time"
	//"github.com/veandco/go-sdl2/sdl"
	"bufio"
	"math"
	"os"
	"sort"
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
	Search
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
	Debug  map[Pos]bool
}
type Entity struct {
	Symbol Tile
	Pos
}
type Pos struct {
	X, Y int32
}
type priorityPos struct {
	Pos
	priority int
}
type priorityArray []priorityPos

func (a priorityArray) Len() int           { return len(a) }
func (a priorityArray) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a priorityArray) Less(i, j int) bool { return a[i].priority < a[j].priority }

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
func getNeighbours(level *Level, pos Pos) []Pos {
	neghbours := make([]Pos, 0, 4)
	//
	//left := Pos{pos.X - 16, pos.Y + 16}
	//right := Pos{pos.X + 48, pos.Y + 16}
	//up := Pos{pos.X + 16, pos.Y - 16}
	//down := Pos{pos.X + 16, pos.Y + 48}
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}
	//
	if canSearch(level, right) {
		neghbours = append(neghbours, right)
	}
	if canSearch(level, left) {
		neghbours = append(neghbours, left)
	}
	if canSearch(level, up) {
		neghbours = append(neghbours, up)
	}
	if canSearch(level, down) {
		neghbours = append(neghbours, down)
	}

	return neghbours
}
func canSearch(level *Level, pos Pos) bool {
	return (level.Map[int(pos.Y)][int(pos.X)] == DirtFloor || level.Map[int(pos.Y)][int(pos.X)] == OpenDoor)
}
func bfs(ui GameUI, level *Level, start Pos) {
	start = Pos{int32(start.X / 32), int32(start.Y / 32)}
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true
	level.Debug = visited
	for len(frontier) > 0 {
		current := frontier[0]
		frontier = frontier[1:]
		for _, nxt := range getNeighbours(level, current) {
			if !visited[nxt] {
				frontier = append(frontier, nxt)
				visited[nxt] = true
				ui.Draw(level)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func astar(ui GameUI, lvl *Level, start Pos, goal Pos) []Pos {
	start = Pos{int32(start.X / 32), int32(start.Y / 32)}
	goal = Pos{int32(goal.X / 32), int32(goal.Y / 32)}
	frontier := make(priorityArray, 0, 8)
	frontier = append(frontier, priorityPos{start, 1})
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0
	lvl.Debug = make(map[Pos]bool)
	for len(frontier) > 0 {
		sort.Stable(frontier) //slow priority queue,make a real one
		current := frontier[0]
		if current.Pos == goal {
			path := make([]Pos, 0)
			p := current.Pos
			for p != start {
				fmt.Println(p)
				path = append(path, p)
				p = cameFrom[p]
			}
			path = append(path, p)
			fmt.Println("done!")
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			for _, pos := range path {
				lvl.Debug[pos] = true
				ui.Draw(lvl)
				time.Sleep(100 * time.Millisecond)
			}
			return path
		}
		frontier = frontier[1:]
		for _, nxt := range getNeighbours(lvl, current.Pos) {
			newCost := costSoFar[current.Pos] + 1 //always 1 for now
			_, exists := costSoFar[nxt]
			if !exists || newCost < costSoFar[nxt] {
				costSoFar[nxt] = newCost
				xDist := int(math.Abs(float64(goal.X - nxt.X)))
				yDist := int(math.Abs(float64(goal.Y - nxt.Y)))
				priority := newCost + xDist + yDist
				frontier = append(frontier, priorityPos{nxt, priority})
				cameFrom[nxt] = current.Pos
			}
		}
	}
	return nil
}
func Run(ui GameUI) {
	fmt.Println("Starting...")
	level := loadLvlFromFile("level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()
		p := level.Player
		if (*input).Typ == Quit {
			return
		} else if (*input).Typ == Up && (level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == DirtFloor || level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == OpenDoor) {
			level.Player.Y -= 3
		} else if (*input).Typ == Down && (level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == DirtFloor || level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == OpenDoor) {
			level.Player.Y += 3
		} else if (*input).Typ == Right && (level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == DirtFloor || level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == OpenDoor) {
			level.Player.X += 3
		} else if (*input).Typ == Left && (level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == DirtFloor || level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == OpenDoor) {
			level.Player.X -= 3
		} else {
			LevelManager(level, input, ui)
		}
	}
}
func LevelManager(level *Level, input *Input, ui GameUI) {
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
	} else if input.Typ == Search {
		//bfs(ui, level, level.Player.Pos)
		_ = astar(ui, level, level.Player.Pos, Pos{278, 200})
	}
}
