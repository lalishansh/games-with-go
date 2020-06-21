package main

import (
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/t-RED-69/games-with-go/packages/noise"
	"github.com/t-RED-69/games-with-go/packages/vec3"
	"github.com/veandco/go-sdl2/sdl"
)

var winWidht = 1000
var winHeight = 700
var balloonStr = []string{"balloon_green.png", "balloon_red.png", "balloon_blue.png"}

type audioStat struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type mouseState struct {
	left, right        bool
	changedL, changedR bool
	x, y               int32
}

func (m *mouseState) processMouse(x, y int32, mouse uint32) {
	m.x, m.y = x, y
	currL := (mouse&sdl.ButtonLMask() == 1)
	currR := (mouse&sdl.ButtonRMask() == 4)
	if m.left != currL {
		m.changedL = true
	} else {
		m.changedL = false
	}
	if m.right != currR {
		m.changedR = true
	} else {
		m.changedR = false
	}
	m.left = currL
	m.right = currR
}

func sqrMag(a float32) float32 {
	return a * a
}

type textur struct {
	tex      *sdl.Texture
	len, bth int
}

type attribute struct {
	posn, dirn     vec3.Vector3
	orignalZ       float32
	scale          float32
	balloonTyp     int
	explode        bool
	timSincExplosn float32
}

