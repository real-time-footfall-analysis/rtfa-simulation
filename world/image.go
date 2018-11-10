package world

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
)

func LoadFromImage(path string) State {

	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	i, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	width := i.Bounds().Dx()
	height := i.Bounds().Dy()
	fmt.Println("image size:", width, height)
	world := State{width: width, height: height, background: i, tiles: make([][]Tile, width)}

	for x := 0; x < world.width; x++ {
		world.tiles[x] = make([]Tile, world.height)
		for y := 0; y < world.height; y++ {
			walk := walkable(i.At(x, y))
			world.tiles[x][y].walkable = walk
		}
	}
	fmt.Println("tiles size", len(world.tiles), len(world.tiles[0]))

	return world
}

func walkable(colour color.Color) bool {
	r, g, b, a := colour.RGBA()
	return r == 0xffff && g == 0xffff && b == 0xffff && a == 0xffff
}
