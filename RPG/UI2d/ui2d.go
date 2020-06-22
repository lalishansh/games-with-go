package ui2d

import (
	"bufio"
	"fmt"
	"github.com/t-RED-69/games-with-go/RPG/game"
	"github.com/veandco/go-sdl2/sdl"
	"image/png"
	"os"
	"strconv"
	"strings"
)

const winWidht, winHeight = 1280, 700

var renderer *sdl.Renderer
var tex *sdl.Texture

//SpriteTexture cantains sprite's enum name,texture,default length and breadth for image
type SpriteTexture struct {
	symbol   game.Tile
	tex      *sdl.Texture
	len, bth int32
}

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
	fmt.Println(noOfColumn, noOfRow)
	var index int
	var r, g, b, a uint32
	spriteArray := make([]SpriteTexture, noOfSprites)
	var tex *sdl.Texture
	var i, j, x, y int32
	for i = 0; i < noOfRow; i++ {
		for j = 0; j < noOfColumn; j++ {
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
			if (i*noOfRow + j) < int32(noOfSprites) {
				spriteArray[i*noOfRow+j] = SpriteTexture{' ', tex, lenPerSprite, widPerSprite}
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
		xy := line[1:]
		splitXY := strings.Split(xy, ",")
		x, err := strconv.ParseInt(splitXY[0], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(splitXY[1], 10, 64)
		if err != nil {
			panic(err)
		}
		(*textureAtlas)[(y-1)*64+x-1].symbol = tileRune
		newAtlas = append(newAtlas, (*textureAtlas)[(y-1)*64+x-1])
	}
	return &newAtlas
}

//Draw to draw over screen
func (ui *UI2d) Draw(level *game.Level) {
	textureAtlas = SpriteOpener(renderer, "UI2d/assets/tiles.png", 32, 32, 6042)
	MiniAtlas = idexAssignerToAtlas()
	//renderer.Copy(tex, nil, nil)
	//renderer.Present()
}
