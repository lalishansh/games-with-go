package game

import "github.com/t-RED-69/games-with-go/RPG/UI2d/sound"

type Monster struct {
	Charecter
	Debug  bool
	Debug2 bool
	activ  bool
}

func NewRat(x, y int32) *Monster {
	monstr := &Monster{}
	monstr.Pos = Pos{x, y}
	monstr.Symbol = 'R'
	monstr.Name = "Rat"
	monstr.Hitpoints = 5
	monstr.Strength = 5
	monstr.Speed = 5
	monstr.sightRange = 10
	monstr.activ = true
	monstr.Items = make([]*Item, 0)
	p := Pos{0, 0}

	monstr.Items = append(monstr.Items, NewSword(p), NewHelmet(p))
	if sound.SFX.EnemySnd[monstr.Name] == nil {
		sound.SFX.EnemySnd[monstr.Name] = sound.CharectorScrems(monstr.Name, 30)
	}
	return monstr
}
func NewSpider(x, y int32) *Monster {
	monstr := &Monster{}
	monstr.Pos = Pos{x, y}
	monstr.Symbol = 'S'
	monstr.Name = "Spider"
	monstr.Hitpoints = 10
	monstr.Strength = 10
	monstr.Speed = 2
	monstr.sightRange = 10
	monstr.activ = true
	if sound.SFX.EnemySnd[monstr.Name] == nil {
		sound.SFX.EnemySnd[monstr.Name] = sound.CharectorScrems(monstr.Name, 30)
	}
	return monstr
}
