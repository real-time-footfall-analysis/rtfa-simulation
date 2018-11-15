package directions

import (
	"log"
	"math"

	"github.com/jupp0r/go-priority-queue"
)

var deltas []pairInts

func initDeltas() {

	deltas := make([]pairInts, 0)
	deltas = append(deltas, pairInts{-1, 0})
	deltas = append(deltas, pairInts{0, -1})
	deltas = append(deltas, pairInts{1, 0})
	deltas = append(deltas, pairInts{0, 1})

}

func FindShortestPath(macroMap *MacroMap, destination Destination) error {

	// Queue to hold fringe verticies
	queue, _ := initQueue(macroMap, destination)

	// While there are tiles in the queue
	for queue.Len() > 0 {

		// Get the closest fringe vertex
		item, err := queue.Pop()
		tile, ok := item.(*Tile)
		if !ok {
			log.Panicln("Failed to parse queue item into tile")
			break
		}
		if err != nil {
			log.Panicln(err)
			break
		}

		// Relax each neighbouring tile
		neighbours := getValidNeighbouringTiles(tile.X, tile.Y, macroMap)
		for _, neighbour := range neighbours {
			if (tile.Dists[destination] + 1) < neighbour.Dists[destination] {
				neighbour.Dists[destination] = tile.Dists[destination] + 1
				queue.UpdatePriority(neighbour, distToPriority(neighbour.Dists[destination]))
				dX := tile.X - neighbour.X
				dY := tile.Y - neighbour.Y
				if !ok {
					log.Printf("Failed to convert (%d, %d) to direction", dX, dY)
					continue
				}
			}
		}

	}
	return nil
}

// Initialises queue, adding all nodes and setting distance to infinity
// Nodes which we can reach
func initQueue(macroMap *MacroMap, dest Destination) (*pq.PriorityQueue, error) {

	// Initialise the queue
	queue := pq.New()

	// Insert all the tiles into the queue with distance infinity
	for i := 0; i < macroMap.Width; i++ {
		for j := 0; j < macroMap.Height; j++ {

			tile, _ := macroMap.GetTile(i, j)
			// Skip the destination and tiles we can't walk on
			if tile.Walkable || (i == dest.X && j == dest.Y) {
				continue
			}

			tile.Dists[dest] = math.Inf(1)
			tile.Directions[dest] = DirectionUnknown
			queue.Insert(tile, distToPriority(tile.Dists[dest]))

		}

	}

	// Insert the destination with distance 0
	destTile, err := macroMap.GetTile(dest.X, dest.Y)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	queue.Insert(destTile, distToPriority(0))

	return &queue, nil

}

func getValidNeighbouringTiles(x, y int, macroMap *MacroMap) []*Tile {

	tiles := make([]*Tile, 0)

	for _, delta := range deltas {

		newX := x + delta.fst
		newY := y + delta.snd

		if newX >= 0 && newX < macroMap.Width &&
			newY >= 0 && newY < macroMap.Height {
			tile, err := macroMap.GetTile(newX, newY)
			if err != nil {
				log.Println(err)
				continue
			}
			// Skip tiles we can't go to
			if !tile.Walkable {
				continue
			}
			tiles = append(tiles, tile)
		}
	}

	return tiles
}

// Since pq is a max-heap and we need a min-heap, invert the priorities
func distToPriority(dist float64) float64 {
	return -dist
}
