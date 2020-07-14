package game

import (
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bufio"
	"math"
	"os"

	"github.com/t-RED-69/games-with-go/RPG/UI2d/sound"
)

var str string

type Game struct {
	LevelChans []chan *Level
	InputChan  chan *Input
	Levels     map[string]*Level
	CurrLevel  *Level
}

var lookDirn = Right
var plyrRootedToGrnd bool
var idlePosCounter int

func NewGame(numWindows int) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}
	inputChan := make(chan *Input)

	sound.NewSI()

	lvls := loadLvls()
	//Temporary
	game := &Game{levelChans, inputChan, lvls, nil}
	game.loadWorldFile()
	game.CurrLevel.lineOfSight(game.CurrLevel.Player.Pos.Div32())
	return game
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
	Seen    bool
}

const (
	StoneWall rune = '#'
	DirtFloor      = '.'
	CloseDoor      = '|'
	OpenDoor       = '/'
	Void           = ' '
	Player         = '@'
	StairUp        = 'u'
	StairDown      = 'd'
)

//Level infor. for game
type Level struct {
	Map      [][]Tile
	Player   Charecter
	Monsters []*Monster
	Portals  map[Pos]levelPos
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

type levelPos struct {
	level *Level
	stair Pos
}

func (game *Game) loadWorldFile() {
	file, err := os.Open("game/maps/world.txt")
	if err != nil {
		panic(err)
	}
	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true
	rows, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	for i, row := range rows {
		if i == 0 {
			game.CurrLevel = game.Levels[row[0]]
			continue
		}
		lvlPortalfrom := row[0]
		x, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			panic(err)
		}
		posToPortalfrom := Pos{int32(x), int32(y)}
		lvlFrom := game.Levels[lvlPortalfrom]
		if lvlFrom == nil {
			panic("could'nt find level specified in world file : " + lvlPortalfrom)
		}
		lvlPortalTo := row[3]
		x, err = strconv.ParseInt(row[4], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err = strconv.ParseInt(row[5], 10, 64)
		if err != nil {
			panic(err)
		}
		posToPortalTo := Pos{int32(x), int32(y)}
		lvlTo := game.Levels[lvlPortalTo]
		if lvlTo == nil {
			panic("could'nt find level specified in world file : " + lvlPortalTo)
		}
		lvlFrom.Portals[posToPortalfrom] = levelPos{lvlTo, posToPortalTo}
	}
}

//loadLvlFromFile will try to load level provided from "game/maps/" folder
func loadLvls() map[string]*Level {

	filepaths, err := filepath.Glob("game/maps/*.map")
	if err != nil {
		panic(err)
	}
	lvls := make(map[string]*Level)
	for _, filpth := range filepaths {
		file, err := os.Open(filpth)
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
		level.Portals = make(map[Pos]levelPos)
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
					level.Player.Hitpoints, level.Player.Strength, level.Player.sightRange = 200, 4, 5
				case 'u':
					t = StairUp
				case 'd':
					t = StairDown
				default:
					panic("undefined charecter in map")
				}
				level.Map[y][x].Rune = t
				level.Map[y][x].Visible = false
			}
		}
		extIndex := strings.LastIndex(filpth, ".map")
		lastSlash := strings.LastIndex(filpth, "\\")
		levelName := filpth[(lastSlash + 1):extIndex]
		lvls[levelName] = level
	}
	//level.lineOfSight(level.Player.Div32())
	return lvls
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

//func canBSeen(level *Level, pos Pos) bool {
//	y := int(pos.Y)
//	x := int(pos.X)
//	fmt.Println(x, y)
//	return !(level.Map[x][y].Rune == StoneWall || level.Map[x][y].Rune == CloseDoor || level.Map[x][y].Rune == Void)
//}
func (game *Game) bfs(start Pos) {
	start = Pos{int32(start.X / 32), int32(start.Y / 32)}
	level := game.CurrLevel
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

//////////////////////////  understand
// Is there line of sight/a window?
func canSeeThrough(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		// Check tile for solid object
		t := level.Map[pos.Y][pos.X]
		switch t.Rune {
		case StoneWall, CloseDoor, Void:
			return false
		default:
			return true
		}
	}
	return false
}
func inRange(level *Level, pos Pos) bool {
	return pos.X < int32(len(level.Map[0])) && pos.Y < int32(len(level.Map)) && pos.X >= 0 && pos.Y >= 0
}

