package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
)

type Tile struct {
	walkable   bool
	People     []*Individual
	HitCount   int
	Dists      map[Destination]float64 // Used internally for dijkstra
	Directions map[Destination]Direction
	X          int
	Y          int
}

func (t *Tile) Walkable() bool {
	if t == nil {
		return false
	}
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
	TileWidth  int // TODO: do we need this?
	allPeople  []*Individual
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
		return nil
	}
	return &(w.tiles[x][y])
}

func (w *State) GetTileHighRes(x, y float64) *Tile {

	// TODO: Ed said we can just floor this?
	return w.GetTile(int(math.Floor(x)), int(math.Floor(y)))

}

var counter int = 0

func (w *State) AddRandom() {

	likelihood := Likelihood{
		ProbabilityFunctions: []func(int) bool{
			func(a int) bool {
				return true
			},
		},
		Probabilities: []float64{
			1,
		},
		Destination: Destination{
			X: 20,
			Y: 20,
		},
	}

	for {
		x := rand.Intn(w.GetWidth()-2) + 1
		y := rand.Intn(w.GetHeight()-2) + 1
		xf := rand.Float64()
		yf := rand.Float64()
		tile := w.GetTile(x, y)
		if tile.Walkable() && !w.IntersectsAnyone(float64(x)+xf, float64(y)+yf) {
			r, g, b := color.YCbCrToRGB(uint8(100), uint8(rand.Intn(256)), uint8(rand.Intn(256)))
			c := color.RGBA{r, g, b, 255}
			person := Individual{
				Loc:    geometry.NewPoint(float64(x)+xf, float64(y)+yf),
				Colour: c, UUID: "SimBot-" + strconv.Itoa(counter),
				Tick:     0,
				StepSize: 0.2,
				Likelihoods: []Likelihood{
					likelihood,
				},
			}
			tile.People = append(tile.People, &person)
			counter++
			w.allPeople = append(w.allPeople, &person)
			return
		}
	}
}

func (w *State) MoveIndividual(person *Individual, theta float64, distance float64) {
	cx, cy := person.Loc.GetXY()
	tile := w.GetTile(int(cx), int(cy))
	_, nx, ny := w.movementIntersects(cx, cy, theta, distance)

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

		// person is trying to leave! though
		log.Fatal("cannot leave bounry like this; ", cx, cy, nx, ny, theta/math.Pi, distance)

	}
}

func (w *State) MoveAll() {

	for _, p := range w.allPeople {
		theta := (rand.Float64() * 2 * math.Pi) * 0.9
		distance := math.Sqrt(rand.Float64()) / 3

		w.MoveIndividual(p, theta, distance)

	}

}