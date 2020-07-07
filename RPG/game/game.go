package game

import (
	"fmt"
	"time"

	//"github.com/veandco/go-sdl2/sdl"
	"bufio"
	"math"
	"os"
)

type Game struct {
	LevelChans []chan *Level
	InputChan  chan *Input
	Level      *Level
}

func NewGame(numWindows int, levelPath string) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}
	inputChan := make(chan *Input)
	return &Game{levelChans, inputChan, loadLvlFromFile(levelPath)}
}

type InputType int

const (
	Blank InputType = iota
	Up
	Down
	Left
	Right
	Open
	QuitGame
	CloseWindow
	Search
)

type Input struct {
	Typ          InputType
	LevelChannel chan *Level
}

type Tile rune

const (
	StoneWall Tile = '#'
	DirtFloor Tile = '.'
	CloseDoor Tile = '|'
	OpenDoor  Tile = '/'
	Void      Tile = ' '
	Player    Tile = '@'
)

//Level infor. for game
type Level struct {
	Map      [][]Tile
	Player   Entity
	Debug    map[Pos]bool
	Monsters []*Monster
}
type Entity struct {
	Symbol Tile
	Pos
	Speed int32
}
type Pos struct {
	X, Y int32
}

type priorityArray []priorityPos

//loadLvlFromFile will try to load level provided from "game/maps/" folder
func loadLvlFromFile(filPth string) (lvl *Level) {
	file, err := os.Open(filPth) //"game/maps/" + filNam)
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
	level.Monsters = make([]*Monster, 0)
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
			case '|', '/':
				t = CloseDoor
			case 'R':
				t = DirtFloor
				level.Monsters = append(level.Monsters, NewRat(int32(x*32), int32(y*32)))
			case 'S':
				t = DirtFloor
				level.Monsters = append(level.Monsters, NewSpider(int32(x*32), int32(y*32)))
			default:
				panic("undefined charecter in map")
			}
			level.Map[y][x] = t
		}
	}
	level.Player.X, level.Player.Y, level.Player.Symbol, level.Player.Speed = 84, 84, '@', 3
	return level
}
func getNeighbours(level *Level, pos Pos) []Pos {
	neghbours := make([]Pos, 0, 4)
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}
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
	y := int(pos.Y)
	x := int(pos.X)
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}
	return (level.Map[y][x] == DirtFloor || level.Map[y][x] == OpenDoor)
}
func (game *Game) bfs(start Pos) {
	start = Pos{int32(start.X / 32), int32(start.Y / 32)}
	level := game.Level
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
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (game *Game) astar(start Pos, goal Pos) []Pos {
	start = Pos{int32(start.X / 32), int32(start.Y / 32)}
	goal = Pos{int32(goal.X / 32), int32(goal.Y / 32)}
	lvl := game.Level
	frontier := make(pqueue, 0, 8)
	frontier = frontier.push(start, 1)
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0
	//lvl.Debug = make(map[Pos]bool)
	var current Pos
	for len(frontier) > 0 {
		frontier, current = frontier.pop()
		if current == goal {
			path := make([]Pos, 0)
			p := current
			for p != start {
				path = append(path, p)
				p = cameFrom[p]
			}
			path = append(path, p)
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			//lvl.Debug = make(map[Pos]bool)
			//for _, pos := range path {
			//	lvl.Debug[pos] = true
			//	//time.Sleep(100 * time.Millisecond)
			//}
			return path
		}
		for _, nxt := range getNeighbours(lvl, current) {
			newCost := costSoFar[current] + 1 //always 1 for now
			_, exists := costSoFar[nxt]
			if !exists || newCost < costSoFar[nxt] {
				costSoFar[nxt] = newCost
				xDist := int(math.Abs(float64(goal.X - nxt.X)))
				yDist := int(math.Abs(float64(goal.Y - nxt.Y)))
				priority := newCost + xDist + yDist
				frontier = frontier.push(nxt, priority)
				//lvl.Debug[nxt] = true
				cameFrom[nxt] = current
			}
		}
	}
	return nil
}
func (game *Game) Run() {
	fmt.Println("Starting...")
	for i := range game.LevelChans {
		game.LevelChans[i] <- game.Level
	}
	for input := range game.InputChan {
		if input.Typ == QuitGame {
			return
		}
		game.LevelManager(input)
		if len(game.LevelChans) == 0 {
			return
		}
		for _, lchan := range game.LevelChans {
			lchan <- game.Level
		}
	}
}
func (game *Game) LevelManager(input *Input) {
	p := game.Level.Player
	for _, j := range game.Level.Monsters {
		j.UpdateMons(p, game)
		if (p.Y >= j.Y-23 && p.Y <= j.Y+23) && (p.X >= j.X-23 && p.X <= j.X+23) {
			//fmt.Println("hit plyr")
			//hit monster
			if (p.Y > j.Y-20 && p.Y < j.Y+20) && (p.X > j.X-20 && p.X < j.X+20) {
				if p.Y < j.Y {
					game.Level.Player.Y -= p.Speed
					//fmt.Println("UP plyr:", p.Pos, ", Mons:", j.X, j.Y)
				} else if p.Y >= j.Y {
					game.Level.Player.Y += p.Speed
					//fmt.Println("Down plyr:", p.Pos, ", Mons:", j.X, j.Y)
				}
				if p.X < j.X {
					game.Level.Player.X -= p.Speed
					//fmt.Println("LEFT plyr:", p.Pos, ", Mons:", j.X, j.Y)
				} else if p.X >= j.X {
					game.Level.Player.X += p.Speed
					//fmt.Println("Right plyr:", p.Pos, ", Mons:", j.X, j.Y)
				}
			}
		}
	}
	if (*input).Typ == Up && (game.Level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == DirtFloor || game.Level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == OpenDoor) {
		game.Level.Player.Y -= p.Speed
	} else if (*input).Typ == Down && (game.Level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == DirtFloor || game.Level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == OpenDoor) {
		game.Level.Player.Y += p.Speed
	} else if (*input).Typ == Right && (game.Level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == DirtFloor || game.Level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == OpenDoor) {
		game.Level.Player.X += p.Speed
	} else if (*input).Typ == Left && (game.Level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == DirtFloor || game.Level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == OpenDoor) {
		game.Level.Player.X -= p.Speed
	} else if input.Typ == Open {
		if game.Level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] == CloseDoor {
			game.Level.Map[int((p.Y-2)/32)][int((p.X+18)/32)] = OpenDoor
		} else if game.Level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] == CloseDoor {
			game.Level.Map[int((p.Y+2)/32)+1][int((p.X+18)/32)] = OpenDoor
		} else if game.Level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] == CloseDoor {
			game.Level.Map[int((p.Y+18)/32)][int((p.X+2)/32)+1] = OpenDoor
		} else if game.Level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] == CloseDoor {
			game.Level.Map[int((p.Y+18)/32)][int((p.X-2)/32)] = OpenDoor
		}
	} else if input.Typ == Search {
		game.astar(p.Pos, Pos{278, 200})
	} else if input.Typ == CloseWindow {
		close(input.LevelChannel)
		chanIndex := 0
		for i, c := range game.LevelChans {
			if c == input.LevelChannel {
				chanIndex = i
				break
			}
		}
		game.LevelChans = append(game.LevelChans[:chanIndex], game.LevelChans[chanIndex+1:]...)
	}
}
func (monstr *Monster) UpdateMons(plyr Entity, game *Game) {
	posns := game.astar(monstr.Pos, plyr.Pos)
	if len(posns) > 3 && canSearch(game.Level, posns[3]) {
		if posns[3].Y*32 < monstr.Y {
			monstr.Y -= monstr.speed
		} else if posns[3].Y*32 > monstr.Y {
			monstr.Y += monstr.speed
		}
		if posns[3].X*32 > monstr.X {
			monstr.X += monstr.speed
		} else if posns[3].X*32 < monstr.X {
			monstr.X -= monstr.speed
		}
		//fmt.Println(posns[1], monstr.Pos)
	} else if len(posns) > 2 && canSearch(game.Level, posns[2]) {
		if posns[2].X*32 > monstr.X {
			monstr.X += monstr.speed
		} else if posns[2].X*32 < monstr.X {
			monstr.X -= monstr.speed
		}
		if posns[2].Y*32 < monstr.Y {
			monstr.Y -= monstr.speed
		} else if posns[2].Y*32 > monstr.Y {
			monstr.Y += monstr.speed
		}
		//fmt.Println(posns[1], monstr.Pos)
	} else if len(posns) > 1 {
		if posns[1].Y*32 < monstr.Y {
			monstr.Y -= monstr.speed
		} else if posns[1].Y*32 > monstr.Y {
			monstr.Y += monstr.speed
		}
		if posns[1].X*32 > monstr.X {
			monstr.X += monstr.speed
		} else if posns[1].X*32 < monstr.X {
			monstr.X -= monstr.speed
		}
		//fmt.Println(posns[1], monstr.Pos)
	}
}
