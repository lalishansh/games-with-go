package game

import (
	"fmt"
	"strconv"
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

var lookDirn = Right
var plyrRootedToGrnd bool
var idlePosCounter int

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
	EmptySpace // to invoke level manager//or to idle player
)

type Input struct {
	Typ          InputType
	LevelChannel chan *Level
}

type Tile struct {
	Rune    rune
	Visible bool
}

const (
	StoneWall rune = '#'
	DirtFloor      = '.'
	CloseDoor      = '|'
	OpenDoor       = '/'
	Void           = ' '
	Player         = '@'
)

//Level infor. for game
type Level struct {
	Map      [][]Tile
	Player   Charecter
	Monsters []*Monster
	Events   []string //like logs
	Debug    map[Pos]bool
}
type Entity struct {
	Symbol rune
	Pos
	Name string
}
type Charecter struct {
	Entity
	Hitpoints   int
	Strength    int
	Speed       int32
	actionTimer int // for attack sequence formulae[ actionTimer > int(3000/float32(j.Speed)) ]
	sightRange  int32
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
	level.Events = make([]string, 0)
	level.Monsters = make([]*Monster, 0)
	level.Map = make([][]Tile, len(levelLines))
	level.Debug = make(map[Pos]bool)
	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow)
	}
	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]
		var t rune
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
			case '@':
				t = DirtFloor
				level.Player.X, level.Player.Y, level.Player.Symbol, level.Player.Speed = int32(x*32), int32(y*32), '@', 3
				level.Player.Hitpoints, level.Player.Strength, level.Player.sightRange = 200, 4, 4
			default:
				panic("undefined charecter in map")
			}
			level.Map[y][x].Rune = t
			level.Map[y][x].Visible = false
		}
	}
	return level
}
func getNeighbours(level *Level, pos Pos) []Pos {
	neghbours := make([]Pos, 0, 4)
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}
	if canWalk(level, right) {
		neghbours = append(neghbours, right)
	}
	if canWalk(level, left) {
		neghbours = append(neghbours, left)
	}
	if canWalk(level, up) {
		neghbours = append(neghbours, up)
	}
	if canWalk(level, down) {
		neghbours = append(neghbours, down)
	}

	return neghbours
}
func canWalk(level *Level, pos Pos) bool {
	y := int(pos.Y)
	x := int(pos.X)
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}
	return (level.Map[y][x].Rune == DirtFloor || level.Map[y][x].Rune == OpenDoor)
}
func canBSeen(level *Level, pos Pos) bool {
	y := int(pos.Y)
	x := int(pos.X)
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}
	return !(level.Map[y][x].Rune == StoneWall || level.Map[y][x].Rune == CloseDoor || level.Map[y][x].Rune == Void)
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
func bresenhum(start, end Pos) (result []Pos) {
	result = make([]Pos, 0)
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X))
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}
	if start.X > end.X {
		start.X, end.X = end.X, start.X
		start.Y, end.Y = end.Y, start.Y
	}
	deltaX := end.X - start.X
	deltaY := int32(math.Abs(float64(end.Y - start.Y)))
	var err int32
	y := start.Y
	var yStep int32 = 1
	if start.Y >= end.Y {
		yStep = -1
	}
	for x := start.X; x < end.X; x++ {
		if steep {
			result = append(result, Pos{y, x})
		} else {
			result = append(result, Pos{x, y})
		}
		err += deltaY
		if 2*err >= deltaX {
			y += yStep
			err -= deltaX
		}
	}
	return result
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
		for _, j := range game.Level.Monsters {
			j.actionTimer += 15
		}
		game.Level.Player.actionTimer += 15
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
	p := game.Level.Player.Pos
	//line := bresenhum(p.Div32(), Pos{p.Div32().X + 5, p.Div32().Y - 5})
	//for _, t := range line {
	//	game.Level.Debug[t] = true
	//}
	plyr := &game.Level.Player
	UpdatePlayer(input.Typ, game.Level)
	if input.Typ == CloseWindow {
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
	var reCalMons = false
	for _, j := range game.Level.Monsters {
		j.UpdateMons(plyr, game)
		if (p.Y >= j.Y-22 && p.Y <= j.Y+22) && (p.X >= j.X-22 && p.X <= j.X+22) {
			plyrCollisnHandler(&j.Pos, plyr, j, game.Level)
			//hit plyr
			j.Debug = true
			if j.actionTimer > int(3000/float32(j.Speed)) {
				game.Level.Player.Hitpoints -= j.Strength
				game.Level.Events = append(game.Level.Events, j.Name+" hit player health remaining:"+strconv.Itoa(game.Level.Player.Hitpoints))
				if len(game.Level.Events) > 12 {
					game.Level.Events = game.Level.Events[1:]
				}
				j.actionTimer = 0
			}
			if game.Level.Player.Hitpoints <= 0 {
				//kill plyr
				//
				//game over
				panic("you died")
			}
			//hit monster
			if plyr.actionTimer > int(3000/float32(plyr.Speed)) {
				if lookDirn == Up && p.Y > j.Y {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.Level.Events = append(game.Level.Events, "player hit "+j.Name+" up its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.Level.Events = append(game.Level.Events, "player Killed "+j.Name)
					}
					if len(game.Level.Events) > 12 {
						game.Level.Events = game.Level.Events[1:]
					}
				} else if lookDirn == Down && p.Y < j.Y {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.Level.Events = append(game.Level.Events, "player hit "+j.Name+" down its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.Level.Events = append(game.Level.Events, "player Killed "+j.Name)
					}
					if len(game.Level.Events) > 12 {
						game.Level.Events = game.Level.Events[1:]
					}
				} else if lookDirn == Right && p.X < j.X {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.Level.Events = append(game.Level.Events, "player hit "+j.Name+" right its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.Level.Events = append(game.Level.Events, "player Killed "+j.Name)
					}
					if len(game.Level.Events) > 12 {
						game.Level.Events = game.Level.Events[1:]
					}
				} else if lookDirn == Left && p.X > j.X {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.Level.Events = append(game.Level.Events, "player hit "+j.Name+" left its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.Level.Events = append(game.Level.Events, "player Killed "+j.Name)
					}
					if len(game.Level.Events) > 12 {
						game.Level.Events = game.Level.Events[1:]
					}
				}
				game.Level.Player.actionTimer = 0
			}
			//killing after loop
			if j.Hitpoints <= 0 {
				reCalMons = true
			}
			//plyrCollisnHandler(&j.Pos, plyr, j, game.Level)
		} else {
			j.Debug = false
		}
	}
	if reCalMons {
		aliveMonsters := make([]*Monster, 0)
		for _, j := range game.Level.Monsters {
			if j.Hitpoints > 0 {
				aliveMonsters = append(aliveMonsters, j)
			}
		}
		game.Level.Monsters = aliveMonsters
		plyrRootedToGrnd = false
		idlePosCounter = 0
	}
}
func (monstr *Monster) UpdateMons(plyr *Charecter, game *Game) {
	if monstr.activ {
		posns := game.astar(monstr.Pos, Pos{plyr.Pos.X + 16, plyr.Pos.Y + 16})
		if len(posns) > 3 && canWalk(game.Level, posns[3]) {
			if posns[3].Y*32 < monstr.Y {
				monstr.Y -= monstr.Speed
			}
			if posns[3].Y*32 > monstr.Y {
				monstr.Y += monstr.Speed
			}
			if posns[3].X*32 > monstr.X {
				monstr.X += monstr.Speed
			}
			if posns[3].X*32 < monstr.X {
				monstr.X -= monstr.Speed
			}
			//fmt.Println(posns[1], monstr.Pos)
		} else if len(posns) > 2 && canWalk(game.Level, posns[2]) {
			if posns[2].X*32 > monstr.X {
				monstr.X += monstr.Speed
			}
			if posns[2].X*32 < monstr.X {
				monstr.X -= monstr.Speed
			}
			if posns[2].Y*32 < monstr.Y {
				monstr.Y -= monstr.Speed
			}
			if posns[2].Y*32 > monstr.Y {
				monstr.Y += monstr.Speed
			}
			//fmt.Println(posns[1], monstr.Pos)
		} else if len(posns) > 1 {
			if posns[1].Y*32 < monstr.Y {
				monstr.Y -= monstr.Speed
			}
			if posns[1].Y*32 > monstr.Y {
				monstr.Y += monstr.Speed
			}
			if posns[1].X*32 > monstr.X {
				monstr.X += monstr.Speed
			}
			if posns[1].X*32 < monstr.X {
				monstr.X -= monstr.Speed
			}
			//fmt.Println(posns[1], monstr.Pos)
		}
		if game.Level.Map[(monstr.Y+16)/32][monstr.X/32].Rune == StoneWall {
			monstr.X += monstr.Speed
		}
		if game.Level.Map[(monstr.Y+16)/32][(monstr.X+30)/32].Rune == StoneWall {
			monstr.X -= monstr.Speed
		}
		if game.Level.Map[monstr.Y/32][(monstr.X+16)/32].Rune == StoneWall {
			monstr.Y += monstr.Speed
		}
		if game.Level.Map[(monstr.Y+30)/32][(monstr.X+16)/32].Rune == StoneWall {
			monstr.Y -= monstr.Speed
		}
		for _, x := range game.Level.Monsters {
			if monstr == x {
				continue
			} else {
				if (x.X-monstr.X < 18 && x.X-monstr.X > -18) && (x.Y-monstr.Y < 18 && x.Y-monstr.Y > -18) {
					if x.Y > monstr.Y {
						monstr.Y -= monstr.Speed
					} else if x.Y < monstr.Y {
						monstr.Y += monstr.Speed
					}
					if x.X > monstr.X {
						monstr.X -= monstr.Speed
					} else if x.X < monstr.X {
						monstr.X += monstr.Speed
					}
				}
			}
		}
		if game.Level.Map[(monstr.Y+16)/32][monstr.X/32].Rune == StoneWall {
			monstr.X += monstr.Speed
		}
	}
}
func (p *Charecter) applicatnOfbresenhum(pos Pos, lvl *Level) {
	//for _, row := range lvl.Map {
	//	for _, tile := range row {
	//		tile.Visible = false
	//	}
	//}
	line := bresenhum(pos, pos.add(p.sightRange, -p.sightRange))
	for _, posn := range line {
		if canBSeen(lvl, posn) {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
		} else {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
			break
		}
	}
	line = bresenhum(pos, pos.add(p.sightRange, p.sightRange))
	for _, posn := range line {
		if canBSeen(lvl, posn) {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
		} else {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
			break
		}
	}
	line = bresenhum(pos, pos.add(-p.sightRange, p.sightRange))
	for _, posn := range line {
		if canBSeen(lvl, posn) {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
		} else {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
			break
		}
	}
	line = bresenhum(pos, pos.add(-p.sightRange, -p.sightRange))
	for _, posn := range line {
		if canBSeen(lvl, posn) {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
		} else {
			if posn.Y < 0 {
				posn.Y = 0
			}
			if posn.X < 0 {
				posn.X = 0
			}
			lvl.Map[posn.Y][posn.X].Visible = true
			break
		}
	}
}

