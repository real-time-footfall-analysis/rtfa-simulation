package main

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"
)

type Tile struct {
	walkable   bool
	People     []*Individual
	HitCount   int
	Dists      map[DestinationID]float64 // Used internally for dijkstra
	Directions map[DestinationID]utils.OptionalFloat64
	X          int
	Y          int

	blockedNorth bool
	blockedEast  bool
	blockedSouth bool
	blockedWest  bool

	destID uint32
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
	tiles              [][]Tile
	width              int
	height             int
	background         image.Image
	Regions            []Region
	allPeople          []*Individual
	time               time.Time
	BulkSend           bool
	maxSenders         int
	currentSenders     int
	SendUpdates        bool
	peopletoAdd        int
	scenario           Scenario
	peopleAdded        int
	peopleCurrent      int
	groups             []*Group
	startWaiter        chan bool
	peopleAddedChan    chan int
	peopleCurrentChan  chan int
	simulationTimeChan chan time.Time
	currentSendersChan chan int
	totalSendsChan     chan int
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

func (w *State) AddRandom() *Individual {

	for i := 0; i < 100; i++ {
		randi := rand.Int() % len(w.scenario.Entrances)
		x := w.scenario.Entrances[randi].X
		y := w.scenario.Entrances[randi].Y
		theta := rand.Float64() * 2 * math.Pi
		radius := rand.Float64() * w.scenario.Entrances[randi].R
		xf := radius * math.Cos(theta)
		yf := radius * math.Sin(theta)
		tile := w.GetTile(int(float64(x)+xf), int(float64(y)+yf))
		if tile.Walkable() && !w.IntersectsAnyone(float64(x)+xf, float64(y)+yf) {
			r, g, b := color.YCbCrToRGB(uint8(100), uint8(rand.Intn(256)), uint8(rand.Intn(256)))
			c := color.RGBA{r, g, b, 255}

			name := fmt.Sprintf("SimBot-%029d", counter)
			//randSetIndex := rand.Intn(4)
			sendUpdates := w.maxSenders > w.currentSenders
			sendUpdates = sendUpdates && rand.Intn(w.scenario.TotalPeople-w.peopleAdded) <= (w.maxSenders-w.currentSenders)
			person := Individual{
				Loc:    geometry.NewPoint(float64(x)+xf, float64(y)+yf),
				Colour: c, UUID: name,
				Tick:        0,
				StepSize:    0.2,
				Likelihoods: w.scenario.GenerateRandomPersonality(),
				RegionIds:   make(map[int32]bool),
				sendUpdates: sendUpdates,
			}
			tile.People = append(tile.People, &person)
			counter++
			w.peopleAdded++
			w.peopleCurrent++
			if person.sendUpdates {
				w.currentSenders++
			}
			w.allPeople = append(w.allPeople, &person)
			return &person
		}
	}
	return nil
}

func (w *State) MoveIndividual(person *Individual, theta float64, distance float64) {
	cx, cy := person.Loc.GetXY()
	tile := w.GetTile(int(cx), int(cy))
	_, nx, ny := w.movementIntersects(cx, cy, theta, distance)

	if nx >= 0 && int(nx) < w.GetWidth() &&
		ny >= 0 && int(ny) < w.GetHeight() {
		distThing := (nx-cx)*(nx-cx) + (ny-cy)*(ny-cy)
		dist := math.Sqrt(distThing)
		person.LastMoveDist = dist
		person.MoveDistAvg = (person.MoveDistAvg * 0.8) + (dist * 0.2)
		person.Loc.SetXY(nx, ny)
		if math.Floor(nx) == math.Floor(cx) &&
			math.Floor(ny) == math.Floor(cy) {

		} else {
			pos := -1
			for ti, p := range tile.People {
				if p.UUID == person.UUID {
					pos = ti
					break
				}
			}
			if pos < 0 {
				//return
				log.Fatal("how?")
			}

			tile.People = append(tile.People[:pos], tile.People[pos+1:]...)
			if !w.scenario.Exit.Contains(int(nx), int(ny)) {
				newTile := w.GetTile(int(nx), int(ny))
				newTile.People = append(newTile.People, person)
			} else {
				w.peopleCurrent--
				w.currentSenders--
				allPos := -1
				for ti, p := range w.allPeople {
					if p.UUID == person.UUID {
						allPos = ti
						break
					}
				}
				if allPos < 0 {
					log.Fatal("how?2")
				}
				gPos := -1
				iPos := -1
				for gi, g := range w.groups {
					for ti, p := range g.Individuals {
						if p.UUID == person.UUID {
							gPos = gi
							iPos = ti
							break
						}
					}
				}
				if gPos < 0 {
					log.Fatal("how?3")
				}
				if iPos < 0 {
					log.Fatal("how?4")
				}
				w.groups[gPos].Individuals = append(w.groups[gPos].Individuals[:iPos], w.groups[gPos].Individuals[iPos+1:]...)
				w.allPeople = append(w.allPeople[:allPos], w.allPeople[allPos+1:]...)
				LeaveAllRegions(w, person, w.time, w.BulkSend, w.SendUpdates)
			}
		}
	} else {
		w.movementIntersects(cx, cy, theta, distance)
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

func (w *State) TickTime() {
	w.time = w.time.Add(time.Second)

}

func (s *State) FindRegion(id int32) *Region {
	for i, r := range s.Regions {
		if r.ID == id {
			return &s.Regions[i]
		}
	}
	return nil
}

func (s *State) MakeChannes() {
	s.startWaiter = make(chan bool)
	s.peopleAddedChan = make(chan int)
	s.peopleCurrentChan = make(chan int)
	s.simulationTimeChan = make(chan time.Time)
	s.currentSendersChan = make(chan int)
	s.totalSendsChan = make(chan int)
}
