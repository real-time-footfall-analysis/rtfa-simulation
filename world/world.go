package world

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/individual"
	"image"
	"image/color"
	"math"
	"math/rand"
)

type Tile struct {
	walkable bool
	People   []individual.Individual
	HitCount int
}

func (t *Tile) Walkable() bool {
	return t.walkable
}
func (t *Tile) SetWalkable(b bool) {
	t.walkable = b
}

type State struct {
	tiles      [][]Tile
	width      int
	height     int
	background image.Image
	Regions    []Region
}

func (w *State) GetWidth() int {
	return w.width
}

func (w *State) GetHeight() int {
	return w.height
}

func (w *State) GetImage() image.Image {
	return w.background
}

func (w *State) GetTile(x, y int) *Tile {
	if x < 0 || x >= w.width || y < 0 || y >= w.height {
		fmt.Println("out of range: %i %i", x, y)
	}
	return &(w.tiles[x][y])
}

func (w *State) AddRandom() {
	x := rand.Intn(w.GetWidth()-2) + 1
	y := rand.Intn(w.GetHeight()-2) + 1
	xf := rand.Float64()
	yf := rand.Float64()
	tile := w.GetTile(x, y)
	r, g, b := color.YCbCrToRGB(uint8(100), uint8(rand.Intn(256)), uint8(rand.Intn(256)))
	c := color.RGBA{r, g, b, 255}
	tile.People = append(tile.People, individual.Individual{Loc: geometry.NewPoint(float64(x)+xf, float64(y)+yf), Colour: c})
}

func (w *State) MoveRandom() {
	for {
		x := rand.Intn(w.GetWidth())
		y := rand.Intn(w.GetHeight())
		tile := w.GetTile(x, y)
		tile.HitCount++

		if len(tile.People) > 0 {
			i := rand.Intn(len(tile.People))
			p := tile.People[i]
			theta := (rand.Float64() * 2 * math.Pi)
			distance := math.Sqrt(rand.Float64())
			cx, cy := p.Loc.GetLatestXY()
			collide, nx, ny := w.movementintersects(cx, cy, theta, distance)
			if collide {
				fmt.Println("COLLIDED")
			}
			if nx >= 0 && int(nx) < w.GetWidth() &&
				ny >= 0 && int(ny) < w.GetHeight() {
				tile.People[i].Loc.SetXY(nx, ny)
				if math.Floor(nx) == math.Floor(cx) &&
					math.Floor(ny) == math.Floor(cy) {

				} else {
					tile.People = append(tile.People[:i], tile.People[i+1:]...)
					newTile := w.GetTile(int(nx), int(ny))
					newTile.People = append(newTile.People, p)
				}
				return
			} else {
				tile.People = append(tile.People[:i], tile.People[i+1:]...)
				return
			}
		}
	}
}
