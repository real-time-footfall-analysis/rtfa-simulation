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
	Tiles [][]Tile // The tiles.
}

type FlowField MacroMap

func generateFlowField(macroMap MacroMap, destination Destination) (FlowField, error) {

	//

}
