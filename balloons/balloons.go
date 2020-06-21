package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/t-RED-69/games-with-go/packages/noise"
	"github.com/veandco/go-sdl2/sdl"
)

var winWidht = 800
var winHeight = 600

type texture struct {
	pixels      []byte
	w, h, pitch int
}

func (tex *texture) drawBlndScal(posX, posY, scaleX, scaleY float32, pixels []byte) {
	var screeY, screeX, scrIndex, fyi int
	var fy, fx float32
	var srcR, srcG, srcB, srcA, dstR, dstG, dstB, rstR, rstG, rstB int
	newWid := int(float32(tex.w) * scaleX)
	newHt := int(float32(tex.h) * scaleY)
	texW4 := tex.w * 4
	for y := 0; y < newHt; y++ {
		fy = float32(y) / float32(newHt) * float32(tex.h-1)
		fyi = int(fy)
		screeY = int(fy*scaleY) + int(posY)
		scrIndex = screeY*winWidht*4 + int(posX)*4
		for x := 0; x < newWid; x++ {
			fx = float32(x) / float32(newWid) * float32(tex.w-1)
			screeX = int(fx*scaleX) + int(posX)
			if screeX >= 0 && screeX < winWidht && screeY >= 0 && screeY < winHeight {
				fxi4 := int(fx) * 4

				srcR = int(tex.pixels[fyi*texW4+fxi4])
				srcG = int(tex.pixels[fyi*texW4+fxi4+1])
				srcB = int(tex.pixels[fyi*texW4+fxi4+2])
				srcA = int(tex.pixels[fyi*texW4+fxi4+3])
				dstR = int(pixels[scrIndex])
				dstG = int(pixels[scrIndex+1])
				dstB = int(pixels[scrIndex+2])
				rstR = (srcR*255 + dstR*(255-srcA)) / 255
				rstG = (srcG*255 + dstG*(255-srcA)) / 255
				rstB = (srcB*255 + dstB*(255-srcA)) / 255

				pixels[scrIndex] = byte(rstR)
				scrIndex++
				pixels[scrIndex] = byte(rstG)
				scrIndex++
				pixels[scrIndex] = byte(rstB)
				scrIndex++ //skipping alpha
				scrIndex++
			}
		}
	}
}
func flerp(b1, b2, pct float32) float32 {
	return b1 + pct*(b2-b1)
}
func blerp(c00, c01, c10, c11, tx, ty float32) float32 {
	return flerp(flerp(c00, c10, tx), flerp(c01, c11, tx), ty)
}
func (tex *texture) drawBlerpScalBlnd(posX, posY, scaleX, scaleY float32, pixels []byte) {
	var screeY, screeX, scrIndex, fyi int
	var fy, fx, tx, ty float32
	var c00, c10, c01, c11 float32
	var c00i, c01i, c10i, c11i int
	var src, srcA, rst, dst int
	newWid := int(float32(tex.w) * scaleX)
	newHt := int(float32(tex.h) * scaleY)
	texW4 := tex.w * 4
	for y := 0; y < newHt; y++ {
		fy = float32(y) / float32(newHt) * float32(tex.h-1)
		fyi = int(fy)
		screeY = int(fy*scaleY) + int(posY)
		scrIndex = screeY*winWidht*4 + int(posX)*4
		ty = fy - float32(fyi)
		for x := 0; x < newWid; x++ {
			fx = float32(x) / float32(newWid) * float32(tex.w-1)
			screeX = int(fx*scaleX) + int(posX)
			if screeX >= 0 && screeX < winWidht && screeY >= 0 && screeY < winHeight {
				fxi := int(fx)
				c00i = fyi*texW4 + fxi*4
				c10i = fyi*texW4 + (fxi+1)*4
				c01i = (fyi+1)*texW4 + fxi*4
				c11i = (fyi+1)*texW4 + (fxi+1)*4
				tx = fx - float32(fxi)
				for i := 0; i < 4; i++ {
					c00 = float32(tex.pixels[c00i+i])
					c10 = float32(tex.pixels[c10i+i])
					c01 = float32(tex.pixels[c01i+i])
					c11 = float32(tex.pixels[c11i+i])

					src = int(blerp(c00, c10, c01, c11, tx, ty))
					srcA = int(blerp(float32(tex.pixels[c00i+3]), float32(tex.pixels[c10i+3]), float32(tex.pixels[c01i+3]), float32(tex.pixels[c11i+3]), tx, ty))
					dst = int(pixels[scrIndex])
					rst = (src*255 + dst*(255-srcA)) / 255
					if i != 3 {
						pixels[scrIndex] = byte(rst)
					}
					scrIndex++
				}
			}
		}
	}
}
func (tex *texture) drawBlerpScalNor(posX, posY, scaleX, scaleY float32, pixels []byte) {
	var screeY, screeX, scrIndex, fyi int
	var fy, fx, tx, ty float32
	var c00, c10, c01, c11 float32
	var c00i, c01i, c10i, c11i int
	newWid := int(float32(tex.w) * scaleX)
	newHt := int(float32(tex.h) * scaleY)
	texW4 := tex.w * 4
	for y := 0; y < newHt; y++ {
		fy = float32(y) / float32(newHt) * float32(tex.h-1)
		fyi = int(fy)
		screeY = int(fy*scaleY) + int(posY)
		scrIndex = screeY*winWidht*4 + int(posX)*4
		ty = fy - float32(fyi)
		for x := 0; x < newWid; x++ {
			fx = float32(x) / float32(newWid) * float32(tex.w-1)
			screeX = int(fx*scaleX) + int(posX)
			if screeX >= 0 && screeX < winWidht && screeY >= 0 && screeY < winHeight {
				fxi := int(fx)
				c00i = fyi*texW4 + fxi*4
				c10i = fyi*texW4 + (fxi+1)*4
				c01i = (fyi+1)*texW4 + fxi*4
				c11i = (fyi+1)*texW4 + (fxi+1)*4
				tx = fx - float32(fxi)
				for i := 0; i < 4; i++ {
					c00 = float32(tex.pixels[c00i+i])
					c10 = float32(tex.pixels[c10i+i])
					c01 = float32(tex.pixels[c01i+i])
					c11 = float32(tex.pixels[c11i+i])

					pixels[scrIndex] = byte(blerp(c00, c10, c01, c11, tx, ty))
					scrIndex++
				}
			}
		}
	}
}
func (tex *texture) draw(posX, posY float32, pixels []byte) {
	var screeY, screeX, texIndex, scrIndex int
	for y := 0; y < tex.h; y++ {
		screeY = y + int(posY)
		for x := 0; x < tex.w; x++ {
			screeX = x + int(posX)
			if screeX > 0 && screeX < winWidht && screeY > 0 && screeY < winHeight {
				texIndex = y*tex.pitch + x*4
				scrIndex = (screeY*winWidht + screeX) * 4
				pixels[scrIndex] = tex.pixels[texIndex]
				pixels[scrIndex+1] = tex.pixels[texIndex+1]
				pixels[scrIndex+2] = tex.pixels[texIndex+2]
				pixels[scrIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}
func (tex *texture) blendDraw(posX, posY float32, pixels []byte) {
	var screeY, screeX, texIndex, scrIndex int
	var srcR, srcG, srcB, srcA, rstR, rstG, rstB, dstR, dstG, dstB int
	for y := 0; y < tex.h; y++ {
		screeY = y + int(posY)
		for x := 0; x < tex.w; x++ {
			screeX = x + int(posX)
			if screeX >= 0 && screeX < winWidht && screeY >= 0 && screeY < winHeight {
				texIndex = y*tex.pitch + x*4
				scrIndex = (screeY*winWidht + screeX) * 4

				srcR = int(tex.pixels[texIndex])
				srcG = int(tex.pixels[texIndex+1])
				srcB = int(tex.pixels[texIndex+2])
				srcA = int(tex.pixels[texIndex+3])

				dstR = int(pixels[scrIndex])
				dstG = int(pixels[scrIndex+1])
				dstB = int(pixels[scrIndex+2])

				rstR = (srcR*255 + dstR*(255-srcA)) / 255
				rstG = (srcG*255 + dstG*(255-srcA)) / 255
				rstB = (srcB*255 + dstB*(255-srcA)) / 255

				pixels[scrIndex] = byte(rstR)
				pixels[scrIndex+1] = byte(rstG)
				pixels[scrIndex+2] = byte(rstB)
			}
		}
	}
}

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
	drawBlockInPixels(x0, y0, x1, y1, 0, para.rgba, pixels)
}
func drawBlockInPixels(x0, y0, x1, y1, borderInPix int, rgba rgba, pixels []byte) {
	x0 = clampInt(x0, 0, winWidht)
	y0 = clampInt(y0, 0, winHeight)
	x1 = clampInt(x1, x0, winWidht)
	y1 = clampInt(y1, y0, winHeight)

	index := y0 * winWidht
	if borderInPix == 0 {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				pixels[(index+x)*4] = rgba.r
				pixels[(index+x)*4+1] = rgba.g
				pixels[(index+x)*4+2] = rgba.b
				pixels[(index+x)*4+3] = rgba.a
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
					pixels[index+3] = 0
					continue
				}
				pixels[index] = rgba.r
				pixels[index+1] = rgba.g
				pixels[index+2] = rgba.b
				pixels[index+3] = rgba.a
			}
			index += winWidht
		}
	}
}

