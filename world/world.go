package world

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/individual"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"
)

type Tile struct {
	walkable bool
	People   []*individual.Individual
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
	allPeople  []*individual.Individual
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

var counter int = 0

func (w *State) AddRandom() {
	x := rand.Intn(w.GetWidth()-2) + 1
	y := rand.Intn(w.GetHeight()-2) + 1
	xf := rand.Float64()
	yf := rand.Float64()
	tile := w.GetTile(x, y)
	r, g, b := color.YCbCrToRGB(uint8(100), uint8(rand.Intn(256)), uint8(rand.Intn(256)))
	c := color.RGBA{r, g, b, 255}
	person := individual.Individual{Loc: geometry.NewPoint(float64(x)+xf, float64(y)+yf), Colour: c, UUID: "SimBot-" + strconv.Itoa(counter)}
	tile.People = append(tile.People, &person)
	counter++
	w.allPeople = append(w.allPeople, &person)
}

func (w *State) MoveIndividual(person *individual.Individual, theta float64, distance float64) {

	x, y := person.Loc.GetXY()
	tile := w.GetTile(int(x), int(y))

	cx, cy := person.Loc.GetXY()
	collide, nx, ny := w.movementintersects(cx, cy, theta, distance)
	if collide {
		fmt.Println("COLLIDED")
	}
	if nx >= 0 && int(nx) < w.GetWidth() &&
		ny >= 0 && int(ny) < w.GetHeight() {
		person.Loc.SetXY(nx, ny)
		if math.Floor(nx) == math.Floor(cx) &&
			math.Floor(ny) == math.Floor(cy) {

		} else {
			pos := -1
			for ti, p := range tile.People {
				if p.UUID == person.UUID {
					pos = ti
				}
			}
			if pos < 0 {
				log.Fatal("how?")
			}

			tile.People = append(tile.People[:pos], tile.People[pos+1:]...)
			newTile := w.GetTile(int(nx), int(ny))
			newTile.People = append(newTile.People, person)
		}
	} else {
		log.Fatal("cannot leave bounry like this")
	}
}

func (w *State) MoveAll() {

	for _, p := range w.allPeople {
		theta := (rand.Float64() * 2 * math.Pi) * 0.9
		distance := math.Sqrt(rand.Float64()) / 8

		w.MoveIndividual(p, theta, distance)

	}

}
