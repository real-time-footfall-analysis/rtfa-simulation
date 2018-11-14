package directions

type Direction int

const (
	DirectionN Direction = iota
	DirectionS
	DirectionE
	DirectionW
)

type Destination struct {
	x int
	y int
}

type Tile struct {
	nextDirection Direction
	walkable      bool
}

type MacroMap struct {
	tiles [][]Tile
}

type FlowField MacroMap

func generateFlowField(macroMap MacroMap, destination Destination) (FlowField, Error) {

	//

}
