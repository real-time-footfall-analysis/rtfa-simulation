package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Scenario struct {
	MapImage     string        `json:"map"`
	RegionsFile  string        `json:"regions"`
	Lat          float64       `json:"lat,omitempty"`
	Lng          float64       `json:"lng,omitempty"`
	Start        time.Time     `json:"start,string"`
	End          time.Time     `json:"end,string"`
	Entrances    []Coord       `json:"entrance"`
	Exit         Destination   `json:"exit"`
	TotalPeople  int           `json:"totalPeople"`
	TotalGroups  int           `json:"totalGroups"`
	Destinations []Destination `json:"Destinations"`
	destMap      map[int]*Destination
}

type Destination struct {
	Coords   []Coord `json:"coords"`
	RegionID int32   `json:"regionId,omitempty"`
	ID       DestinationID

	Name     string  `json:"name"`
	Events   []event `json:"events"`
	MeanTime float64 `json:"meanUseTime"`
	VarTime  float64 `json:"useTimeVar"`
	Closed   bool
}

type Coord struct {
	X int     `json:"x"`
	Y int     `json:"y"`
	R float64 `json:"r"`
}

type event struct {
	Name       string    `json:"name"`
	Start      time.Time `json:"start,string"`
	End        time.Time `json:"end,string"`
	Popularity float64   `json:"popularity"`
}

func LoadScenario(path string) State {
	if !strings.HasSuffix(path, ".json") {
		log.Fatal("Scenario must be a .json file")
	}
	configFile, err := os.Open(path)
	if err != nil {
		log.Fatal("opening region file", err.Error())
	}

	var scenario Scenario

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&scenario); err != nil {
		log.Fatal("parsing config file", err.Error())
	}
	log.Println("map: ", scenario.MapImage)
	s := LoadFromImage(scenario.MapImage)
	s.ScenarioName = strings.TrimSuffix(path, ".json")
	s.LoadRegions(scenario.RegionsFile, scenario.Lat, scenario.Lng)

	scenario.destMap = make(map[int]*Destination)
	idCount := 1
	for i, d := range scenario.Destinations {
		if d.RegionID > 0 {
			region := s.FindRegion(d.RegionID)
			coord := Coord{
				X: int(region.X),
				Y: int(region.Y),
				R: float64(region.Radius),
			}
			scenario.Destinations[i].Coords = append(scenario.Destinations[i].Coords, coord)
		}
		scenario.Destinations[i].ID = DestinationID{idCount}
		idCount++
	}

	scenario.Exit.ID = DestinationID{idCount}
	scenario.Destinations = append(scenario.Destinations, scenario.Exit)

	for i, d := range scenario.Destinations {
		scenario.destMap[d.ID.ID] = &scenario.Destinations[i]
	}

	log.Println(scenario)
	log.Println(scenario.Destinations[0].Events[0].Start)
	log.Println(scenario.Destinations[0].Events[0].End)

	s.scenario = scenario
	s.time = scenario.Start
	return s
}

func (s *Scenario) GenerateRandomPersonality() []Likelihood {
	var ls []Likelihood

	for _, d := range s.Destinations {
		ls = append(ls, d.GenerateRandomLikelihood())
	}
	return ls
}

func (s *Scenario) GetDestination(target DestinationID) *Destination {
	return s.destMap[target.ID]
}

func (d *Destination) GenerateRandomLikelihood() Likelihood {
	l := Likelihood{
		Destination: d.ID,
	}

	for _, e := range d.Events {

		start := e.Start
		end := e.End
		pfunc :=
			func(t time.Time) bool {
				if t.After(start) && t.Before(end) {
					return true
				}
				return false
			}
		prob := e.Popularity * rand.Float64() * 20

		l.ProbabilityFunctions = append(l.ProbabilityFunctions, pfunc)
		l.Probabilities = append(l.Probabilities, prob)
	}

	return l

}

func (d *Destination) NextEventToEnd(t time.Time) *event {
	earliest := time.Unix(1<<63-62135596801, 999999999)
	ret := -1
	for i, e := range d.Events {
		if e.Start.Before(t) && e.End.After(t) {
			if e.End.Before(earliest) {
				earliest = e.End
				ret = i
			}
		}
	}
	if ret == -1 {
		return nil
	}
	return &d.Events[ret]
}

func (d *Destination) Contains(x, y int) bool {
	for _, c := range d.Coords {
		dxsq := (c.X - x) * (c.X - x)
		dysq := (c.Y - y) * (c.Y - y)
		if int(c.R*c.R) > dxsq+dysq {
			return true
		}
	}
	return false
}

func (d *Destination) ContainsCenter(x, y int) bool {
	for _, c := range d.Coords {
		if c.X == x && c.Y == y {
			return true
		}
	}
	return false
}

func (d *Destination) Close() bool {
	closed := d.Closed
	d.Closed = true
	return !closed
}

func (d *Destination) Open() bool {
	closed := d.Closed
	d.Closed = false
	return closed
}

func (d *Destination) isClosed() bool {
	return d.Closed
}
