package directions

import (
	"fmt"
	"image"
	"image/color"
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

type Tile struct {
	Walkable   bool                      // If this cell is a wall or not
	RegionIds  []int                     // The region ID's this tile is in
	Dists      map[Destination]float64   // Used internally for dijkstra
	Directions map[Destination]Direction // Direction to destination
	X          int                       // X co-ordinate
	Y          int                       // Y co-ordinate
}

type MacroMap struct {
	Width      int
	Height     int
	TileWidth  float64  // Width of a tile
	tiles      [][]Tile // The tiles.
	background image.Image
}

func Init() {

	initDeltas()

}

func (d Direction) String() string {

	if d == DirectionN {
		return "↑"
	}
	if d == DirectionE {
		return "→"
	}
	if d == DirectionS {
		return "↓"
	}
	if d == DirectionW {
		return "←"
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

// func Run() {

// 	mm := LoadFromImage("test_world.png")

// 	mm.generateFlowField(Destination{38, 0})

// }

// func LoadFromImage(path string) MacroMap {

// 	reader, err := os.Open(path)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer reader.Close()
// 	i, _, err := image.Decode(reader)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	width := i.Bounds().Dx()
// 	height := i.Bounds().Dy()
// 	fmt.Println("image size:", width, height)
// 	world := MacroMap{Width: width, Height: height, background: i, tiles: make([][]Tile, width)}

// 	for x := 0; x < world.Width; x++ {
// 		world.tiles[x] = make([]Tile, world.Height)
// 		for y := 0; y < world.Height; y++ {
// 			walk := walkable(i.At(x, y))
// 			world.tiles[x][y].Walkable = walk
// 			world.tiles[x][y].X = x
// 			world.tiles[x][y].Y = y
// 		}
// 	}
// 	fmt.Println("tiles size", len(world.tiles), len(world.tiles[0]))

// 	return world
// }

func walkable(colour color.Color) bool {
	r, g, b, a := colour.RGBA()
	return r == 0xffff && g == 0xffff && b == 0xffff && a == 0xffff
}

func (mm *MacroMap) GetTileHighRes(x, y float64) (*Tile, error) {

	return mm.GetTile(int(math.Floor(x/mm.TileWidth)), int(math.Floor(y/mm.TileWidth)))

}

func (mm *MacroMap) GetTile(x, y int) (*Tile, error) {
	if x < 0 || x >= mm.Width || y < 0 || y >= mm.Height {
		return nil, fmt.Errorf("Invalid co-ordinates (%d, %d)", x, y)
	}
	return &mm.tiles[x][y], nil
}

func (mm *MacroMap) Print(destination Destination) {

	for i := 0; i < mm.Height; i++ {
		for j := 0; j < mm.Width; j++ {

			tile, _ := mm.GetTile(j, i)
			// fmt.Print(tile.Directions[destination])
			fmt.Printf("%04.f ", tile.Dists[destination])

		}
		fmt.Println()

	}
	fmt.Println()

}

// Generates the flow field to the destination
func (mm *MacroMap) generateFlowField(destination Destination) error {

	for y := 0; y < mm.Height; y++ {
		for x := 0; x < mm.Width; x++ {
			tile, _ := mm.GetTile(x, y)
			if tile.Walkable {
				fmt.Print(" ")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println()
	}
	fmt.Println()

	FindShortestPath(mm, destination)

	mm.Print(destination)

	return nil

}

// Assuming that distance information has been filled in by dijkstra,
// calculate the directions needed
// func (*mm MacroMap) computeDirections(dest Destination) error {

// 	// If a wall is in a square's border, don't consider diagonals

// 	for int i = 0; i < mm.Width; ++i {
// 		for int j = 0; j < mm.Height; ++j {

// 			tile, err := mm.GetTile(i, j)
// 			if err != nil {
// 				log.Println(err)
// 				exit(1)
// 			}

// 		}
// 	}

// }

// If there is a wall in the border, don't consider diagonals