type object struct {
	x, y              float32
	halfLen, halfBrth float32
	rgba              rgba
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

func openBaloon(str string) *texture {
	inFile, err := os.Open(str)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()
	img, err := png.Decode(inFile)
	if err != nil {
		panic(err)
	}
	len := img.Bounds().Max.X
	bth := img.Bounds().Max.Y
	baloonPixls := make([]byte, len*bth*4)
	var index int
	var r, g, b, a uint32
	for y := 0; y < bth; y++ {
		for x := 0; x < len; x++ {
			r, g, b, a = img.At(x, y).RGBA()
			baloonPixls[index] = byte(r / 256)
			index++
			baloonPixls[index] = byte(g / 256)
			index++
			baloonPixls[index] = byte(b / 256)
			index++
			baloonPixls[index] = byte(a / 256)
			index++
		}
	}
	return &texture{baloonPixls, len, bth, len * 4}
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

	bstr := []string{"balloon_green.png", "balloon_red.png", "balloon_blue.png"}
	baloonTex := make([]*texture, len(bstr))
	pixels := make([]byte, winWidht*winHeight*4)
	noiseBG, min, max := noise.MakeNoise(noise.TURBULANCE, .0004, 2.5, 0.7, 9, winWidht, winHeight)
	getDualGradient(rgba{255, 255, 255, 0}, rgba{0, 255, 255, 0}, rgba{0, 255, 255, 0}, rgba{255, 255, 255, 0})
	backG := reScaleAndDraw(noiseBG, min, max, getDualGradient(rgba{0, 70, 205, 0}, rgba{0, 105, 155, 0}, rgba{195, 195, 195, 0}, rgba{255, 255, 255, 0}), winWidht, winHeight)
	for i, str := range bstr {
		baloonTex[i] = openBaloon(str)
	}

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		//simulate
		pixels = backG
		for i := range baloonTex {
			baloonTex[i].drawBlerpScalBlnd(float32(i*80), float32(i*50), 1+float32(i)*0.5, 1+float32(i)*0.5, pixels)
		}
		//render
		tex.Update(nil, pixels, winWidht*4)
		renderer.Present()
		renderer.Copy(tex, nil, nil)
		//
		sdl.Delay(20)
	}
}

type rgba struct {
	r, g, b, a byte
}

func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}
func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct), 0}
}
func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}
func getDualGradient(c1, c2, c3, c4 rgba) []rgba {
	result := make([]rgba, 256)
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
func reScaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
	pixels := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale
	var p int
	var c rgba
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
