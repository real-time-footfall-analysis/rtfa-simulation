package individual

import (
	"math/rand"

	"github.com/real-time-footfall-analysis/rtfa-simulation/directions"
)

type Point struct {
	X float64
	Y float64
}

type Likelihood struct {
	Destination          directions.Destination // If this likelihood is picked - where should we go
	probabilityFunctions []func(int) bool       // Array of mutually exclusive functions to return "true" when we should use the corresponding probability
	probabilities        []float64              // Array of probabilities. MUST be same cardinality as the above.
}

type Individual struct {
	Loc         Point        // The point where this person current is stood
	Tick        int          // The current tick, from a base of 0, to measure time
	Likelihoods []Likelihood // The array of likelihoods for their preferences
	StepSize    float64      // The average size step that the person will walk at (picked from a normal distribution)
}

func (l *Likelihood) ProbabilityAtTick(tick int) float64 {
	for i, useProb := range l.probabilityFunctions {
		if useProb(tick) {
			return l.probabilities[i]
		}
	}
	return 0
}

func (i *Individual) Next() directions.Destination {
	i.Tick += 1
	return i.requestedDestination()
}

type ProbabilityPair struct {
	prob float64
	dest directions.Destination
}

func (a *Individual) requestedDestination() directions.Destination {
	// Get all of the likelihood probabilities
	var sum float64 = 0
	probs := make([]ProbabilityPair, 0)
	for _, likelihood := range a.Likelihoods {
		prob := likelihood.ProbabilityAtTick(a.Tick)
		sum += prob
		probs = append(probs, ProbabilityPair{
			prob: prob,
			dest: likelihood.Destination,
		})
	}

	// Normalise them
	for i, prob := range probs {
		probs[i].prob = prob.prob / sum
	}

	// Take a sample from them
	randPick := rand.Float64()
	var sumSoFar float64 = 0
	for _, prob := range probs {
		sumSoFar += prob.prob
		if randPick < sumSoFar {
			// Return the destination
			return prob.dest
		}
	}

	return probs[0].dest
}

func (i *Individual) DirectionForDestination(dest directions.Destination, macroMap *directions.MacroMap) directions.Direction {
	tile, err := macroMap.GetTileHighRes(i.Loc.X, i.Loc.Y)
	if err != nil {
		return directions.DirectionUnknown
	}
	return tile.Directions[dest]
}
