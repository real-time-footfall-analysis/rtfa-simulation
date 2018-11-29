package main

import (
	"fmt"
	"golang.org/x/image/colornames"
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
			c := i.At(x, y)
			if !(sameColour(color.White, c) || sameColour(color.Black, c)) {
				log.Println("colour", c)
			}
			walk := walkable(c)
			world.tiles[x][y].walkable = walk
			world.tiles[x][y].X = x
			world.tiles[x][y].Y = y
			world.tiles[x][y].blockedNorth = blockedNorth(c)
			world.tiles[x][y].blockedEast = blockedEast(c)
			world.tiles[x][y].blockedSouth = blockedSouth(c)
			world.tiles[x][y].blockedWest = blockedWest(c)
			if world.tiles[x][y].blockedNorth || world.tiles[x][y].blockedEast || world.tiles[x][y].blockedSouth || world.tiles[x][y].blockedWest {
				log.Println("BLOCKING")
			}
		}
	}
	fmt.Println("tiles size", len(world.tiles), len(world.tiles[0]))

	return world
}

func sameColour(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func walkable(colour color.Color) bool {
	r, g, b, a := colour.RGBA()
	return !(r == 0x0000 && g == 0x0000 && b == 0x0000 && a == 0xffff)
}

func blockedNorth(c color.Color) bool {
	return sameColour(c, color.RGBA{R: 0, G: 0, B: 255, A: 255})
}
func blockedEast(c color.Color) bool {
	return sameColour(c, colornames.Yellow)
}
func blockedSouth(c color.Color) bool {
	return sameColour(c, color.RGBA{R: 251, G: 0, B: 7, A: 255})
}
func blockedWest(c color.Color) bool {
	return sameColour(c, colornames.Green)
}
