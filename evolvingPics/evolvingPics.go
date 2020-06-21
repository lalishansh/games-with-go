package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/t-RED-69/games-with-go/packages/apt"
	"github.com/t-RED-69/games-with-go/packages/gui"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidht, winHeight = 1200, 700

var cols, rows, numPics = 3, 3, cols * rows

type rgba struct {
	r, g, b, a byte
}
type guiState struct {
	zoom bool
	img  *sdl.Texture
	tree *pic
}
type pixelResult struct {
	pixels []byte
	index  int
}
type audioStat struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}
type pic struct {
	r Node
	g Node
	b Node
}

func (p *pic) String() string {
	return "( Picture \n " + p.r.String() + "\n " + p.g.String() + "\n " + p.b.String() + "\n)\n"
	//String() of Node is interface function different from this function
}
func NewPicture() *pic {
	p := &pic{}
	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()

	num := rand.Intn(20) + 1
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode())
	}
	num = rand.Intn(20) + 1
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode())
	}
	num = rand.Intn(20) + 1
	for i := 0; i < num; i++ {
		p.b.AddRandom(GetRandomNode())
	}
	for p.r.AddLeaf(GetRandomLeaf()) {
	}
	for p.g.AddLeaf(GetRandomLeaf()) {
	}
	for p.b.AddLeaf(GetRandomLeaf()) {
	}
	return p
}
func saveTree(p *pic) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}
	var biggestNo int
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".apt") {
			numStr := strings.TrimSuffix(name, ".apt")
			num, err := strconv.Atoi(numStr)
			if err == nil {
				if num > biggestNo {
					biggestNo = num
				}
			}
		}
	}
	saveName := strconv.Itoa(biggestNo+1) + ".apt"
	file, err := os.Create(saveName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintf(file, p.String())
}
func (p *pic) Mutate() {
	var nodeToMutate Node
	var write int
	r := rand.Intn(3)
	switch r {
	case 0:
		nodeToMutate = p.r
		write = 0
	case 1:
		nodeToMutate = p.g
		write = 1
	case 2:
		nodeToMutate = p.b
		write = 2
	}
	count := nodeToMutate.NodeCount()
	r = rand.Intn(count)
	nodeToMutate, count = GetNthNode(nodeToMutate, r, 0)
	mutation := Mutate(nodeToMutate)
	if write == 0 || nodeToMutate == p.r {
		p.r = mutation
	} else if write == 1 || nodeToMutate == p.g {
		p.g = mutation
	} else if write == 2 || nodeToMutate == p.b {
		p.b = mutation
	}
}
func sqrMag(a float32) float32 {
	return a * a
}
func imgToTex(renderer *sdl.Renderer, pixels *[]byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, *pixels, w*4)
	return tex
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
func (p *pic) pickRandColor() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return p.r
	case 1:
		return p.g
	case 2:
		return p.b
	default:
		panic("error calculating random")
	}

}
func mixPics(a *pic, b *pic) *pic {
	aCopy := &pic{CopyTree(a.r, nil), CopyTree(a.g, nil), CopyTree(a.b, nil)}
	aColor := aCopy.pickRandColor()
	bColor := b.pickRandColor()
	//
	aIndex := rand.Intn(aColor.NodeCount())
	aNode, _ := GetNthNode(aColor, aIndex, 0)

	bIndex := rand.Intn(bColor.NodeCount())
	bNode, _ := GetNthNode(bColor, bIndex, 0)
	bNodeCopy := CopyTree(bNode, bNode.GetParent())

	ReplaceNode(aNode, bNodeCopy)
	return aCopy
}
func evolve(selects []*pic) []*pic {
	newPics := make([]*pic, numPics)
	i := 0
	for i < len(selects) {
		a := selects[i]
		b := selects[rand.Intn(len(selects))]
		newPics[i] = mixPics(a, b)
		i++
	}
	for i < numPics {
		a := selects[rand.Intn(len(selects))]
		b := selects[rand.Intn(len(selects))]
		newPics[i] = mixPics(a, b)
		i++
	}
	for _, pics := range newPics {
		r := rand.Intn(4)
		for i = 0; i < r; i++ {
			pics.Mutate()
		}
	}
	return newPics
}
func aptToPixels(p *pic, w, h int) []byte {
	// -1.0 to 1.0
	scale := float32(255 / 2)
	offset := float32(-1.0 * scale)
	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	var r, g, b float32
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1

			r = p.r.Eval(x, y)
			g = p.g.Eval(x, y)
			b = p.b.Eval(x, y)

			pixels[pixelIndex] = byte(r*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(g*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(b*scale - offset)
			pixelIndex++
			pixelIndex++
		}
	}
	return pixels
}
func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Evolving Images", int32((1366-winWidht)/2), int32((768-winHeight)/2),
		int32(winWidht), int32(winHeight), sdl.WINDOW_SHOWN)
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
	//var deltaTim float32
	/*
		explosionBytes, audioSpec := sdl.LoadWAV("explode.wav")
		audioID, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
		if err != nil {
			fmt.Println(err)
		}
		defer sdl.FreeWAV(explosionBytes)

		audioState := audioStat{explosionBytes, audioID, audioSpec}
	*/
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")
	var mouse gui.MouseState
	mouse.ProcessMouse()
	rand.Seed(time.Now().UTC().Unix())

	picButtons := make([]*gui.ImageButton, numPics)
	picWidth := int(0.9 * float32(winWidht) / float32(cols))
	picHeight := int(0.88*float32(winHeight)/float32(rows) - (float32(winHeight)*0.04)/float32(rows))
	pixelChannel := make(chan pixelResult, numPics)
	picTree := make([]*pic, numPics)
	for i := range picTree {
		picTree[i] = NewPicture()
	}
	for i := range picTree {
		go func(i int) {
			pixels := aptToPixels(picTree[i], picWidth*2, picHeight*2)
			pixelChannel <- pixelResult{pixels, i}
		}(i)
	}
	evolButtonTex := gui.GetSinglePixelTex(renderer, sdl.Color{255, 255, 255, 0})
	evolButtonRect := sdl.Rect{int32(float32(winWidht)/2 - float32(picWidth)/2), int32(float32(winHeight)-float32(winHeight)*0.05) - 10, int32(picWidth), int32(float32(winHeight) * 0.05)}
	evolButton := gui.NewImageButton(renderer, evolButtonTex, &evolButtonRect, sdl.Color{255, 100, 0, 0})

	keyBoard := make([]gui.KeyStates, len(sdl.GetKeyboardState()))
	gui.ProcessKeys(&keyBoard)
	guiStatee := guiState{false, nil, nil}

	args := os.Args
	if len(args) > 1 {
		fileBytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
			//fmt.Println("provided file cannot be read")
		}
		fileStr := string(fileBytes)
		pictureNode := BeginLexing(fileStr)

		p := &pic{pictureNode.GetChildren()[0], pictureNode.GetChildren()[1], pictureNode.GetChildren()[2]}
		pixels := aptToPixels(p, winWidht*2, winHeight*2)
		guiStatee.img = imgToTex(renderer, &pixels, winWidht*2, winHeight*2)
		guiStatee.tree = p
		guiStatee.zoom = true
		fmt.Println(p.String())
	}
	for {
		//tim := time.Now()
		mouse.ProcessMouse()
		gui.ProcessKeys(&keyBoard)
		if keyBoard[sdl.SCANCODE_ESCAPE].IsDown && keyBoard[sdl.SCANCODE_ESCAPE].Changed {
			return
		}
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					mouse.X = int32(e.X * float32(winWidht))
					mouse.Y = int32(e.Y * float32(winHeight))
					if !mouse.Left {
						mouse.ChangedL = true
					} else {
						mouse.ChangedL = false
					}
					mouse.Left = true
				} else {
					mouse.Left = false
					mouse.ChangedL = false
				}
			}
		}
		//simulate
		select {
		case struc, ok := <-pixelChannel:
			if ok {
				tex := imgToTex(renderer, &struc.pixels, picWidth*2, picHeight*2)
				xi := struc.index % cols
				yi := (struc.index - xi) / cols
				x := int32(xi * picWidth)
				y := int32(yi * picHeight)
				xPad := int32(0.1 * float32(winWidht) / float32(cols+1))
				yPad := int32(0.1 * float32(winHeight) / float32(rows+1))
				x += xPad * (int32(xi) + 1)
				y += yPad * (int32(yi) + 1)
				rect := &sdl.Rect{x, y, int32(picWidth), int32(picHeight)}
				button := gui.NewImageButton(renderer, tex, rect, sdl.Color{200, 200, 255, 0})
				picButtons[struc.index] = button
			}
		default:
		}
		renderer.Clear()
		if !guiStatee.zoom {
			for i, button := range picButtons {
				if button != nil {
					button.Update(&mouse)
					if button.WasLeftClicked {
						button.IsSelected = !button.IsSelected
					} else if button.WasRightClicked {
						guiStatee.tree = picTree[i]
						zoomPix := aptToPixels(guiStatee.tree, winWidht*1.5, winHeight*1.5)
						guiStatee.img = imgToTex(renderer, &zoomPix, winWidht*1.5, winHeight*1.5)
						guiStatee.zoom = true
					}
					button.Draw(renderer)
				}
			}
			evolButton.Update(&mouse)
			if evolButton.WasLeftClicked {
				selectedPictures := make([]*pic, 0)
				for i, button := range picButtons {
					if button.IsSelected {
						selectedPictures = append(selectedPictures, picTree[i])
					}
				}
				if len(selectedPictures) != 0 {
					for i := range picButtons {
						picButtons[i] = nil
					}
					picTree = evolve(selectedPictures)
					for i := 0; i < numPics; i++ {
						go func(i int) {
							pixelss := aptToPixels(picTree[i], picWidth*2, picHeight*2)
							pixelChannel <- pixelResult{pixelss, i}
						}(i)
					}
				}
				evolButton.WasLeftClicked = false
			}
			evolButton.Draw(renderer)
		} else {
			var x, y int32
			xPad := int32(0.07 * float32(winWidht) / 2)
			yPad := int32(0.1 * float32(winHeight) / 2)
			x += xPad
			y += yPad
			rect := &sdl.Rect{x, y, int32(winWidht) - 2*xPad, int32(winHeight) - 2*yPad}
			if (keyBoard[sdl.SCANCODE_BACKSPACE].IsDown && keyBoard[sdl.SCANCODE_BACKSPACE].Changed) || (mouse.ChangedR && mouse.Right) {
				guiStatee.zoom = false
			}
			if keyBoard[sdl.SCANCODE_S].IsDown && keyBoard[sdl.SCANCODE_S].Changed {
				saveTree(guiStatee.tree)
			}
			renderer.Copy(guiStatee.img, nil, rect)
		}
		if keyBoard[sdl.SCANCODE_BACKSPACE].IsDown && keyBoard[sdl.SCANCODE_BACKSPACE].Changed {
			guiStatee.zoom = false
		}
		//render
		//
		renderer.Present()
		//deltaTim = float32(time.Since(tim).Seconds() * 1000)
		//fmt.Println(deltaTim)
	}
}
