package directions

import "math"

type Direction int

const (
	DirectionN Direction = iota
	DirectionS
	DirectionE
	DirectionW
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
}

type MacroMap struct {
	Width     int
	Height    int
	TileWidth float64  // Width of a tile
	tiles     [][]Tile // The tiles.
}

func (mm *MacroMap) GetTileHighRes(x, y float64) (*Tile, error) {

	return mm.GetTile(int(math.Floor(x/mm.TileWidth)), int(math.Floor(y/mm.TileWidth)))

}

func (mm *MacroMap) GetTile(x, y int) (*Tile, error) {
	return &mm.tiles[y][x], nil
}

func generateFlowField(macroMap MacroMap, destination Destination) error {

	//

}
