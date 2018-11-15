package group

import (
	"github.com/real-time-footfall-analysis/rtfa-simulation/directions"
	"github.com/real-time-footfall-analysis/rtfa-simulation/individual"
)

type Group struct {
	individuals []*individual.Individual
	macroMap    *directions.MacroMap
}

func (g *Group) Next(channel chan map[*individual.Individual]directions.Direction) {

	dests := make(map[directions.Destination]int, 0)
	bestVal := 0

	// Get each destination that someone would like to go to, and tally them up
	for _, individual := range g.individuals {
		dest := individual.Next()

		val, ok := dests[dest]
		valSet := 0
		if !ok {
			valSet = 1
			dests[dest] = valSet
		}
		valSet = val + 1
		dests[dest] = valSet

		if valSet > bestVal {
			bestVal = valSet
		}
	}

	var chosenDest directions.Destination
	// Return the destination with the highest weight
	for dest, val := range dests {
		if val == bestVal {
			chosenDest = dest
		}
	}

	// Tell each person in the group where they need to go, and they will tell you which direction they need to go in

	directions := make(map[*individual.Individual]directions.Direction, 0)
	for _, individual := range g.individuals {
		directions[individual] = individual.DirectionForDestination(chosenDest, g.macroMap)
	}
	channel <- directions
}
