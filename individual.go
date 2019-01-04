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
	Destination          DestinationID          // If this likelihood is picked - where should we go
	ProbabilityFunctions []func(time.Time) bool // Array of mutually exclusive functions to return "true" when we should use the corresponding probability
	Probabilities        []float64              // Array of probabilities. MUST be same cardinality as the above.
}

type Individual struct {
	Loc          geometry.Point // The point where this person current is stood
	Tick         int            // The current tick, from a base of 0, to measure time
	Likelihoods  []Likelihood   // The array of likelihoods for their preferences
	StepSize     float64        // The average size step that the person will walk at (picked from a normal distribution)
	UpdateSender bool           // set if this AI is to send updates
	RegionIds    map[int32]bool // map (set) containing keys of all regions the actor is in
	UUID         string         // UUID of this actor for sending updates
	Colour       color.Color    // colour to render this individual
	LastMoveDist float64
	MoveDistAvg  float64      /// running avg of movements
	CurrentOri   float64      // Current orientation
	CurrentSway  float64      // Current sway
	target       *Destination // current target Destination
	leaveTime    time.Time    // time to leave current place
	UpdateChan   *UpdateChan
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

func (i *Individual) Next(w *State) DestinationID {

	if i.target == nil || i.target.isClosed() {
		i.leaveTime = time.Time{}
		destID := i.requestedDestination(w)
		i.target = w.scenario.GetDestination(destID)
		return destID
	}
	dest := w.scenario.GetDestination(i.target.ID)
	x, y := i.Loc.GetXY()
	if dest.Contains(int(x), int(y)) {
		// inside target
		if i.leaveTime.IsZero() {
			dest := w.scenario.GetDestination(i.target.ID)
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
			return i.target.ID
		} else {
			i.leaveTime = time.Time{}
			destID := i.requestedDestination(w)
			i.target = w.scenario.GetDestination(destID)
			return destID
		}
	} else {
		// outside target
		return i.target.ID
	}

}

type ProbabilityPair struct {
	prob float64
	dest DestinationID
}

func (a *Individual) requestedDestination(w *State) DestinationID {
	// Get all of the likelihood probabilities
	var sum float64 = 0
	probs := make([]ProbabilityPair, 0)
	for _, likelihood := range a.Likelihoods {
		dest := w.scenario.GetDestination(likelihood.Destination)
		if !dest.isClosed() {
			prob := likelihood.ProbabilityAtTick(w.time)
			sum += prob
			probs = append(probs, ProbabilityPair{
				prob: prob,
				dest: likelihood.Destination,
			})
		}
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
			// Return the Destination
			return prob.dest
		}
	}

	return probs[0].dest
}

func (i *Individual) DirectionForDestination(dest DestinationID, w *State) utils.OptionalFloat64 {

	x, y := i.Loc.GetXY()
	if i.target.Contains(int(x), int(y)) {
		// inside the area
		if rand.Float64() < 0.4 {
			return utils.OptionalFloat64WithValue((rand.Float64() - 0.5) * 2 * math.Pi)
		}
	}

	tile := w.GetTileHighRes(i.Loc.GetXY())
	if tile == nil {
		return utils.OptionalFloat64WithEmptyValue()
	}

	var newOri float64

	// Pick a random "sway" so they dont walk just in ordinal directions - more realistic

	// TODO: Look for people in their ordinal direction and follow them
	v, ok := tile.Directions[dest]
	if !ok {
		newOri = i.dumbDirection(w, dest)
	} else {
		Ori, present := v.Value()

		if !present {
			newOri = i.dumbDirection(w, dest)
		} else {
			newOri = Ori
		}
	}

	newSway := (rand.Float64() - 0.5) * math.Pi / 2

	if i.LastMoveDist < 0.05 {
		newSway *= 2.5
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

func (i *Individual) dumbDirection(w *State, dest DestinationID) float64 {
	x, y := i.Loc.GetXY()
	destination := w.scenario.GetDestination(dest)
	r := rand.Intn(len(destination.Coords))
	coord := destination.Coords[r]
	dx := float64(coord.X) - x
	dy := float64(coord.Y) - y
	theta := math.Atan2(dy, dx)
	return theta
}
