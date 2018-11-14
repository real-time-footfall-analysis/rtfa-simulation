package directions

type Direction int

const (
	DirectionN Direction = iota
	DirectionS
	DirectionE
	DirectionW
)

type Destination struct {
	X int
	Y int
}

type Tile struct {
	NextDirection Direction
	Walkable      bool
	Dist          int // This is only used internally for dijkstra
}

type MacroMap struct {
	tiles [][]Tile
}

type FlowField MacroMap

func generateFlowField(macroMap MacroMap, destination Destination) (FlowField, error) {

	//

}
