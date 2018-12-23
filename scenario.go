package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

type Scenario struct {
	Start        time.Time     `json:"start,string"`
	End          time.Time     `json:"end,string"`
	EntranceX    int           `json:"entranceX"`
	EntranceY    int           `json:"entranceY"`
	Exit         Destination   `json:"exit"`
	TotalPeople  int           `json:"totalPeople"`
	TotalGroups  int           `json:"totalGroups"`
	Destinations []Destination `json:"Destinations"`
	destMap      map[int]*Destination
}

type Destination struct {
	X        int   `json:"X"`
	Y        int   `json:"Y"`
	RegionID int32 `json:"regionId,omitempty"`
	ID       DestinationID

	Name     string  `json:"name"`
	Events   []event `json:"events"`
	Radius   float64 `json:"radius"`
	MeanTime float64 `json:"meanUseTime"`
	VarTime  float64 `json:"useTimeVar"`
}

type event struct {
	Name       string    `json:"name"`
	Start      time.Time `json:"start,string"`
	End        time.Time `json:"end,string"`
	Popularity float64   `json:"popularity"`
}

func (s *State) LoadScenario(path string) {
	configFile, err := os.Open(path)
	if err != nil {
		log.Fatal("opening region file", err.Error())
	}

	var scenario Scenario

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&scenario); err != nil {
		log.Fatal("parsing config file", err.Error())
	}

	scenario.destMap = make(map[int]*Destination)
	idCount := 1
	for i, d := range scenario.Destinations {
		if d.RegionID > 0 {
			scenario.Destinations[i].X = int(s.FindRegion(d.RegionID).X)
			scenario.Destinations[i].Y = int(s.FindRegion(d.RegionID).Y)
			scenario.Destinations[i].Radius = float64(int(s.FindRegion(d.RegionID).Radius))
		}
		log.Println(scenario.Destinations[i].Name, " - ", scenario.Destinations[i].X, ",", scenario.Destinations[i].Y)
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
