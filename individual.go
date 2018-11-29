package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
)

type Likelihood struct {
	Destination          Destination            // If this likelihood is picked - where should we go
	ProbabilityFunctions []func(time.Time) bool // Array of mutually exclusive functions to return "true" when we should use the corresponding probability
	Probabilities        []float64              // Array of probabilities. MUST be same cardinality as the above.
}

type Individual struct {
	Loc          geometry.Point // The point where this person current is stood
	Tick         int            // The current tick, from a base of 0, to measure time
	Likelihoods  []Likelihood   // The array of likelihoods for their preferences
	StepSize     float64        // The average size step that the person will walk at (picked from a normal distribution)
	RegionIds    map[int32]bool // map (set) containing keys of all regions the actor is in
	UUID         string         // UUID of this actor for sending updates
	Colour       color.Color    // colour to render this individual
	LastMoveDist float64
	MoveDistAvg  float64     /// running avg of movements
	CurrentOri   float64     // Current orientation
	CurrentSway  float64     // Current sway
	target       Destination // current target destination
	leaveTime    time.Time   // time to leave current place
}

const (
	ORIENTATION_THRESHOLD = 0.1
)

func (l *Likelihood) ProbabilityAtTick(time time.Time) float64 {
	bestProb := 0.0
	for i, useProb := range l.ProbabilityFunctions {
		if useProb(time) && l.Probabilities[i] > bestProb {
			bestProb = l.Probabilities[i]
		}
	}
	return bestProb
}

func (i *Individual) Next(w *State) Destination {

	nilDest := Destination{}
	if i.target == nilDest {
		i.target = i.requestedDestination(w)
		return i.target
	}
	x, y := i.Loc.GetXY()
	dx := float64(i.target.X) - x
	dy := float64(i.target.Y) - y
	if dx*dx+dy*dy < i.target.R*i.target.R {
		// inside target
		if i.leaveTime.IsZero() {
			dest := w.scenario.GetDestination(i.target)
			if dest.MeanTime == 0 {
				event := dest.NextEventToEnd(w.time)
				if event == nil {
					i.leaveTime = w.time
				} else {
					i.leaveTime = event.End
				}
			} else {
				seconds := time.Duration((rand.NormFloat64()*dest.VarTime)+dest.MeanTime) * time.Second
				i.leaveTime = w.time.Add(seconds)
			}
		}
		if w.time.Before(i.leaveTime) {
			return i.target
		} else {
			i.leaveTime = time.Time{}
			i.target = i.requestedDestination(w)
			return i.target
		}
	} else {
		// outside target
		return i.target
	}

}

type ProbabilityPair struct {
	prob float64
	dest Destination
}

func (a *Individual) requestedDestination(w *State) Destination {
	// Get all of the likelihood probabilities
	var sum float64 = 0
	probs := make([]ProbabilityPair, 0)
	for _, likelihood := range a.Likelihoods {
		prob := likelihood.ProbabilityAtTick(w.time)
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

func (i *Individual) DirectionForDestination(dest Destination, w *State) utils.OptionalFloat64 {

	x, y := i.Loc.GetXY()
	dx := float64(i.target.X) - x
	dy := float64(i.target.Y) - y
	if dx*dx+dy*dy < i.target.R*i.target.R {
		// inside the area
		if rand.Float64() < 0.4 {
			return utils.OptionalFloat64WithValue((rand.Float64() - 0.5) * 2 * math.Pi)
		}
	}

	tile := w.GetTileHighRes(i.Loc.GetXY())
	if tile == nil {
		return utils.OptionalFloat64WithEmptyValue()
	}

	// Pick a random "sway" so they dont walk just in ordinal directions - more realistic

	// TODO: Look for people in their ordinal direction and follow them

	newOri, present := tile.Directions[dest].Value()
	newSway := (rand.Float64() - 0.5) * math.Pi / 2

	if i.LastMoveDist < 0.05 {
		newSway *= 2.5
	}

	if !present {
		// TODO: what to do here?
		return utils.OptionalFloat64WithEmptyValue()
	}

	// Only change direction if we are going sufficiently against the flow field
	if math.Abs(i.CurrentOri-newOri) >= ORIENTATION_THRESHOLD ||
		i.LastMoveDist < 0.05 {
		// Set the new orientation and sway
		i.CurrentOri = newOri
		i.CurrentSway = newSway
	}

	return utils.OptionalFloat64WithValue(i.CurrentOri + i.CurrentSway)

}
