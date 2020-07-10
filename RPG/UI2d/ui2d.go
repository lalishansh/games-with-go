package ui2d

import (
	"bufio"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/t-RED-69/games-with-go/RPG/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ui struct {
	winWidht, winHeight int32
	renderer            *sdl.Renderer
	window              *sdl.Window
	tex                 *sdl.Texture
	zoom                int32
	centerX, centerY    int32
	textureAtlas        *[]SpriteTexture
	MiniAtlas           *[]SpriteTexture
	mouse               MouseState
	keyBoard            []KeyStates
	r                   *rand.Rand
	levelChan           chan *game.Level
	inputChan           chan *game.Input
	fontSmall           *ttf.Font
	fontMed             *ttf.Font
	fontLarg            *ttf.Font
	//
	str2TexSmll map[string]*sdl.Texture
	str2TexMed  map[string]*sdl.Texture
	str2TexLarg map[string]*sdl.Texture
	helloWorld  *sdl.Texture
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.winWidht, ui.winHeight = 1280, 720
	ui.zoom = 3
	ui.str2TexSmll = make(map[string]*sdl.Texture)
	ui.str2TexMed = make(map[string]*sdl.Texture)
	ui.str2TexLarg = make(map[string]*sdl.Texture)
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.r = rand.New(rand.NewSource(1))
	window, err := sdl.CreateWindow("RPG !!", int32(1366/2-ui.winWidht/2), int32(766/2-ui.winHeight/2), int32(ui.winWidht), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window
	//defer window.Destroy() //defer executes this statement after reaching the end of function/finishing the execution of funtion
	//and we dont wanna destroy it

	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas = ui.SpriteOpener("UI2d/assets/tiles.png", 32, 32, 6042)
	ui.MiniAtlas = ui.idexAssignerToAtlas()

	ui.keyBoard = make([]KeyStates, len(sdl.GetKeyboardState()))
	ui.mouse.ProcessMouse()
	ProcessKeys(&ui.keyBoard)
	if ui.fontSmall, err = ttf.OpenFont("UI2d/assets/Kingthings_Foundation.ttf", 18); err != nil {
		fmt.Println("cannot open font: UI2d/assets/Kingthings_Foundation.ttf")
	}
	if ui.fontMed, err = ttf.OpenFont("UI2d/assets/Kingthings_Foundation.ttf", 32); err != nil {
		fmt.Println("cannot open font: UI2d/assets/Kingthings_Foundation.ttf")
	}
	if ui.fontLarg, err = ttf.OpenFont("UI2d/assets/Kingthings_Foundation.ttf", 48); err != nil {
		fmt.Println("cannot open font: UI2d/assets/Kingthings_Foundation.ttf")
	}
	fontSurface, err := ui.fontMed.RenderUTF8Blended("Hello world", sdl.Color{255, 0, 0, 100})
	if err != nil {
		panic(err)
	}
	ui.helloWorld, err = ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}
	//
	return ui
}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

func (ui *ui) stringToTexture(s string, clor sdl.Color, size FontSize) *sdl.Texture {
	var tex *sdl.Texture
	var font *ttf.Font
	switch size {
	case FontSmall:
		tex, exist := ui.str2TexSmll[s]
		if exist {
			return tex
		}
		font = ui.fontSmall
	case FontMedium:
		tex, exist := ui.str2TexMed[s]
		if exist {
			return tex
		}
		font = ui.fontMed
	case FontLarge:
		tex, exist := ui.str2TexLarg[s]
		if exist {
			return tex
		}
		font = ui.fontLarg
	}
	fontSurface, err := font.RenderUTF8Blended(s, clor)
	if err != nil {
		panic(err)
	}
	tex, err = ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}
	switch size {
	case FontSmall:
		ui.str2TexSmll[s] = tex
	case FontMedium:
		ui.str2TexMed[s] = tex
	case FontLarge:
		ui.str2TexLarg[s] = tex
	}
	return tex
}

type MouseState struct {
	Left, Right        bool
	ChangedL, ChangedR bool
	X, Y               int32
}
type KeyStates struct {
	IsDown  bool
	Changed bool
}

func (m *MouseState) ProcessMouse() {
	x, y, mouse := sdl.GetMouseState()
	m.X, m.Y = x, y
	currL := (mouse&sdl.ButtonLMask() == 1)
	currR := (mouse&sdl.ButtonRMask() == 4)
	if m.Left != currL {
		m.ChangedL = true
	} else {
		m.ChangedL = false
	}
	if m.Right != currR {
		m.ChangedR = true
	} else {
		m.ChangedR = false
	}
	m.Left = currL
	m.Right = currR
}
func ProcessKeys(kb *[]KeyStates) {
	keystrokes := sdl.GetKeyboardState()
	for i := range *kb {
		if (*kb)[i].IsDown != (keystrokes[i] != 0) {
			(*kb)[i].Changed = true
		} else {
			(*kb)[i].Changed = false
		}
		(*kb)[i].IsDown = (keystrokes[i] != 0)
	}
}

