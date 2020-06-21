package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

var winWidht = 800
var winHeight = 600

type color struct {
	r, g, b byte
}

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidht + x) * 4
	if index <= len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
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

	for y := 0; y < winWidht; y++ {
		for x := 0; x < winWidht; x++ {
			setPixel(x,y,color{230,110,40},pixels)
		}	
	}
	tex.Update(nil,pixels,winWidht*4)
	renderer.Copy(tex,nil,nil)
	renderer.Present()

	sdl.Delay(2000)
}