// Draw a circle around the player and draw a line to each endpoint
func (level *Level) bresenhum(start Pos, end Pos) {
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X)) // Is the line steep or not?
	// Swap the x and y for start and end
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}

	deltaY := int32(math.Abs(float64(end.Y - start.Y)))

	var err int32
	y := start.Y
	var ystep int32 = 1 // How far we are stepping when err is above threshold
	if start.Y >= end.Y {
		ystep = -1 // Reverse it when we step
	}
	// Are we on the left or right side of graph
	if start.X > end.X {
		deltaX := start.X - end.X // We know start.X will be larger than end.X
		// Count down so lines extend FROM the player, not TO
		for x := start.X; x > end.X; x-- {
			var pos Pos
			if steep {
				pos = Pos{y, x} // If we are steep, x and y will be swapped
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			level.Map[pos.Y][pos.X].Seen = true
			if !canSeeThrough(level, pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep // Go up or down depending on the direction of our line
				err -= deltaX
			}
		}
	} else {
		deltaX := end.X - start.X // We know start.X will be larger than end.X
		for x := start.X; x < end.X; x++ {
			var pos Pos
			if steep {
				pos = Pos{y, x} // If we are steep, x and y will be swapped
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			level.Map[pos.Y][pos.X].Seen = true
			if !canSeeThrough(level, pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep // Go up or down depending on the direction of our line
				err -= deltaX
			}
		}
	}
}

//////////////////////////
/*
func (lvl *Level) bresenhum(start, end Pos) {
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X))
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}
	deltaY := int32(math.Abs(float64(end.Y - start.Y)))
	var err int32 = 0
	y := start.Y
	var yStep int32 = 1
	if start.Y >= end.Y {
		yStep = -1
	}
	if start.X > end.X {
		deltaX := start.X - end.X
		for x := start.X; x > end.X; x-- {
			if canBSeen(lvl, Pos{x, y}) {
				if steep {
					lvl.Map[x][y].Visible = true
				} else {
					lvl.Map[y][x].Visible = true
				}
			} else {
				if steep {
					lvl.Map[x][y].Visible = true
				} else {
					lvl.Map[y][x].Visible = true
				}
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += yStep
				err -= deltaX
			}
		}
	} else {
		deltaX := end.X - start.X
		for x := start.X; x < end.X; x++ {
			if canBSeen(lvl, Pos{x, y}) {
				if steep {
					lvl.Map[x][y].Visible = true
				} else {
					lvl.Map[y][x].Visible = true
				}
			} else {
				if steep {
					lvl.Map[x][y].Visible = true
				} else {
					lvl.Map[y][x].Visible = true
				}
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += yStep
				err -= deltaX
			}
		}
	}
	/*if start.X > end.X {
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
	if start.X < 0 {
		start.X = 0
	}
	for x := start.X; x < end.X; x++ {
		if canBSeen(lvl, Pos{x, y}) {
			if y < 0 {
				y = 0
			}
			if steep {
				lvl.Map[x][y].Visible = true
			} else {
				lvl.Map[y][x].Visible = true
			}
		} else {
			if y < 0 {
				y = 0
			}
			fmt.Println("x,y:", x, y)
			if steep {
				lvl.Map[x][y].Visible = true
			} else {
				lvl.Map[y][x].Visible = true
			}
			return
		}
		err += deltaY
		if 2*err >= deltaX {
			y += yStep
			err -= deltaX
		}
	}*/
/*
}
*/
func (lvl *Level) lineOfSight(pos Pos) {
	for i, row := range lvl.Map {
		for j := range row {
			lvl.Map[i][j].Visible = false
		}
	}
	dist := lvl.Player.sightRange
	//fmt.Println(pos)
	y := pos.Y - dist
	limitY := pos.Y + dist
	for ; y <= limitY; y++ {
		x := pos.X - dist
		limitX := pos.X + dist
		for ; x <= limitX; x++ {
			deltaX := pos.X - x
			deltaY := pos.Y - y
			d2 := deltaX*deltaX + deltaY*deltaY
			if d2 <= dist*dist {
				//fmt.Printf("0")
				lvl.bresenhum(pos, Pos{x, y})
			} else {
				//fmt.Printf(".")
			}
		}
		//fmt.Println()
	}
}
func (game *Game) astar(start Pos, goal Pos) []Pos {
	start = Pos{int32(start.X / 32), int32(start.Y / 32)}
	goal = Pos{int32(goal.X / 32), int32(goal.Y / 32)}
	lvl := game.CurrLevel
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
		game.LevelChans[i] <- game.CurrLevel
	}
	for _, lchan := range game.LevelChans {
		lchan <- game.CurrLevel
	}
	for input := range game.InputChan {
		for _, j := range game.CurrLevel.Monsters {
			j.actionTimer += 15
		}
		game.CurrLevel.Player.actionTimer += 15
		if input.Typ == QuitGame {
			return
		}
		game.LevelManager(input)
		if len(game.LevelChans) == 0 {
			return
		}
		for _, lchan := range game.LevelChans {
			lchan <- game.CurrLevel
		}
	}
}
func (game *Game) LevelManager(input *Input) {
	p := game.CurrLevel.Player.Pos
	UpdatePlayer(input.Typ, game.CurrLevel)
	if game.CurrLevel.Map[int((p.Y+16)/32)][int((p.X+16)/32)].Rune == StairUp {
		fmt.Println("newLevel,pos(stairsUp)", p.Div32())
		game.CurrLevel.Player.Pos = game.CurrLevel.Portals[p.Div32()].stair.Mult32() //.add(-15, 0)
		pl := game.CurrLevel.Player
		game.CurrLevel = game.CurrLevel.Portals[p.Div32()].level
		game.CurrLevel.Player = pl
		game.CurrLevel.lineOfSight(game.CurrLevel.Player.Pos.Div32())
		return
	} else if game.CurrLevel.Map[int((p.Y+16)/32)][int((p.X+16)/32)].Rune == StairDown {
		fmt.Println("newLevel,pos(stairsDn)", p.Div32())
		game.CurrLevel.Player.Pos = game.CurrLevel.Portals[p.Div32()].stair.Mult32() //.add(0, -15)
		pl := game.CurrLevel.Player
		game.CurrLevel = game.CurrLevel.Portals[p.Div32()].level
		game.CurrLevel.Player = pl
		game.CurrLevel.lineOfSight(game.CurrLevel.Player.Pos.Div32())
	}
	plyr := &game.CurrLevel.Player
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
	for _, j := range game.CurrLevel.Monsters {
		j.UpdateMons(plyr, game)
		if (p.Y >= j.Y-22 && p.Y <= j.Y+22) && (p.X >= j.X-22 && p.X <= j.X+22) {
			plyrCollisnHandler(&j.Pos, plyr, j, game.CurrLevel)
			//hit plyr
			j.Debug = true
			if j.actionTimer > int(3000/float32(j.Speed)) {
				game.CurrLevel.Player.Hitpoints -= j.Strength
				game.CurrLevel.Events = append(game.CurrLevel.Events, j.Name+" hit player health remaining:"+strconv.Itoa(game.CurrLevel.Player.Hitpoints))
				if len(game.CurrLevel.Events) > 12 {
					game.CurrLevel.Events = game.CurrLevel.Events[1:]
				}
				j.actionTimer = 0
				sound.Play(sound.EnmyHitINT, j.Name, 'a')
			}
			if game.CurrLevel.Player.Hitpoints <= 0 {
				//kill plyr
				//
				//game over
				panic("you died")
			}
			//hit monster
			if plyr.actionTimer > int(3500/float32(plyr.Speed)) {
				sound.Play(sound.PlyrHitINT, "", 0)
				if lookDirn == Up && p.Y > j.Y {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player hit "+j.Name+" up its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player Killed "+j.Name)
					}
					if len(game.CurrLevel.Events) > 12 {
						game.CurrLevel.Events = game.CurrLevel.Events[1:]
					}
				} else if lookDirn == Down && p.Y < j.Y {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player hit "+j.Name+" down its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player Killed "+j.Name)
					}
					if len(game.CurrLevel.Events) > 12 {
						game.CurrLevel.Events = game.CurrLevel.Events[1:]
					}
				} else if lookDirn == Right && p.X < j.X {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player hit "+j.Name+" right its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player Killed "+j.Name)
					}
					if len(game.CurrLevel.Events) > 12 {
						game.CurrLevel.Events = game.CurrLevel.Events[1:]
					}
				} else if lookDirn == Left && p.X > j.X {
					j.Hitpoints -= plyr.Strength
					if j.Hitpoints > 0 {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player hit "+j.Name+" left its hlth remaining:"+strconv.Itoa(j.Hitpoints))
					} else {
						game.CurrLevel.Events = append(game.CurrLevel.Events, "player Killed "+j.Name)
					}
					if len(game.CurrLevel.Events) > 12 {
						game.CurrLevel.Events = game.CurrLevel.Events[1:]
					}
				}
				sound.Play(sound.EnmyHitINT, j.Name, 'h')
				game.CurrLevel.Player.actionTimer = 0
			}
			//killing after loop
			if j.Hitpoints <= 0 {
				sound.Play(sound.EnmyHitINT, j.Name, 'd')
				reCalMons = true
			}
			//plyrCollisnHandler(&j.Pos, plyr, j, game.CurrLevel)
		} else {
			j.Debug = false
		}
	}
	if reCalMons {
		aliveMonsters := make([]*Monster, 0)
		for _, j := range game.CurrLevel.Monsters {
			if j.Hitpoints > 0 {
				aliveMonsters = append(aliveMonsters, j)
			}
		}
		game.CurrLevel.Monsters = aliveMonsters
		plyrRootedToGrnd = false
		idlePosCounter = 0
	}
}
func (monstr *Monster) UpdateMons(plyr *Charecter, game *Game) {
	if monstr.activ {
		posns := game.astar(monstr.Pos, Pos{plyr.Pos.X + 16, plyr.Pos.Y + 16})
		if len(posns) > 3 && canWalk(game.CurrLevel, posns[3]) {
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
		} else if len(posns) > 2 && canWalk(game.CurrLevel, posns[2]) {
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
		if game.CurrLevel.Map[(monstr.Y+16)/32][monstr.X/32].Rune == StoneWall {
			monstr.X += monstr.Speed
		}
		if game.CurrLevel.Map[(monstr.Y+16)/32][(monstr.X+30)/32].Rune == StoneWall {
			monstr.X -= monstr.Speed
		}
		if game.CurrLevel.Map[monstr.Y/32][(monstr.X+16)/32].Rune == StoneWall {
			monstr.Y += monstr.Speed
		}
		if game.CurrLevel.Map[(monstr.Y+30)/32][(monstr.X+16)/32].Rune == StoneWall {
			monstr.Y -= monstr.Speed
		}
		for _, x := range game.CurrLevel.Monsters {
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
		if game.CurrLevel.Map[(monstr.Y+16)/32][monstr.X/32].Rune == StoneWall {
			monstr.X += monstr.Speed
		}
	}
}

func UpdatePlayer(input InputType, lvl *Level) {
	p := &lvl.Player
	if input == Up && (lvl.Map[int((p.Y-1)/32)][int((p.X+16)/32)].Rune == DirtFloor || lvl.Map[int((p.Y-1)/32)][int((p.X+16)/32)].Rune == OpenDoor || lvl.Map[int((p.Y-1)/32)][int((p.X+16)/32)].Rune == StairUp || lvl.Map[int((p.Y-1)/32)][int((p.X+16)/32)].Rune == StairDown) {
		lvl.Player.Y -= p.Speed
		lvl.lineOfSight(p.Pos.Div32())
		sound.Play(sound.FootstpsINT, "", 0)
		lookDirn = Up
	} else if input == Down && (lvl.Map[int((p.Y+1)/32)+1][int((p.X+16)/32)].Rune == DirtFloor || lvl.Map[int((p.Y+1)/32)+1][int((p.X+16)/32)].Rune == OpenDoor || lvl.Map[int((p.Y+1)/32)+1][int((p.X+16)/32)].Rune == StairUp || lvl.Map[int((p.Y+1)/32)+1][int((p.X+16)/32)].Rune == StairDown) {
		lvl.Player.Y += p.Speed
		lvl.lineOfSight(p.Pos.Div32())
		sound.Play(sound.FootstpsINT, "", 0)
		lookDirn = Down
	} else if input == Right && (lvl.Map[int((p.Y+16)/32)][int((p.X+1)/32)+1].Rune == DirtFloor || lvl.Map[int((p.Y+16)/32)][int((p.X+1)/32)+1].Rune == OpenDoor || lvl.Map[int((p.Y+16)/32)][int((p.X+1)/32)+1].Rune == StairUp || lvl.Map[int((p.Y+16)/32)][int((p.X+1)/32)+1].Rune == StairDown) {
		lvl.Player.X += p.Speed
		lvl.lineOfSight(p.Pos.Div32())
		sound.Play(sound.FootstpsINT, "", 0)
		lookDirn = Right
	} else if input == Left && (lvl.Map[int((p.Y+16)/32)][int((p.X-1)/32)].Rune == DirtFloor || lvl.Map[int((p.Y+16)/32)][int((p.X-1)/32)].Rune == OpenDoor || lvl.Map[int((p.Y+16)/32)][int((p.X-1)/32)].Rune == StairUp || lvl.Map[int((p.Y+16)/32)][int((p.X-1)/32)].Rune == StairDown) {
		lvl.Player.X -= p.Speed
		lvl.lineOfSight(p.Pos.Div32())
		sound.Play(sound.FootstpsINT, "", 0)
		lookDirn = Left
	} else {
		sound.HaltSounds(sound.FootstpsINT)
	}
	if input == Open {
		if lvl.Map[int((p.Y-1)/32)][int((p.X+16)/32)].Rune == CloseDoor {
			lvl.Map[int((p.Y-1)/32)][int((p.X+16)/32)].Rune = OpenDoor
			sound.Play(sound.DoorOpnINT, "", 0)
		} else if lvl.Map[int((p.Y+1)/32)+1][int((p.X+16)/32)].Rune == CloseDoor {
			lvl.Map[int((p.Y+1)/32)+1][int((p.X+16)/32)].Rune = OpenDoor
			sound.Play(sound.DoorOpnINT, "", 0)
		} else if lvl.Map[int((p.Y+16)/32)][int((p.X+1)/32)+1].Rune == CloseDoor {
			lvl.Map[int((p.Y+16)/32)][int((p.X+1)/32)+1].Rune = OpenDoor
			sound.Play(sound.DoorOpnINT, "", 0)
		} else if lvl.Map[int((p.Y+16)/32)][int((p.X-1)/32)].Rune == CloseDoor {
			lvl.Map[int((p.Y+16)/32)][int((p.X-1)/32)].Rune = OpenDoor
			sound.Play(sound.DoorOpnINT, "", 0)
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
func (pos Pos) Mult32() Pos {
	pos.X = pos.X * 32
	pos.Y = pos.Y * 32
	return pos
}
func plyrCollisnHandler(j *Pos, plyr *Charecter, mons *Monster, lvl *Level) {
	p := plyr.Pos
	if (j.X-p.X < 18 && j.X-p.X > -18) && (j.Y-p.Y < 18 && j.Y-p.Y > -18) {
		mons.Debug2 = true
		if j.Y > p.Y {
			if !canWalk(lvl, p.UpN()) || plyrRootedToGrnd {
				j.Y += (18 - j.Y + p.Y)
			} else {
				plyr.Y -= (18 - j.Y + p.Y)
				lvl.lineOfSight(p.Div32())
			}
		} else if j.Y < plyr.Y {
			if !canWalk(lvl, p.DownN()) || plyrRootedToGrnd {
				j.Y -= (18 - p.Y + j.Y)
			} else {
				plyr.Y += (18 - p.Y + j.Y)
				lvl.lineOfSight(p.Div32())
			}
		}
		if j.X > p.X {
			if !canWalk(lvl, p.LeftN()) || plyrRootedToGrnd {
				j.X += (18 - j.X + p.X)
			} else {
				plyr.X -= (18 - j.X + p.X)
				lvl.lineOfSight(p.Div32())
			}
		} else if j.X < plyr.X {
			if !canWalk(lvl, p.RightN()) || plyrRootedToGrnd {
				j.X -= (18 - p.X + j.X)
			} else {
				plyr.X += (18 - p.X + j.X)
				lvl.lineOfSight(p.Div32())
			}
		}
	} else {
		mons.Debug2 = false
	}
}