func imgToTex(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func (tex *textur) draw(obj *attribute, scale float32, renderer *sdl.Renderer) {
	newWid := int32(float32(tex.len) * scale)
	newHt := int32(float32(tex.bth) * scale)
	x := int32(obj.posn.X - float32(newWid)/2)
	y := int32(obj.posn.Y - float32(newHt)/2)
	rect := &sdl.Rect{x, y, newWid, newHt}
	renderer.Copy(tex.tex, nil, rect)
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
func clampFlt(val *float32, min, max float32) {
	if *val < min {
		*val = min
	}
	if *val > max {
		*val = max
	}
}
func instancImgWthRandMov(renderer *sdl.Renderer, noOfballoons int, balloonStr []string, autoDefineTyp bool) (balloonAtt *[]attribute, baloonTx *[]*textur, maxScl float32) {
	balloonTyps := len(balloonStr)
	baloonTex := make([]*textur, len(balloonStr))
	balloonAttr := make([]attribute, noOfballoons)
	for i, str := range balloonStr {
		baloonTex[i] = openBaloon(str, renderer)
	}
	for i := 0; i < noOfballoons; i++ {
		balloonAttr[i].posn.Z = rand.Float32() * 1000
		if autoDefineTyp {
			balloonAttr[i].balloonTyp = i % balloonTyps
		}
	}
	// sorting
	for i := 0; i < noOfballoons; i++ {
		for j := i + 1; j < noOfballoons; j++ {
			if balloonAttr[i].posn.Z > balloonAttr[j].posn.Z {
				balloonAttr[i].posn.Z, balloonAttr[j].posn.Z = balloonAttr[j].posn.Z, balloonAttr[i].posn.Z
			}
		}
		balloonAttr[i].orignalZ = balloonAttr[i].posn.Z
		clampFlt(&balloonAttr[i].orignalZ, 100, 900)
	}
	for i := 0; i < noOfballoons; i++ {
		balloonAttr[i].posn = vec3.Vector3{rand.Float32() * float32(winWidht), rand.Float32() * float32(winHeight), balloonAttr[i].posn.Z}
		balloonAttr[i].scale = 1
		balloonAttr[i].dirn = vec3.Vector3{1 + rand.Float32()*5, 1 + rand.Float32()*5, 0.6 + rand.Float32()/2}
	}
	return &balloonAttr, &baloonTex, maxScl
}
func spriteOpener(renderer *sdl.Renderer, str string, noOfColumn, noOfRow int) []*textur {
	inFile, err := os.Open(str)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()
	img, err := png.Decode(inFile)
	if err != nil {
		panic(err)
	}
	len := img.Bounds().Max.X / noOfColumn
	bth := img.Bounds().Max.Y / noOfRow
	var index int
	var r, g, b, a uint32
	spriteArray := make([]*textur, noOfColumn*noOfRow)
	var tex *sdl.Texture
	for i := 0; i < noOfRow; i++ {
		for j := 0; j < noOfColumn; j++ {
			explnPixls := make([]byte, len*bth*4)
			index = 0
			for y := bth * i; y < bth*(i+1); y++ {
				for x := len * j; x < len*(j+1); x++ {
					r, g, b, a = img.At(x, y).RGBA()
					explnPixls[index] = byte(r / 256)
					index++
					explnPixls[index] = byte(g / 256)
					index++
					explnPixls[index] = byte(b / 256)
					index++
					explnPixls[index] = byte(a / 256)
					index++
				}
			}
			tex = imgToTex(renderer, explnPixls, len, bth)
			err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
			if err != nil {
				panic(err)
			}
			spriteArray[i*noOfRow+j] = &textur{tex, len, bth}
		}
	}
	return spriteArray
}
func openBaloon(str string, renderer *sdl.Renderer) *textur {
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
	tex := imgToTex(renderer, baloonPixls, len, bth)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return &textur{tex, len, bth}
}
func (mouse *mouseState) updateBalloons(noOfballoons *int, balloonAttr *[]attribute, baloonTex *[]*textur, sprtArr *[]*textur, renderer *sdl.Renderer, maxScl, deltaTim float32, audState *audioStat) {
	destroy := false
	for i := 0; i < *noOfballoons; i++ {
		((*balloonAttr)[i].posn).Plus((*balloonAttr)[i].dirn)
		if (*balloonAttr)[i].posn.X > float32(winWidht) || (*balloonAttr)[i].posn.X < 0 {
			if (*balloonAttr)[i].posn.X > float32(winWidht+50) && (*balloonAttr)[i].dirn.X > 0 {
				(*balloonAttr)[i].dirn.X *= -1
			} else if (*balloonAttr)[i].posn.X < -50 && (*balloonAttr)[i].dirn.X < 0 {
				(*balloonAttr)[i].dirn.X *= -1
			}
		}
		if (*balloonAttr)[i].posn.Y > float32(winHeight) || (*balloonAttr)[i].posn.Y < 0 {
			if (*balloonAttr)[i].posn.Y > float32(winHeight+50) && (*balloonAttr)[i].dirn.Y > 0 {
				(*balloonAttr)[i].dirn.Y *= -1
			} else if (*balloonAttr)[i].posn.Y < -50 && (*balloonAttr)[i].dirn.Y < 0 {
				(*balloonAttr)[i].dirn.Y *= -1
			}
		}
		if (*balloonAttr)[i].posn.Z > (*balloonAttr)[i].orignalZ+50 || (*balloonAttr)[i].posn.Z < (*balloonAttr)[i].orignalZ-50 {
			if (*balloonAttr)[i].posn.Z > (*balloonAttr)[i].orignalZ+50 && (*balloonAttr)[i].dirn.Z > 0 {
				(*balloonAttr)[i].dirn.Z *= -1
			} else if (*balloonAttr)[i].posn.Z < (*balloonAttr)[i].orignalZ-50 && (*balloonAttr)[i].dirn.Z < 0 {
				(*balloonAttr)[i].dirn.Z *= -1
			}
		}
		for j := i + 1; j < *noOfballoons; j++ {
			if (*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ) > (*balloonAttr)[j].scale*((*balloonAttr)[j].posn.Z/(*balloonAttr)[j].orignalZ) {
				(*balloonAttr)[i], (*balloonAttr)[j] = (*balloonAttr)[j], (*balloonAttr)[i]
			}
		}

		if (*balloonAttr)[i].explode {
			(*balloonAttr)[i].timSincExplosn += deltaTim
			//fmt.Println((*balloonAttr)[i].timSincExplosn)
			if (*balloonAttr)[i].explode && (*balloonAttr)[i].timSincExplosn > 400 {
				destroy = true
			} else if (*balloonAttr)[i].timSincExplosn > 375 {
				(*sprtArr)[0].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 350 {
				(*sprtArr)[1].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 325 {
				(*sprtArr)[2].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 300 {
				(*sprtArr)[3].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 275 {
				(*sprtArr)[4].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 250 {
				(*sprtArr)[5].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 225 {
				(*sprtArr)[6].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 200 {
				(*sprtArr)[7].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 175 {
				(*sprtArr)[8].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 150 {
				(*sprtArr)[9].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 125 {
				(*sprtArr)[10].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 100 {
				(*sprtArr)[11].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 75 {
				(*sprtArr)[12].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 50 {
				(*sprtArr)[13].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 25 {
				(*sprtArr)[14].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			} else if (*balloonAttr)[i].timSincExplosn > 0 {
				(*sprtArr)[15].draw(&(*balloonAttr)[i], 4*(*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
			}
		} else {
			(*baloonTex)[(*balloonAttr)[i].balloonTyp].draw(&(*balloonAttr)[i], (*balloonAttr)[i].scale*((*balloonAttr)[i].posn.Z/(*balloonAttr)[i].orignalZ), renderer)
		}
	} //never use routines with pointers
	if mouse.changedL || mouse.changedR {
		for i := *noOfballoons - 1; i >= 0; i-- {
			if mouse.changedL && mouse.left {
				rSq := sqrMag(102 * (*balloonAttr)[i].scale * ((*balloonAttr)[i].posn.Z / (*balloonAttr)[i].orignalZ))
				xSq := sqrMag(float32(mouse.x) - (*balloonAttr)[i].posn.X)
				ySq := sqrMag(float32(mouse.y) - (*balloonAttr)[i].posn.Y)
				if xSq+ySq < rSq {
					(*balloonAttr)[i].explode = true
					sdl.ClearQueuedAudio(audState.deviceID)
					sdl.QueueAudio(audState.deviceID, audState.explosionBytes)
					sdl.PauseAudioDevice(audState.deviceID, false)
					break
				}
			}
			if mouse.changedR && mouse.right {
				if len(*balloonAttr) > *noOfballoons {
					a, _, _ := instancImgWthRandMov(renderer, 1, balloonStr, false)
					(*balloonAttr)[*noOfballoons].posn.X = float32(mouse.x)
					(*balloonAttr)[*noOfballoons].posn.Y = float32(mouse.y)
					(*balloonAttr)[*noOfballoons].posn.Z = (*a)[0].posn.Z
					(*balloonAttr)[*noOfballoons].dirn = (*a)[0].dirn
					(*balloonAttr)[*noOfballoons].orignalZ = (*a)[0].orignalZ
					(*balloonAttr)[*noOfballoons].scale = (*a)[0].scale
					(*balloonAttr)[*noOfballoons].explode = false
					(*balloonAttr)[*noOfballoons].timSincExplosn = 0
					*noOfballoons++
					break
				}
			}
		}
	}
	if destroy {
		for i := *noOfballoons - 1; i >= 0; i-- {
			if (*balloonAttr)[i].explode && (*balloonAttr)[i].timSincExplosn > 400 {
				//explode balloon
				(*balloonAttr)[*noOfballoons-1].balloonTyp = (*balloonAttr)[i].balloonTyp
				for j := i; j < *noOfballoons-1; j++ {
					(*balloonAttr)[j] = (*balloonAttr)[j+1]
				}
				*noOfballoons--
				break
			}
		}
		fmt.Println(*noOfballoons)
	}
}
func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()
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
	var deltaTim float32

	explosionBytes, audioSpec := sdl.LoadWAV("explode.wav")
	audioID, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	if err != nil {
		fmt.Println(err)
	}
	defer sdl.FreeWAV(explosionBytes)

	audioState := audioStat{explosionBytes, audioID, audioSpec}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	pixels := make([]byte, winWidht*winHeight*4)
	var mouse mouseState
	mouse.processMouse(sdl.GetMouseState())

	//
	noOfballoons := 25
	balloonAttr, baloonTex, maxScl := instancImgWthRandMov(renderer, noOfballoons, balloonStr, true)

	explImg := "explosion.png"
	spriteArray := spriteOpener(renderer, explImg, 4, 4)

	noiseBG, min, max := noise.MakeNoise(noise.TURBULANCE, .0004, 2.5, 0.7, 9, winWidht, winHeight)
	getDualGradient(rgba{255, 255, 255, 0}, rgba{0, 255, 255, 0}, rgba{0, 255, 255, 0}, rgba{255, 255, 255, 0})
	cloudBG := reScaleAndDraw(noiseBG, min, max, getDualGradient(rgba{0, 70, 205, 0}, rgba{0, 105, 155, 0}, rgba{195, 195, 195, 0}, rgba{255, 255, 255, 0}), winWidht, winHeight)
	cloudTex := imgToTex(renderer, cloudBG, winWidht, winHeight)
	for {
		tim := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					mouse.x = int32(e.X * float32(winWidht))
					mouse.y = int32(e.Y * float32(winHeight))
					if !mouse.left {
						mouse.changedL = true
					} else {
						mouse.changedL = false
					}
					mouse.left = true
				} else {
					mouse.left = false
					mouse.changedL = false
				}
			}
		}
		//mouseStat:=processMouse()
		mouse.processMouse(sdl.GetMouseState())
		//fmt.Println(mouse)
		//simulate
		renderer.Copy(cloudTex, nil, nil)
		mouse.updateBalloons(&noOfballoons, balloonAttr, baloonTex, &spriteArray, renderer, maxScl, deltaTim, &audioState)

		//render
		tex.Update(nil, pixels, winWidht*4)
		renderer.Present()
		renderer.Copy(tex, nil, nil)
		deltaTim = float32(time.Since(tim).Seconds() * 1000)
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
