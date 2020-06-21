package gui

import "github.com/veandco/go-sdl2/sdl"

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

type ImageButton struct {
	Img             *sdl.Texture
	Rect            *sdl.Rect
	WasLeftClicked  bool
	WasRightClicked bool
	IsSelected      bool
	SelectionTex    *sdl.Texture
}

func NewImageButton(renderer *sdl.Renderer, image *sdl.Texture, rect *sdl.Rect, selectColor sdl.Color) *ImageButton {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = selectColor.R
	pixels[1] = selectColor.G
	pixels[2] = selectColor.B
	pixels[3] = selectColor.A
	tex.Update(nil, pixels, 4)
	return &ImageButton{image, rect, false, false, false, tex}
}
func (button *ImageButton) Update(mouse *MouseState) {
	if button.Rect.HasIntersection(&sdl.Rect{mouse.X, mouse.Y, 1, 1}) {
		button.WasLeftClicked = mouse.ChangedL && mouse.Left
		button.WasRightClicked = mouse.ChangedR && mouse.Right
	}
}
func (button *ImageButton) Draw(renderer *sdl.Renderer) {
	if button.IsSelected {
		borderRect := *button.Rect
		borderThickness := int32(float32(borderRect.W) * 0.02)
		borderRect.W += borderThickness * 2
		borderRect.H += borderThickness * 2
		borderRect.X -= borderThickness
		borderRect.Y -= borderThickness
		renderer.Copy(button.SelectionTex, nil, &borderRect)
	}
	renderer.Copy(button.Img, nil, button.Rect)
}
func GetSinglePixelTex(renderer *sdl.Renderer, color sdl.Color) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = color.R
	pixels[1] = color.G
	pixels[2] = color.B
	pixels[3] = color.A
	tex.Update(nil, pixels, 4)
	return tex
}
