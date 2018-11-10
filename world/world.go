package world

import (
	"fmt"
	"image"
)

type Tile struct {
	walkable bool
}

func (t *Tile) Walkable() bool {
	return t.walkable
}

type State struct {
	tiles      [][]Tile
	width      int
	height     int
	background image.Image
}

func (w *State) GetWidth() int {
	return w.width
}

func (w *State) GetHeight() int {
	return w.height
}

func (w *State) GetImage() image.Image {
	return w.background
}

func (w *State) GetTile(x, y int) *Tile {
	if x < 0 || x >= w.width || y < 0 || y >= w.height {
		fmt.Println("out of range: %i %i", x, y)
	}
	return &(w.tiles[x][y])
}
