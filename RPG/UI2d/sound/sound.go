package sound

import (
	"math/rand"
	"strconv"

	"github.com/veandco/go-sdl2/mix"
)

var randSRC *rand.Rand
var SFX sounds

//keep these in mind
const (
	DoorOpnINT int = iota
	FootstpsINT
	EnmyHitINT
	PlyrHitINT
)

type sounds struct {
	DoorOpen  *mix.Chunk
	Footsteps []*mix.Chunk
	EnemySnd  map[string]*Screams
	PlyrHit   []*mix.Chunk
}
type Screams struct {
	Attack *mix.Chunk
	Hit    *mix.Chunk
	Death  *mix.Chunk
}

func init() {
	err := mix.Init(mix.INIT_OGG)
	if err != nil {
		panic(err)
	}
}
func NewSI() {
	err := mix.OpenAudio(22050, mix.DEFAULT_FORMAT, 2, 4096)
	if err != nil {
		panic(err)
	}
	mus, err := mix.LoadMUS("UI2d/sound/ambient.ogg")
	if err != nil {
		panic(err)
	}
	mus.Play(-1)
	randSRC = rand.New(rand.NewSource(1))
	loadSounds()
}
func loadSounds() {
	SFX.DoorOpen = NewSound("DoorOpen", 30)

	SFX.Footsteps = make([]*mix.Chunk, 6)
	for i := 0; i < 6; i++ {
		SFX.Footsteps[i] = NewSound("footstep"+strconv.Itoa(i+1), 10)
	}
	SFX.PlyrHit = make([]*mix.Chunk, 3)
	for i := 0; i < 3; i++ {
		SFX.PlyrHit[i] = NewSound("swing"+strconv.Itoa(i+1), 40)
	}
	SFX.EnemySnd = make(map[string]*Screams)
}
func Play(Sound int, str string, typ rune) {
	var snd *mix.Chunk
	switch Sound {
	case DoorOpnINT:
		snd = SFX.DoorOpen
	case FootstpsINT:
		SFX.Footsteps[randSRC.Intn(6)].Play(Sound, 0)
		return
	case EnmyHitINT:
		switch typ {
		case 'a', 'A':
			snd = SFX.EnemySnd[str].Attack
		case 'h', 'H':
			snd = SFX.EnemySnd[str].Hit
		case 'd', 'D':
			snd = SFX.EnemySnd[str].Death
		}
	case PlyrHitINT:
		SFX.PlyrHit[randSRC.Intn(3)].Play(Sound, 0)
		return
	}
	snd.Play(Sound, 0)
}
func HaltSounds(channel int) {
	mix.HaltChannel(channel)
}
func NewSound(name string, volume int) *mix.Chunk {
	snd, err := mix.LoadWAV("UI2d/sound/" + name + ".ogg")
	if err != nil {
		panic(err)
	}
	snd.Volume(volume)
	return snd
}
func CharectorScrems(name string, volume int) *Screams {
	snd := Screams{}
	snd.Attack = NewSound(name+"Attack", volume)
	snd.Hit = NewSound(name+"Hit", volume)
	snd.Death = NewSound(name+"Death", volume)
	return &snd
}
