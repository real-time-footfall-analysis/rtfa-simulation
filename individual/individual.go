package individual

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"

	"github.com/real-time-footfall-analysis/rtfa-simulation/directions"
)

type Likelihood struct {
	Destination          directions.Destination // If this likelihood is picked - where should we go
	ProbabilityFunctions []func(int) bool       // Array of mutually exclusive functions to return "true" when we should use the corresponding probability
	Probabilities        []float64              // Array of probabilities. MUST be same cardinality as the above.
}

type Individual struct {
	Loc         geometry.Point     // The point where this person current is stood
	Tick        int                // The current tick, from a base of 0, to measure time
	Likelihoods []Likelihood       // The array of likelihoods for their preferences
	StepSize    float64            // The average size step that the person will walk at (picked from a normal distribution)
	RegionIds   map[int32]struct{} // map (set) containing keys of all regions the actor is in
	UUID        string             // UUID of this actor for sending updates
	Colour      color.Color        // colour to render this individual
}

func (l *Likelihood) ProbabilityAtTick(tick int) float64 {
	for i, useProb := range l.ProbabilityFunctions {
		if useProb(tick) {
			return l.Probabilities[i]
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

func (i *Individual) DirectionForDestination(dest directions.Destination, macroMap *directions.MacroMap) utils.OptionalFloat64 {
	tile, err := macroMap.GetTileHighRes(i.Loc.GetXY())
	if err != nil {
		return utils.OptionalFloat64WithEmptyValue()
	}

	// Pick a random "sway" so they dont walk just in ordinal directions - more realistic

	// TODO: Look for people in their ordinal direction and follow them

	sway := (rand.Float64() * math.Pi / 4) - math.Pi/2
	theta := 0.0

	switch tile.Directions[dest] {
	case directions.DirectionN:
		theta = -math.Pi / 2
	case directions.DirectionNE:
		theta = -math.Pi / 4
	case directions.DirectionE:
		theta = 0
	case directions.DirectionSE:
		theta = math.Pi / 4
	case directions.DirectionS:
		theta = math.Pi / 2
	case directions.DirectionSW:
		theta = 3 * math.Pi / 4
	case directions.DirectionW:
		theta = math.Pi
	case directions.DirectionNW:
		theta = -3 * math.Pi / 4
	default:
		return utils.OptionalFloat64WithEmptyValue()
	}
	return utils.OptionalFloat64WithValue(theta + sway)
}
