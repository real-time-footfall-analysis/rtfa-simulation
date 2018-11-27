package main

import (
	"errors"
	"log"
	"math"

	"github.com/jupp0r/go-priority-queue"
)

var deltas []pairInts

func initDeltas() {

	deltas = make([]pairInts, 0)
	deltas = append(deltas, pairInts{-1, 0})
	deltas = append(deltas, pairInts{0, -1})
	deltas = append(deltas, pairInts{1, 0})
	deltas = append(deltas, pairInts{0, 1})

}

func FindShortestPath(w *State, destination Destination) error {

	// Queue to hold fringe verticies
	queue, _ := initQueue(w, destination)

	// While there are tiles in the queue
	for queue.Len() > 0 {

		// Get the closest fringe vertex
		item, err := queue.Pop()
		tile, ok := item.(*Tile)
		if !ok {
			log.Panicln("Failed to parse queue item into tile")
			return errors.New("Failed to parse queue item into tile")
		}
		if err != nil {
			log.Panicln(err)
			return err
		}

		// Relax each neighbouring tile
		neighbours := getValidNeighbouringTiles(tile.X, tile.Y, w)
		for _, neighbour := range neighbours {
			if (tile.Dists[destination] + 1) < neighbour.Dists[destination] {
				neighbour.Dists[destination] = tile.Dists[destination] + 1
				queue.UpdatePriority(neighbour, neighbour.Dists[destination])
			}
		}

	}

	return nil

}

// Initialises queue, adding all nodes and setting distance to infinity
// Nodes which we can reach
func initQueue(w *State, dest Destination) (*pq.PriorityQueue, error) {

	// Initialise the queue
	queue := pq.New()

	// Insert all the tiles into the queue with distance infinity
	for i := 0; i < w.GetWidth(); i++ {
		for j := 0; j < w.GetHeight(); j++ {

			tile := w.GetTile(i, j)

			if tile.Dists == nil {
				tile.Dists = make(map[Destination]float64)
			}

			// Set the direction for un-walkable tiles to be none
			if !tile.Walkable() {
				tile.Dists[dest] = math.Inf(1)
			}

			// Skip the destination and tiles we can't walk on
			if !tile.Walkable() || (i == dest.X && j == dest.Y) {
				continue
			}

			tile.Dists[dest] = math.Inf(1)
			queue.Insert(tile, tile.Dists[dest])

		}

	}

	// Insert the destination with distance 0
	destTile := w.GetTile(dest.X, dest.Y)
	if destTile.Dists == nil {
		destTile.Dists = make(map[Destination]float64)
	}
	if destTile.Directions == nil {
		destTile.Directions = make(map[Destination]Direction)
	}
	destTile.Dists[dest] = 0
	queue.Insert(destTile, destTile.Dists[dest])

	return &queue, nil

}

func getValidNeighbouringTiles(x, y int, w *State) []*Tile {

	tiles := make([]*Tile, 0)

	for _, delta := range deltas {

		newX := x + delta.fst
		newY := y + delta.snd

		if newX >= 0 && newX < w.GetWidth() &&
			newY >= 0 && newY < w.GetHeight() {
			tile := w.GetTile(newX, newY)
			// Skip tiles we can't go to
			if !tile.Walkable() {
				continue
			}
			tiles = append(tiles, tile)
		}
	}

	return tiles
}
