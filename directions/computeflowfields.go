package directions

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
	NextDirection Direction // The next direction someone should follow to go to the destination from this cell
	Walkable      bool      // If this cell is a wall or not
	RegionIds     []int     // The region ID's this tile is in
	Dist          int       // This is only used internally for dijkstra
}

type MacroMap struct {
	Width  int
	Height int
	tiles  [][]Tile // The tiles.
}

func (mm *MacroMap) GetTile(x, y int) (*Tile, error) {
	return &mm.tiles[y][x], nil
}

type FlowField MacroMap

// Map of directions to their flow fields
var FlowFields = make(map[Direction]*FlowField)

func generateFlowField(macroMap MacroMap, destination Destination) (FlowField, error) {

	//

}