//SpriteTexture cantains sprite's enum name,texture,default length and breadth for image
type SpriteTexture struct {
	symbol   game.Tile
	varCount int
	index    int
	tex      *sdl.Texture
	len, bth int32
}

//SpriteOpener to load specified number of sprite textures
func (ui *ui) SpriteOpener(str string, lenPerSprite, widPerSprite int32, noOfSprites int) *[]SpriteTexture {
	inFile, err := os.Open(str)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	img, err := png.Decode(inFile)
	if err != nil {
		panic(err)
	}

	noOfColumn := int32(img.Bounds().Max.X / int(lenPerSprite))
	noOfRow := int32(int(float32(noOfSprites)/float32(noOfColumn)) + 1)
	var index int
	var r, g, b, a uint32
	spriteArray := make([]SpriteTexture, noOfSprites)
	var tex *sdl.Texture
	var i, j, x, y int32
	var counter, counter2 int
	for i = 0; i < noOfRow; i++ {
		for j = 0; j < noOfColumn; j++ {
			counter2++
			pixels := make([]byte, lenPerSprite*widPerSprite*4)
			index = 0
			for y = widPerSprite * i; y < widPerSprite*(i+1); y++ {
				for x = lenPerSprite * j; x < lenPerSprite*(j+1); x++ {
					r, g, b, a = img.At(int(x), int(y)).RGBA()
					pixels[index] = byte(r / 256)
					index++
					pixels[index] = byte(g / 256)
					index++
					pixels[index] = byte(b / 256)
					index++
					pixels[index] = byte(a / 256)
					index++
				}
			}
			tex, err = ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, lenPerSprite, widPerSprite)
			if err != nil {
				panic(err)
			}
			tex.Update(nil, pixels, int(lenPerSprite)*4)
			err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
			if err != nil {
				panic(err)
			}
			if (i*noOfColumn + j) < int32(noOfSprites) {
				spriteArray[i*noOfColumn+j] = SpriteTexture{' ', 0, 0, tex, lenPerSprite, widPerSprite}
				counter++
			} else {
				break
			}
		}
	}
	return &spriteArray
}

func init() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ttf.Init()
	if err != nil {
		fmt.Println(err)
		return
	}
}

//Draw to draw over screen
func (ui *ui) Draw(level *game.Level) {
	if (level.Player.X*ui.zoom - ui.centerX) > (ui.winWidht/2 + 64*ui.zoom) {
		ui.centerX += 3 * ui.zoom
	} else if (level.Player.X*ui.zoom - ui.centerX) < (ui.winWidht/2 - 64*ui.zoom) {
		ui.centerX -= 3 * ui.zoom
	} else if (level.Player.X*ui.zoom - ui.centerX) > (ui.winWidht / 2) {
		ui.centerX++
	} else if (level.Player.X*ui.zoom - ui.centerX) < (ui.winWidht / 2) {
		ui.centerX--
	}
	if (level.Player.Y*ui.zoom - ui.centerY) > (ui.winHeight/2 + 55*ui.zoom) {
		ui.centerY += 3 * ui.zoom
	} else if (level.Player.Y*ui.zoom - ui.centerY) < (ui.winHeight/2 - 55*ui.zoom) {
		ui.centerY -= 3 * ui.zoom
	} else if (level.Player.Y*ui.zoom - ui.centerY) > (ui.winHeight / 2) {
		ui.centerY++
	} else if (level.Player.Y*ui.zoom - ui.centerY) < (ui.winHeight / 2) {
		ui.centerY--
	}

	ui.renderer.Clear()
	ui.r.Seed(1)
	for y, row := range level.Map {
		var r int
		for x, tile := range row {
			dstRect := sdl.Rect{int32(x*32)*ui.zoom - ui.centerX, int32(y*32)*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom}
			//pos := game.Pos{int32(x), int32(y)}
			for t := range *ui.MiniAtlas {
				if tile == (*ui.MiniAtlas)[t].symbol {
					r = ui.r.Intn((*ui.MiniAtlas)[t].varCount)
					//if level.Debug[pos] {
					//	(*ui.MiniAtlas)[t+r].tex.SetColorMod(128, 0, 0)
					//} else {
					//	(*ui.MiniAtlas)[t+r].tex.SetColorMod(255, 255, 255)
					//}
					switch tile {
					case game.OpenDoor:
						ui.renderer.Copy((*ui.MiniAtlas)[12].tex, nil, &dstRect)
					}
					ui.renderer.Copy((*ui.MiniAtlas)[t+r].tex, nil, &dstRect)
					break
				}
			}
		}
	}
	for t := range *ui.MiniAtlas {
		if level.Player.Symbol == (*ui.MiniAtlas)[t].symbol {
			ui.renderer.Copy((*ui.MiniAtlas)[t].tex, nil, &sdl.Rect{level.Player.X*ui.zoom - ui.centerX, level.Player.Y*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom})
			break
		}
	}
	for i := range level.Monsters {
		dstRect := sdl.Rect{int32(level.Monsters[i].X)*ui.zoom - ui.centerX, int32(level.Monsters[i].Y)*ui.zoom - ui.centerY, 32 * ui.zoom, 32 * ui.zoom}
		for t := range *ui.MiniAtlas {
			if level.Monsters[i].Symbol == (*ui.MiniAtlas)[t].symbol {
				ui.renderer.Copy((*ui.MiniAtlas)[t].tex, nil, &dstRect)
				break
			}
		}
	}
	//
	for i := range level.Events {
		tex := ui.stringToTexture(level.Events[i], sdl.Color{255, 0, 0, 0}, FontSmall)
		_, _, w, h, _ := tex.Query()
		ui.renderer.Copy(tex, nil, &sdl.Rect{10, ui.winHeight - int32((i+1)*22), w, h})
	}
	//
	ui.renderer.Present()
	sdl.Delay(10)
}