func UpdatePlayer(input InputType, lvl *Level) {
	p := &lvl.Player
	if input == Up && (lvl.Map[int((p.Y-2)/32)][int((p.X+16)/32)].Rune == DirtFloor || lvl.Map[int((p.Y-2)/32)][int((p.X+16)/32)].Rune == OpenDoor) {
		lvl.Player.Y -= p.Speed
		//
		p.applicatnOfbresenhum(p.Pos.Div32(), lvl)
		lookDirn = Up
	} else if input == Down && (lvl.Map[int((p.Y+2)/32)+1][int((p.X+16)/32)].Rune == DirtFloor || lvl.Map[int((p.Y+2)/32)+1][int((p.X+16)/32)].Rune == OpenDoor) {
		lvl.Player.Y += p.Speed
		//
		p.applicatnOfbresenhum(p.Pos.Div32(), lvl)
		lookDirn = Down
	} else if input == Right && (lvl.Map[int((p.Y+16)/32)][int((p.X+2)/32)+1].Rune == DirtFloor || lvl.Map[int((p.Y+16)/32)][int((p.X+2)/32)+1].Rune == OpenDoor) {
		lvl.Player.X += p.Speed
		//
		p.applicatnOfbresenhum(p.Pos.Div32(), lvl)
		lookDirn = Right
	} else if input == Left && (lvl.Map[int((p.Y+16)/32)][int((p.X-2)/32)].Rune == DirtFloor || lvl.Map[int((p.Y+16)/32)][int((p.X-2)/32)].Rune == OpenDoor) {
		lvl.Player.X -= p.Speed
		//
		p.applicatnOfbresenhum(p.Pos.Div32(), lvl)
		lookDirn = Left
	} else if input == Open {
		if lvl.Map[int((p.Y-2)/32)][int((p.X+16)/32)].Rune == CloseDoor {
			lvl.Map[int((p.Y-2)/32)][int((p.X+16)/32)].Rune = OpenDoor
		} else if lvl.Map[int((p.Y+2)/32)+1][int((p.X+16)/32)].Rune == CloseDoor {
			lvl.Map[int((p.Y+2)/32)+1][int((p.X+16)/32)].Rune = OpenDoor
		} else if lvl.Map[int((p.Y+16)/32)][int((p.X+2)/32)+1].Rune == CloseDoor {
			lvl.Map[int((p.Y+16)/32)][int((p.X+2)/32)+1].Rune = OpenDoor
		} else if lvl.Map[int((p.Y+16)/32)][int((p.X-2)/32)].Rune == CloseDoor {
			lvl.Map[int((p.Y+16)/32)][int((p.X-2)/32)].Rune = OpenDoor
		}
	} else if input == EmptySpace {
		idlePosCounter++
		if idlePosCounter > 60 {
			plyrRootedToGrnd = true
		}
	}
	if input != EmptySpace {
		idlePosCounter = 0
		plyrRootedToGrnd = false
	}
}
func (p Pos) LeftN() Pos {
	p.X -= 1
	p.Y += 16
	p.X /= 32
	p.Y /= 32
	return p
}
func (p Pos) RightN() Pos {
	p.X += 32
	p.Y += 16
	p.X /= 32
	p.Y /= 32
	return p
}
func (p Pos) UpN() Pos {
	p.X += 16
	p.Y -= 1
	p.X /= 32
	p.Y /= 32
	return p
}
func (p Pos) DownN() Pos {
	p.X += 16
	p.Y += 32
	p.X /= 32
	p.Y /= 32
	return p
}
func (pos Pos) add(x, y int32) Pos {
	pos.X += x
	pos.Y += y
	return pos
}

