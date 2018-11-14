package directions

import (
	"fmt"
	"math"
)

type pairInts struct {
	fst, snd int
}
type Direction int

const (
	DirectionN Direction = iota
	DirectionS
	DirectionE
	DirectionW
	DirectionUnknown
)

type Destination struct { // Indicies into the macromap
	X int
	Y int
}

type Tile struct {
	NextDirection Direction                 // The next direction someone should follow to go to the destination from this cell
	Walkable      bool                      // If this cell is a wall or not
	RegionIds     []int                     // The region ID's this tile is in
	Dists         map[Destination]float64   // Used internally for dijkstra
	Directions    map[Destination]Direction // Direction to destination
	X             int                       // X co-ordinate
	Y             int                       // Y co-ordinate
}

type MacroMap struct {
	Width     int
	Height    int
	TileWidth float64  // Width of a tile
	tiles     [][]Tile // The tiles.
}

// Maps pairs of (x, y) co-ordinates (+-1 or 0) to directions
var deltasToDirection map[pairInts]Direction

func Init() {

	initDeltasToDirection()

}

func initDeltasToDirection() {

	deltasToDirection = make(map[pairInts]Direction)
	deltasToDirection[pairInts{0, 1}] = DirectionN
	deltasToDirection[pairInts{1, 0}] = DirectionE
	deltasToDirection[pairInts{0, -1}] = DirectionS
	deltasToDirection[pairInts{-1, 0}] = DirectionW

}

func (mm *MacroMap) GetTileHighRes(x, y float64) (*Tile, error) {

	return mm.GetTile(int(math.Floor(x/mm.TileWidth)), int(math.Floor(y/mm.TileWidth)))

}

func (mm *MacroMap) GetTile(x, y int) (*Tile, error) {
	if x < 0 || x >= mm.Width || y < 0 || y >= mm.Height {
		return nil, fmt.Errorf("Invalid co-ordinates (%d, %d)", x, y)
	}
	return &mm.tiles[y][x], nil
}

func generateFlowField(macroMap MacroMap, destination Destination) error {

	//

}
