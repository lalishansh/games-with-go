package game

type Monster struct {
	Pos
	Symbol    Tile
	Name      string
	hitpoints int
	strength  int
	speed     int32
}

func NewRat(x, y int32) *Monster {
	return &Monster{Pos{x, y}, 'R', "rat", 5, 5, 5}
}
func NewSpider(x, y int32) *Monster {
	return &Monster{Pos{x, y}, 'S', "spider", 10, 10, 2}
}
