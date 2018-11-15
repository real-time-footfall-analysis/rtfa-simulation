package world

import (
	"math"
)

const (
	PersonRadius             = 0.2
	TwicePersonRadiusSquared = (PersonRadius * 2) * (PersonRadius * 2)
)

func (s *State) movementintersects(x, y float64, theta float64, distance float64) (bool, float64, float64) {
	tx, ty := int(x), int(y)
	nx := x + distance*math.Cos(theta)
	ny := y + distance*math.Sin(theta)
	collided := false
	// Possibly worry about running off the edge of the map?
	if nx-math.Floor(x) < PersonRadius {
		// intersect the x-- tile
		if !s.GetTile(tx-1, ty).Walkable() {
			nx = math.Floor(x) + PersonRadius
			collided = true
		}
	} else if nx-math.Floor(x) > 1-PersonRadius {
		// intersect the x++ tile
		if !s.GetTile(tx+1, ty).Walkable() {
			nx = math.Floor(x) + 1 - PersonRadius
			collided = true
		}
	}
	if ny-math.Floor(y) < PersonRadius {
		// intersect the y-- tile
		if !s.GetTile(tx, ty-1).Walkable() {
			ny = math.Floor(y) + PersonRadius
			collided = true
		}
	} else if ny-math.Floor(y) > 1-PersonRadius {
		// intersect the y++ tile
		if !s.GetTile(tx, ty+1).Walkable() {
			ny = math.Floor(y) + 1 - PersonRadius
			collided = true
		}
	}

	if collided {
		distance = math.Sqrt(math.Pow(x-nx, 2) + math.Pow(y-ny, 2))
		theta = math.Atan2(nx-x, ny-y)
	}

	// for all people in this and adjacent tiles
	for ix := -1; ix < 2; ix++ {
		for iy := -1; iy < 2; iy++ {
			tile := s.GetTile(tx+ix, ty+iy)
			for i, _ := range tile.People {
				ax, ay := tile.People[i].Loc.GetLatestXY()
				if !(ax == x && ay == y) && intersect(nx, ny, ax, ay) {
					collided = true
					cx, cy := closestPointOnLine(x, y, nx, ny, ax, ay)

					closestSquared := math.Pow(cx-ax, 2) + math.Pow(cy-ay, 2)
					backdist := math.Sqrt(TwicePersonRadiusSquared - closestSquared)

					distance = math.Dim(distance, backdist)

					if distance == 0 {
						return true, x, y
					}
					nx = x + distance*math.Cos(theta)
					ny = y + distance*math.Sin(theta)
				}
			}
		}
	}
	return collided, nx, ny
}

func intersect(ax, ay, bx, by float64) bool {
	return math.Pow(ax-bx, 2)+math.Pow(ay-by, 2) < TwicePersonRadiusSquared
}

func closestPointOnLine(x, y, nx, ny, ax, ay float64) (float64, float64) {
	a := ny - y
	b := x - nx
	c1 := (ny-y)*x + (x-nx)*y
	c2 := -b*ax + a*ay
	det := a*a - (-b)*b
	cx := ax
	cy := ay
	if det != 0 {
		cx = (a*c1 - b*c2) / det
		cy = (a*c2 - (-b)*c1) / det
	}
	return cx, cy

}
