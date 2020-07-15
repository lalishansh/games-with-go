package game

type Item struct {
	Entity
}

func NewSword(p Pos) *Item {
	return &Item{Entity: Entity{'s', p, "Sword"}}
}
func NewHelmet(p Pos) *Item {
	return &Item{Entity: Entity{'h', p, "Helmet"}}
}
