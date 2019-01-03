package main

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"math"
	"os"
	"strings"
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
			world.tiles[x][y].destID = destID(c)
			if world.tiles[x][y].blockedNorth || world.tiles[x][y].blockedEast || world.tiles[x][y].blockedSouth || world.tiles[x][y].blockedWest {
				log.Println("BLOCKING")
			}
		}
	}
	fmt.Println("tiles size", len(world.tiles), len(world.tiles[0]))

	return world
}

func destID(c color.Color) uint32 {
	_, g, _, _ := c.RGBA()
	return g
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
	return sameColour(c, color.RGBA{R: 255, G: 0, B: 255, A: 255})
}
func blockedSouth(c color.Color) bool {
	return sameColour(c, color.RGBA{R: 251, G: 0, B: 7, A: 255})
}
func blockedWest(c color.Color) bool {
	return sameColour(c, color.RGBA{R: 120, G: 0, B: 120, A: 255})
}

type noFlowFieldsError struct {
}

func (e noFlowFieldsError) Error() string {
	return "No FlowFileds found"
}

func (s *State) LoadFlowField(dest DestinationID) error {
	destination := s.scenario.GetDestination(dest)
	filename := strings.Replace(destination.Name, " ", "-", -1)
	path := fmt.Sprintf("%s/flowfields/%s.png", s.ScenarioName, filename)
	file, err := os.Open(path)
	if err != nil {
		log.Println("Cannot open, ", path, err)
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("Unable to close file properly")
		}
	}()

	newImage, err := png.Decode(file)
	if err != nil {
		log.Println("error decoding png")
		return err
	}

	for y := 0; y < s.GetHeight(); y++ {
		for x := 0; x < s.GetWidth(); x++ {
			tile := s.GetTile(x, y)
			colour := newImage.At(x, y)
			dir, des := decode(colour)
			if tile.Directions == nil {
				tile.Directions = make(map[DestinationID]utils.OptionalFloat64)
			}
			tile.Directions[dest] = dir
			if tile.Dists == nil {
				tile.Dists = make(map[DestinationID]float64)
			}
			tile.Dists[dest] = des
		}
	}
	return nil

}

func (s *State) SaveFlowField(dest DestinationID) error {
	err := os.MkdirAll(fmt.Sprintf("%s/flowfields/", s.ScenarioName), 0777)
	if err != nil {
		log.Println("Cannot open or make directory, ", err)
		return err
	}
	destination := s.scenario.GetDestination(dest)
	filename := strings.Replace(destination.Name, " ", "-", -1)
	file, err := os.OpenFile(fmt.Sprintf("%s/flowfields/%s.png", s.ScenarioName, filename), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println("Cannot open or make file, ", err)
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("Unable to close file properly")
		}
	}()
	newImage := image.NewNRGBA(image.Rect(0, 0, s.GetWidth(), s.GetHeight()))
	exists := false
	for y := 0; y < s.GetHeight(); y++ {
		for x := 0; x < s.GetWidth(); x++ {
			tile := s.GetTile(x, y)
			optionalDir, ok := tile.Directions[dest]
			distance, ok2 := tile.Dists[dest]
			if ok && ok2 {
				colour := encode(optionalDir, distance)
				newImage.Set(x, y, colour)
				if !exists && tile.Directions != nil && tile.Dists != nil {
					exists = true
				}
			}
		}
	}

	if !exists {
		return noFlowFieldsError{}
	}

	maxDirDiff := 0.0
	maxDisDiff := 0.0
	for y := 0; y < s.GetHeight(); y++ {
		for x := 0; x < s.GetWidth(); x++ {
			tile := s.GetTile(x, y)
			optionalDir, ok := tile.Directions[dest]
			distance, ok2 := tile.Dists[dest]
			if ok && ok2 {
				colour := encode(optionalDir, distance)
				checkDir, checkDis := decode(colour)
				d, p := optionalDir.Value()
				d1, p1 := checkDir.Value()
				if p {
					if !p1 {
						log.Println("image direction present error")
					} else {
						dirDiff := math.Abs(d - d1)
						if dirDiff > math.Pi {
							dirDiff = (2 * math.Pi) - dirDiff
						}
						if dirDiff > maxDirDiff {
							maxDirDiff = dirDiff
						}
						disDiff := math.Abs(distance - checkDis)
						if disDiff > maxDisDiff {
							maxDisDiff = disDiff
						}
					}
				}
				newImage.Set(x, y, colour)
			}
		}
	}

	log.Println("max direction error is ", maxDirDiff)
	log.Println("max distance error is ", maxDisDiff)

	err = png.Encode(file, newImage)
	if err != nil {
		log.Println("cannot encode PNG")
		return err
	}
	return nil
}

