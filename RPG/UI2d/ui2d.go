package ui2d

import (
	"bufio"
	"github.com/t-RED-69/games-with-go/RPG/game"
	"github.com/veandco/go-sdl2/sdl"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type MouseState struct {
	Left, Right        bool
	ChangedL, ChangedR bool
	X, Y               int32
}
type KeyStates struct {
	IsDown  bool
	Changed bool
}

var zoom float32 = 1

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

const winWidht, winHeight = 1280, 700

var renderer *sdl.Renderer
var tex *sdl.Texture

var mouse MouseState
var keyBoard []KeyStates

//SpriteTexture cantains sprite's enum name,texture,default length and breadth for image
type SpriteTexture struct {
	symbol   game.Tile
	varCount int
	index    int
	tex      *sdl.Texture
	len, bth int32
}

type Entity struct {
	x, y int
}
type Player struct {
	Entity
	Texture *sdl.Texture
}

var Player1 Player
var textureAtlas *[]SpriteTexture
var MiniAtlas *[]SpriteTexture

//SpriteOpener to load specified number of sprite textures
func SpriteOpener(renderer *sdl.Renderer, str string, lenPerSprite, widPerSprite int32, noOfSprites int) *[]SpriteTexture {
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
			//tex = imgToTex(renderer, pixels, lenPerSprite, widPerSprite)
			tex, err = renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, lenPerSprite, widPerSprite)
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
func imgToTex(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

type UI2d struct {
}

func init() {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	window, err := sdl.CreateWindow("RPG !!", int32(1366/2-winWidht/2), int32(766/2-winHeight/2), int32(winWidht), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	//defer window.Destroy() //defer executes this statement after reaching the end of function/finishing the execution of funtion
	//and we dont wanna destroy it

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	textureAtlas = SpriteOpener(renderer, "UI2d/assets/tiles.png", 32, 32, 6042)
	MiniAtlas = idexAssignerToAtlas()

	Player1.x, Player1.y = 84, 84
	Player1.Texture = (*textureAtlas)[59*64+25].tex

	keyBoard = make([]KeyStates, len(sdl.GetKeyboardState()))
	UpdateKeys()

}
func idexAssignerToAtlas() *[]SpriteTexture {
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
			(*textureAtlas)[y*64+(x+z)].symbol = tileRune
			(*textureAtlas)[y*64+(x+z)].varCount = int(v)
			(*textureAtlas)[y*64+(x+z)].index = int(z)
			newAtlas = append(newAtlas, (*textureAtlas)[y*64+(x+z)])
		}
	}
	return &newAtlas
}
func UpdateKeys() {
	mouse.ProcessMouse()
	ProcessKeys(&keyBoard)
}
func UpdatePlayer(level *game.Level) {
	//if level.Map[int(Player1.y/(32*zoom))][int(Player1.x/(32*zoom))]=='.'{
	//}
	if keyBoard[sdl.SCANCODE_DOWN].IsDown {
		Player1.y += 1
	} else if keyBoard[sdl.SCANCODE_UP].IsDown {
		Player1.y -= 1
	} else if keyBoard[sdl.SCANCODE_RIGHT].IsDown {
		Player1.x += 1
	} else if keyBoard[sdl.SCANCODE_LEFT].IsDown {
		Player1.x -= 1
	}
}

//Draw to draw over screen
func (ui *UI2d) Draw(level *game.Level) {
	UpdateKeys()
	for y, row := range level.Map {
		var r int
		for x, tile := range row {
			dstRect := sdl.Rect{int32(x * 32), int32(y * 32), 32, 32}
			for t := range *MiniAtlas {
				if tile == (*MiniAtlas)[t].symbol {
					r = rand.Intn((*MiniAtlas)[t].varCount)
					renderer.Copy((*MiniAtlas)[t+r].tex, nil, &dstRect)
					break
				}
			}
		}
	}
	UpdatePlayer(level)
	if keyBoard[sdl.SCANCODE_ESCAPE].IsDown && keyBoard[sdl.SCANCODE_ESCAPE].Changed {
		return
	}
	sdl.Delay(100)
	renderer.Copy(Player1.Texture, nil, &sdl.Rect{int32(Player1.x), int32(Player1.y), 32, 32})
	renderer.Present()
}