func (ui *ui) idexAssignerToAtlas() *[]SpriteTexture {
	file, err := os.Open("UI2d/assets/tileSymbol-Index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	newAtlas := make([]SpriteTexture, 0)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := game.Tile(line[0])
		xyv := line[1:]
		splitXYV := strings.Split(xyv, ",")
		x, err := strconv.ParseInt(splitXYV[0], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(splitXYV[1], 10, 64)
		if err != nil {
			panic(err)
		}
		v, err := strconv.ParseInt(splitXYV[2], 10, 64)
		if err != nil {
			panic(err)
		}
		var z int64
		for z = 0; z < v; z++ {
			(*ui.textureAtlas)[y*64+(x+z)].symbol = tileRune
			(*ui.textureAtlas)[y*64+(x+z)].varCount = int(v)
			(*ui.textureAtlas)[y*64+(x+z)].index = int(z)
			newAtlas = append(newAtlas, (*ui.textureAtlas)[y*64+(x+z)])
		}
	}
	return &newAtlas
}
func (ui *ui) Run() {
	var lvle *game.Level
	for {
		ui.mouse.ProcessMouse()
		ProcessKeys(&ui.keyBoard)
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChannel: ui.levelChan}
				}
			}
		}
		if lvle != nil {
			ui.Draw(lvle)
		}
		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
				lvle = newLevel
			}
		default:
		}
		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			input := &game.Input{Typ: game.Blank}
			if ui.keyBoard[sdl.SCANCODE_DOWN].IsDown {
				input = &game.Input{Typ: game.Down}
			} else if ui.keyBoard[sdl.SCANCODE_UP].IsDown {
				input = &game.Input{Typ: game.Up}
			} else if ui.keyBoard[sdl.SCANCODE_LEFT].IsDown {
				input = &game.Input{Typ: game.Left}
			} else if ui.keyBoard[sdl.SCANCODE_RIGHT].IsDown {
				input = &game.Input{Typ: game.Right}
			} else if ui.keyBoard[sdl.SCANCODE_O].Changed && ui.keyBoard[sdl.SCANCODE_O].IsDown {
				input = &game.Input{Typ: game.Open}
			} else if ui.keyBoard[sdl.SCANCODE_SPACE].IsDown {
				input = &game.Input{Typ: game.Open}
			} else if ui.keyBoard[sdl.SCANCODE_S].Changed && ui.keyBoard[sdl.SCANCODE_S].IsDown {
				fmt.Println("search")
				input = &game.Input{Typ: game.Search}
			} else if ui.keyBoard[sdl.SCANCODE_KP_PLUS].Changed && ui.keyBoard[sdl.SCANCODE_KP_PLUS].IsDown {
				ui.zoom++
			} else if ui.keyBoard[sdl.SCANCODE_KP_MINUS].Changed && ui.keyBoard[sdl.SCANCODE_KP_MINUS].IsDown {
				ui.zoom--
			}
			if input.Typ != game.Blank {
				ui.inputChan <- input
			}
		}
		sdl.Delay(12)
	}
}
