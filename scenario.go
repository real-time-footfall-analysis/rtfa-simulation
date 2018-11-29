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
	Exit         destination   `json:"exit"`
	TotalPeople  int           `json:"totalPeople"`
	TotalGroups  int           `json:"totalGroups"`
	Destinations []destination `json:"destinations"`
}

type destination struct {
	X        int   `json:"X"`
	Y        int   `json:"Y"`
	RegionID int32 `json:"regionId,omitempty"`

	Name   string  `json:"name"`
	Events []event `json:"events"`
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

	scenario.Destinations = append(scenario.Destinations, scenario.Exit)

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

func (d *destination) GenerateRandomLikelihood() Likelihood {
	l := Likelihood{
		Destination: Destination{
			X: d.X,
			Y: d.Y,
		},
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