//Div32 returns corresponding posn in map
func (pos Pos) Div32() Pos {
	pos.X = (pos.X + 16) / 32
	pos.Y = (pos.Y + 16) / 32
	return pos
}
func plyrCollisnHandler(j *Pos, plyr *Charecter, mons *Monster, lvl *Level) {
	p := plyr.Pos
	if (j.X-p.X < 18 && j.X-p.X > -18) && (j.Y-p.Y < 18 && j.Y-p.Y > -18) {
		mons.Debug2 = true
		if j.Y > p.Y {
			//fmt.Println("!canWalk(lvl,", p.UpN(), ")=", !canWalk(lvl, p.UpN()), lvl.Map[p.UpN().Y][p.UpN().X])
			if !canWalk(lvl, p.UpN()) || plyrRootedToGrnd {
				//fmt.Printf("j.Y += (18 - j.Y + p.Y) >=> %v += (18 - %v + %v),(%v)\n\n", j.Y, j.Y, p.Y, (18 - j.Y + p.Y))
				j.Y += (18 - j.Y + p.Y)
				//fmt.Println("j.Pos :", j)
			} else {
				//fmt.Printf("up plyr.Y -= (18 - j.Y + p.Y) >=> %v -= (18 - %v + %v),(%v)\n\n", plyr.Y, j.Y, p.Y, (18 - j.Y + p.Y))
				plyr.Y -= (18 - j.Y + p.Y)
			}
		} else if j.Y < plyr.Y {
			//fmt.Println("down !canWalk(lvl,", p.DownN(), ")=", !canWalk(lvl, p.DownN()), lvl.Map[p.DownN().Y][p.DownN().X])
			if !canWalk(lvl, p.DownN()) || plyrRootedToGrnd {
				//fmt.Printf("j.Y -= (18 - p.Y + j.Y) >=> %v -= (18 - %v + %v),(%v)\n\n", j.Y, p.Y, j.Y, (18 - p.Y + j.Y))
				j.Y -= (18 - p.Y + j.Y)
			} else {
				//fmt.Printf("plyr.Y += (18 - p.Y + j.Y) >=> %v += (18 - %v + %v),(%v)\n\n", plyr.Y, p.Y, j.Y, (18 - p.Y + j.Y))
				plyr.Y += (18 - p.Y + j.Y)
			}
		}
		if j.X > p.X {
			//fmt.Println("left !canWalk(lvl,", p.LeftN(), ")=", !canWalk(lvl, p.LeftN()), lvl.Map[p.LeftN().Y][p.LeftN().X])
			if !canWalk(lvl, p.LeftN()) || plyrRootedToGrnd {
				//fmt.Printf("j.X += (18 - j.X + p.X) >=> %v += (18 - %v + %v),(%v)\n\n", j.X, j.X, p.X, (18 - j.X + p.X))
				j.X += (18 - j.X + p.X)
			} else {
				//fmt.Printf("plyr.X -= (18 - j.X + p.X) >=> %v -= (18 - %v + %v),(%v)\n\n", plyr.X, j.X, p.X, (18 - j.X + p.X))
				plyr.X -= (18 - j.X + p.X)
			}
		} else if j.X < plyr.X {
			//fmt.Println("right !canWalk(lvl,", p.RightN(), ")=", !canWalk(lvl, p.RightN()), lvl.Map[p.RightN().Y][p.RightN().X])
			if !canWalk(lvl, p.RightN()) || plyrRootedToGrnd {
				//fmt.Printf("j.X -= (18 - p.X + j.X) >=> %v -= (18 - %v + %v),(%v)\n\n", j.X, p.X, j.X, (18 - p.X + j.X))
				j.X -= (18 - p.X + j.X)
			} else {
				//fmt.Printf("plyr.X += (18 - p.X + j.X) >=> %v += (18 - %v + %v),(%v)\n\n", plyr.X, p.X, j.X, (18 - p.X + j.X))
				plyr.X += (18 - p.X + j.X)
			}
		}
	} else {
		mons.Debug2 = false
	}
}

