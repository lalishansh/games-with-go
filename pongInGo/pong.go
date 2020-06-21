package main

import (
	"fmt"

	"github.com/t-RED-69/games-with-go/packages/noise"
	"github.com/veandco/go-sdl2/sdl"
)

var winWidht = 800
var winHeight = 600

const renderScale = 0.01

func (para object) drawBlock(pixels []byte) {
	para.x *= float32(winHeight) * renderScale
	para.y *= float32(winHeight) * renderScale
	para.halfLen *= float32(winHeight) * renderScale
	para.halfBrth *= float32(winHeight) * renderScale

	para.x += float32(winWidht) * 0.5
	para.y += float32(winHeight) * 0.5

	x0 := int(para.x - para.halfLen)
	x1 := int(para.x + para.halfLen)
	y0 := int(para.y - para.halfBrth)
	y1 := int(para.y + para.halfBrth)
	drawBlockInPixels(x0, y0, x1, y1, 0, para.color, pixels)
}
func drawBlockInPixels(x0, y0, x1, y1, borderInPix int, color color, pixels []byte) {
	x0 = clampInt(x0, 0, winWidht)
	y0 = clampInt(y0, 0, winHeight)
	x1 = clampInt(x1, x0, winWidht)
	y1 = clampInt(y1, y0, winHeight)

	index := y0 * winWidht
	if borderInPix == 0 {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				pixels[(index+x)*4] = color.r
				pixels[(index+x)*4+1] = color.g
				pixels[(index+x)*4+2] = color.b
			}
			index += winWidht
		}
	} else {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if y < (y0+borderInPix) || y > (y1-borderInPix) || x < (x0+borderInPix) || x > (x1-borderInPix) {
					pixels[index] = 0
					pixels[index+1] = 0
					pixels[index+2] = 0
					continue
				}
				pixels[index] = color.r
				pixels[index+1] = color.g
				pixels[index+2] = color.b
			}
			index += winWidht
		}
	}
}
func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

type color struct {
	r, g, b byte
}
type object struct {
	x, y              float32
	halfLen, halfBrth float32
	color             color
}
type element struct {
	object
	speedX, speedY float32
	accln          int8
}

func clearScreen(color color, pixels []byte) {
	var index int = 0
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidht; x++ {
			pixels[(index+x)*4] = color.r
			pixels[(index+x)*4+1] = color.g
			pixels[(index+x)*4+2] = color.b
		}
		index += winWidht
	}
}

func main() {
	window, err := sdl.CreateWindow("Test sdl2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidht), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy() //defer executes this statement after reaching the end of function/finishing the execution of funtion

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidht), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, winWidht*winHeight*4)
	ball := element{object{0, 0, 1, 1, color{0, 0, 255}}, 50, 50, 3}
	paddleL := element{object{-50, 0, 1, 10, color{255, 0, 0}}, 0, 50, 10}
	paddleR := element{object{50, 0, 1, 10, color{0, 255, 0}}, 0, 50, 10}

	noisec, min, max := noise.MakeNoise(noise.TURBULANCE, .0005, 2.5, 1, 9, winWidht, winHeight)
	noisePixels := reScaleAndDraw(noisec, min, max, getGradient(color{255, 255, 0}, color{0, 0, 0}), winWidht, winHeight)
	for i := range noisePixels {
		pixels[i] = noisePixels[i]
	}
	ball.drawBlock(pixels)
	paddleL.drawBlock(pixels)
	paddleR.drawBlock(pixels)
	tex.Update(nil, pixels, winWidht*4)
	renderer.Copy(tex, nil, nil)
	renderer.Present()

	sdl.Delay(5000)
}
func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}
func colorLerp(c1, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}
func getGradient(c1, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}
func getDualGradient(c1, c2, c3, c4 color) []color {
	result := make([]color, 256)
	var pct float32
	for i := range result {
		pct = float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*2)
		} else {
			pct -= 0.5
			result[i] = colorLerp(c3, c4, pct*2)
		}
	}
	return result
}
func reScaleAndDraw(noise []float32, min, max float32, gradient []color, w, h int) []byte {
	pixels := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale
	var p int
	var c color
	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c = gradient[(clampInt(int(noise[i]), 0, 255))]
		p = i * 4
		pixels[p] = c.r
		pixels[p+1] = c.g
		pixels[p+2] = c.b
		pixels[p+3] = 0
		//fmt.Println(noise[i])
	}
	return pixels
}
