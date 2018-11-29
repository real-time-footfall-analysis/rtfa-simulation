package main

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
	_ "image/png"
	"log"
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
	R float64
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
			angle, present := tile.Directions[destination].Value()
			if !present || angle == math.Inf(1) {
				fmt.Print("## ")
			} else {
				fmt.Printf("%02.f ", angle)
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

			destToPrint := tile.Dists[destination]
			if destToPrint == math.Inf(1) {
				fmt.Print("## ")
			} else {
				fmt.Printf("%02.f ", destToPrint)
			}

		}
		fmt.Println()

	}
	fmt.Println()

}

func (w *State) PrintDestTiles(destination Destination) {

	for i := 0; i < w.GetHeight(); i++ {
		for j := 0; j < w.GetWidth(); j++ {

			tile := w.GetTile(j, i)

			if tile.DestTile == nil {
				fmt.Print(" #### ")
			} else {
				fmt.Printf("[%d, %d]", tile.DestTile.X, tile.DestTile.Y)
			}

		}
		fmt.Println()

	}
	fmt.Println()

}

// Generates the flow field to the destination
func (w *State) GenerateFlowField(destination Destination) error {
	log.Println("find shorted path")
	FindShortestPath(w, destination)
	log.Println("compute directions")
	w.computeDirections(destination)

	return nil

}

// Assuming that distance information has been filled in by dijkstra,
// calculate the directions needed
func (w *State) computeDirections(dest Destination) {

	for i := 0; i < w.GetWidth(); i++ {
		for j := 0; j < w.GetHeight(); j++ {
			computeDirectionForTile(i, j, dest, w)
		}
	}

}
func computeDirectionForTile(x int, y int, dest Destination, w *State) {

	tile := w.GetTile(x, y)

	if tile.Directions == nil {
		tile.Directions = make(map[Destination]utils.OptionalFloat64)
	}

	// Skip walls
	if !tile.Walkable() {
		tile.Directions[dest] = utils.OptionalFloat64WithEmptyValue()
		tile.DestTile = nil
		return
	}

	// Find the next as-the-crow-flies destination tile
	destX := x
	destY := y

	// Work out whether we should look up or down
	stepY := 0
	validUpCoord := w.validCoord(destX, destY+1)
	validDownCoord := w.validCoord(destX, destY-1)
	if !validUpCoord && !validDownCoord {
		stepY = 0
	} else if !validUpCoord && w.GetTile(destX, destY-1).Dists[dest] < tile.Dists[dest] {
		stepY = -1
	} else if !validDownCoord && w.GetTile(destX, destY+1).Dists[dest] < tile.Dists[dest] {
		stepY = 1
	} else if validUpCoord && validDownCoord &&
		w.GetTile(destX, destY+1).Dists[dest] < w.GetTile(destX, destY-1).Dists[dest] &&
		w.GetTile(destX, destY+1).Dists[dest] < tile.Dists[dest] {
		stepY = 1
	} else if validUpCoord && validDownCoord &&
		w.GetTile(destX, destY-1).Dists[dest] < w.GetTile(destX, destY+1).Dists[dest] &&
		w.GetTile(destX, destY-1).Dists[dest] < tile.Dists[dest] {
		stepY = -1
	}

	// Find how far we have to go in the y-axis
	if stepY != 0 {
		for newDestY := destY + stepY; w.validCoord(destX, newDestY) && w.GetTile(destX, newDestY).Walkable() && w.GetTile(destX, newDestY).Dists[dest] < w.GetTile(destX, destY).Dists[dest]; newDestY += stepY {
			destY = newDestY
		}
	}

	// Work out whether we should look left or right
	stepX := 0
	validRightCoord := w.validCoord(destX+1, destY)
	validLeftCoord := w.validCoord(destX-1, destY)
	if !validRightCoord && !validLeftCoord {
		stepX = 0
	} else if !validRightCoord && w.GetTile(destX-1, destY).Dists[dest] < w.GetTile(destX, destY).Dists[dest] {
		stepX = -1
	} else if !validLeftCoord && w.GetTile(destX+1, destY).Dists[dest] < w.GetTile(destX, destY).Dists[dest] {
		stepX = 1
	} else if validLeftCoord && validRightCoord &&
		w.GetTile(destX+1, destY).Dists[dest] < w.GetTile(destX-1, destY).Dists[dest] &&
		w.GetTile(destX+1, destY).Dists[dest] < w.GetTile(destX, destY).Dists[dest] {
		stepX = 1
	} else if validLeftCoord && validRightCoord &&
		w.GetTile(destX-1, destY).Dists[dest] < w.GetTile(destX+1, destY).Dists[dest] &&
		w.GetTile(destX-1, destY).Dists[dest] < w.GetTile(destX, destY).Dists[dest] {
		stepX = -1
	}

	// Find out how far we have to go in the x-axis
	if stepX != 0 {
		for newDestX := destX + stepX; w.validCoord(newDestX, destY) && w.GetTile(newDestX, destY).Walkable() && w.GetTile(newDestX, destY).Dists[dest] < w.GetTile(destX, destY).Dists[dest] && w.noWallsInCol(destX, y, destY); newDestX += stepX {
			destX = newDestX
		}
	}

	// TODO: Follow this angle UNTIL the difference between the
	// suggested angle at the square you are at and the current angle is greater than some threshold
	// Add randomness to the angle chosen - but take this into account when using the threshold^

	tile.DestTile = &Destination{
		X: destX,
		Y: destY,
		// radius not used for debuging stuff
	}
	tile.Directions[dest] = utils.OptionalFloat64WithValue(math.Atan2(float64(destY-y), float64(destX-x)))

}

func (w *State) validCoord(x, y int) bool {
	return x >= 0 && y >= 0 && y < w.GetHeight() && x < w.GetWidth()
}

func (w *State) noWallsInCol(x, y1, y2 int) bool {

	maxY := y1
	minY := y2
	if maxY < minY {
		maxY = y2
		minY = y1
	}

	for y := minY; y <= maxY; y++ {
		if !w.validCoord(x, y) || !w.GetTile(x, y).Walkable() {
			return false
		}
	}

	return true
}

// Converts dX, dY to a direction:
// flipping the y axis
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
