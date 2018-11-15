package actor

import (
	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"image/color"
)

type Destination struct {
	name string
}

type Likelyhood struct {
	Destination          Destination      // If this likelyhood is picked - where should we go
	probabilityFunctions []func(int) bool // Array of mutually exclusive functions to return "true" when we should use the corresponding probability
	probabilities        []float64        // Array of probabilities. MUST be same cardinality as the above.
}

type Actor struct {
	Loc         geometry.Point     // The point where this person current is stood
	Tick        int                // The current tick, from a base of 0, to measure time
	Likelyhoods []Likelyhood       // The array of likelyhoods for their preferences
	RegionIds   map[int32]struct{} // map (set) containing keys of all regions the actor is in
	UUID        string             // UUID of this actor for sending updates
	Colour      color.Color        // Colour to display person as
}

func (l *Likelyhood) ProbabilityAtTick(tick int) float64 {
	for i, useProb := range l.probabilityFunctions {
		if useProb(tick) {
			return l.probabilities[i]
		}
	}
	return 0
}

func (a *Actor) Next() Destination {
	a.Tick += 1
	return a.requestedDestination()
}

type ProbabilityPair struct {
	prob float64
	dest Destination
}

func (a *Actor) requestedDestination() Destination {
	// Get all of the likelyhood probabilities
	var total float64 = 0
	probs := make([]ProbabilityPair, 0)
	for _, likelyhood := range a.Likelyhoods {
		prob := likelyhood.ProbabilityAtTick(a.Tick)
		total += prob
		probs = append(probs, ProbabilityPair{
			prob: prob,
			dest: likelyhood.Destination,
		})
	}

	// Normalise them
	for i, prob := range probs {
		probs[i].prob = prob.prob / total
	}

	// Take a sample from them
	// TODO: Pick random between 0 and 1
	var randPick float64 = 0.5
	var totalSoFar float64 = 0
	for _, prob := range probs {
		totalSoFar += prob.prob
		if randPick < totalSoFar {
			// Return the destination
			return prob.dest
		}
	}

	return probs[0].dest
}