func decode(colour color.Color) (utils.OptionalFloat64, float64) {
	ri, gi, bi, ai := colour.RGBA()
	if ai == 0 && ri == 0 && gi == 0 && bi == 0 {
		return utils.OptionalFloat64WithEmptyValue(), 0
	}
	r := float64(ri) / 0xffff
	g := float64(gi) / 0xffff
	b := float64(bi) / 0xffff
	max := r
	min := r
	if g > max {
		max = g
	} else if g < min {
		min = g
	}
	if b > max {
		max = b
	} else if b < min {
		min = b
	}

	var h float64
	if max == min {
		h = 0
	} else if max == r {
		h = 60.0 * (0.0 + ((g - b) / (max - min)))
	} else if max == g {
		h = 60.0 * (2.0 + ((b - r) / (max - min)))
	} else if max == b {
		h = 60.0 * (4.0 + ((r - g) / (max - min)))
	}

	if h < 0 {
		h += 360.0
	}

	/*s := 0.0
	if max != 0 {
		s = (max - min) / max
	}*/

	v := max

	dirh := ((h / 360.0) * (2 * math.Pi)) - math.Pi
	if dirh > math.Pi {
		dirh -= 2 * math.Pi
	} else if dirh < -math.Pi {
		dirh += 2 * math.Pi
	}

	//dirs := (s * (2*math.Pi)) - math.Pi

	//dir := (dirh + dirs)/2

	dis := (asigmoid(v) * -200) + 200

	return utils.OptionalFloat64WithValue(dirh), dis
}

func encode(direction utils.OptionalFloat64, distance float64) color.NRGBA {
	dir, p := direction.Value()
	if !p {
		return color.NRGBA{}
	}

	h := (dir + math.Pi) / (2 * math.Pi) * 360
	s := 1.0
	v := sigmoid(-(distance - 200) / 200.0)

	c := v * s
	hp := h / 60.0
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	m := v - c

	if hp <= 1 {
		return color.NRGBA{R: enc(c + m), G: enc(x + m), B: enc(m), A: 255}
	} else if hp <= 2 {
		return color.NRGBA{R: enc(x + m), G: enc(c + m), B: enc(m), A: 255}
	} else if hp <= 3 {
		return color.NRGBA{R: enc(m), G: enc(c + m), B: enc(x + m), A: 255}
	} else if hp <= 4 {
		return color.NRGBA{R: enc(m), G: enc(x + m), B: enc(c + m), A: 255}
	} else if hp <= 5 {
		return color.NRGBA{R: enc(x + m), G: enc(m), B: enc(c + m), A: 255}
	} else if hp <= 6 {
		return color.NRGBA{R: enc(c + m), G: enc(m), B: enc(x + m), A: 255}
	} else {
		log.Println("no idea whats happening")
	}

	return color.NRGBA{}
}

func enc(f float64) uint8 {
	return uint8(f * 255)
}

func sigmoid(a float64) float64 {
	return 1.0 / (1.0 + math.Exp(-a))
}

func asigmoid(a float64) float64 {
	return -math.Log((1.0 / a) - 1)
}
