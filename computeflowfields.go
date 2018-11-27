package main

import (
	"fmt"
	_ "image/png"
	"math"
)

type pairInts struct {
	fst, snd int
}
type Direction int

const (
	DirectionN Direction = iota
	DirectionNE
	DirectionE
	DirectionSE
	DirectionS
	DirectionSW
	DirectionW
	DirectionNW
	DirectionUnknown
	NoMovementNeeded
	Solid
)

type Destination struct { // Indicies into the macromap
	X int
	Y int
}

func InitFlowFields() {

	initDeltas()

}

func (d Direction) String() string {

	if d == DirectionN {
		return "↑"
	}
	if d == DirectionNE {
		return "↗"
	}
	if d == DirectionE {
		return "→"
	}
	if d == DirectionSE {
		return "↘"
	}
	if d == DirectionS {
		return "↓"
	}
	if d == DirectionSW {
		return "↙"
	}
	if d == DirectionW {
		return "←"
	}
	if d == DirectionNW {
		return "↖"
	}
	if d == DirectionUnknown {
		return "?"
	}
	if d == NoMovementNeeded {
		return "X"
	}
	if d == Solid {
		return "#"
	}

	return "!"
}

func (w *State) PrintDirections(destination Destination) {

	for i := 0; i < w.GetHeight(); i++ {
		for j := 0; j < w.GetWidth(); j++ {

			tile := w.GetTile(j, i)

			if tile.Dists[destination] == 0 {
				fmt.Printf("X")
			} else if !tile.Walkable() {
				fmt.Print("#")
			} else {
				fmt.Print(tile.Directions[destination])
			}

		}
		fmt.Println()

	}
	fmt.Println()

}

func (w *State) PrintDistances(destination Destination) {

	for i := 0; i < w.GetHeight(); i++ {
		for j := 0; j < w.GetWidth(); j++ {

			tile := w.GetTile(j, i)


			fmt.Printf("%05.f", tile.Dists[destination])

		}
		fmt.Println()

	}
	fmt.Println()

}

// Generates the flow field to the destination
func (w *State) GenerateFlowField(destination Destination) error {

	FindShortestPath(w, destination)
	w.computeDirections(destination)

	return nil

}

// Assuming that distance information has been filled in by dijkstra,
// calculate the directions needed
func (w *State) computeDirections(dest Destination) error {

	for i := 0; i < w.GetWidth(); i++ {
		for j := 0; j < w.GetHeight(); j++ {

			// Disallowed delta movements
			disallowed := make([]pairInts, 0)

			tile := w.GetTile(i, j)

			// Check for wall in border
			hasWallInBorder := false
			for dX := -1; dX <= 1; dX++ {
				for dY := -1; dY <= 1; dY++ {

					// Skip looking at myself
					if dX == 0 && dY == 0 {
						continue
					}

					newX := i + dX
					newY := j + dY

					if newX >= 0 && newX < w.GetWidth() &&
						newY >= 0 && newY < w.GetHeight() {

						newTile := w.GetTile(newX, newY)
						if !newTile.Walkable() {
							hasWallInBorder = true

							if dX == -1 && dY == 0 {
								disallowed = append(disallowed, pairInts{-1, 1})
								disallowed = append(disallowed, pairInts{-1, -1})
							}
							if dX == 0 && dY == 1 {
								disallowed = append(disallowed, pairInts{-1, 1})
								disallowed = append(disallowed, pairInts{1, 1})
							}
							if dX == 0 && dY == -1 {
								disallowed = append(disallowed, pairInts{-1, -1})
								disallowed = append(disallowed, pairInts{1, -1})
							}
							if dX == 1 && dY == 0 {
								disallowed = append(disallowed, pairInts{1, 1})
								disallowed = append(disallowed, pairInts{1, -1})
							}

						}

					}

				}
			}

			// Compute the bestDX and bestDY to move in
			shortestDistance := math.Inf(1)
			bestDX := -1
			bestDY := -1
			for dX := -1; dX <= 1; dX++ {
				for dY := -1; dY <= 1; dY++ {

					// Skip looking at myself
					if dX == 0 && dY == 0 {
						continue
					}

					// Check for disallowed deltas
					if hasWallInBorder {
						allow := true
						for _, disallowedPairInts := range disallowed {
							if dX == disallowedPairInts.fst && dY == disallowedPairInts.snd {
								allow = false
							}
						}
						if !allow {
							continue
						}
					}

					newX := i + dX
					newY := j + dY

					if newX >= 0 && newX < w.GetWidth() &&
						newY >= 0 && newY < w.GetHeight() {

						newTile := w.GetTile(newX, newY)
						if newTile.Walkable() && newTile.Dists[dest] < shortestDistance {
							shortestDistance = newTile.Dists[dest]
							bestDX = dX
							bestDY = dY
						}

					}

				}
			}

			// Translate this to a direction
			if tile.Directions == nil {
				tile.Directions = make(map[Destination]Direction)
			}
			tile.Directions[dest] = deltaToDirection(bestDX, bestDY)

		}
	}

	return nil

}

func deltaToDirection(dX, dY int) Direction {

	if dX == -1 && dY == -1 {
		return DirectionNW
	} else if dX == 0 && dY == -1 {
		return DirectionN
	} else if dX == 1 && dY == -1 {
		return DirectionNE
	} else if dX == -1 && dY == 0 {
		return DirectionW
	} else if dX == 1 && dY == 0 {
		return DirectionE
	} else if dX == -1 && dY == 1 {
		return DirectionSW
	} else if dX == 0 && dY == 1 {
		return DirectionS
	} else if dX == 1 && dY == 1 {
		return DirectionSE
	} else {
		return DirectionUnknown
	}

}