/*
if (p.Y > j.Y-18 && p.Y < j.Y+18) && (p.X > j.X-18 && p.X < j.X+18) {
				if p.Y < j.Y {
					if canWalk(game.Level, game.Level.Player.Pos.UpN()) {
						j.Y += p.Speed
					} else {
						game.Level.Player.Y -= p.Speed
					}
					//fmt.Println("UP plyr:", p.Pos, ", Mons:", j.X, j.Y)
				} else {
					if canWalk(game.Level, game.Level.Player.Pos.DownN()) {
						j.Y -= p.Speed
					} else {
						game.Level.Player.Y += p.Speed
					} //fmt.Println("Down plyr:", p.Pos, ", Mons:", j.X, j.Y)
				}
				if p.X < j.X {
					if canWalk(game.Level, game.Level.Player.Pos.RightN()) {
						j.X += p.Speed
					} else {
						game.Level.Player.X -= p.Speed
					} //fmt.Println("LEFT plyr:", p.Pos, ", Mons:", j.X, j.Y)
				} else {
					if canWalk(game.Level, game.Level.Player.Pos.LeftN()) {
						j.X -= p.Speed
					} else {
						game.Level.Player.X += p.Speed
					} //fmt.Println("Right plyr:", p.Pos, ", Mons:", j.X, j.Y)
				}
			}
*/
