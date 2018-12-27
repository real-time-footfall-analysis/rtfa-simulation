package main

import (
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
)

type Group struct {
	Individuals []*Individual
}

func (g *Group) Next(channel chan map[*Individual]utils.OptionalFloat64, w *State) {

	dests := make(map[DestinationID]int, 0)
	bestVal := 0

	// Get each Destination that someone would like to go to, and tally them up
	for _, individual := range g.Individuals {
		dest := individual.Next(w)

		val, ok := dests[dest]
		valSet := 0
		if !ok {
			valSet = 1
			dests[dest] = valSet

		} else {
			valSet = val + 1
			dests[dest] = valSet
		}

		if valSet > bestVal {
			bestVal = valSet
		}
	}

	var chosenDest DestinationID
	// Return the Destination with the highest weight
	for dest, val := range dests {
		if val == bestVal {
			chosenDest = dest
		}
	}

	// Tell each person in the group where they need to go, and they will tell you which direction they need to go in

	directions := make(map[*Individual]utils.OptionalFloat64, 0)
	for _, individual := range g.Individuals {
		directions[individual] = individual.DirectionForDestination(chosenDest, w)
	}
	channel <- directions
}
