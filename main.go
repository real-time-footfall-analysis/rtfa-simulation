package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
)

func main() {
	w := LoadFromImage("demo_complex.png")
	w.LoadRegions("testRegions.json", 51.506478, -0.172219)
	fmt.Println("state size:", w.GetWidth(), w.GetHeight())
	for y := 0; y < w.GetHeight(); y++ {
		for x := 0; x < w.GetWidth(); x++ {
			if w.GetTile(x, y).Walkable() {
				fmt.Print(" ")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println()
	}
	fmt.Println()

	driver.Main(func(s screen.Screen) {
		r := SetupRender(s, w.GetImage())
		defer r.Release()

		go simulate(&w, &r)

		for r.Step() {
		}
		fmt.Println("EOL")

	})

	fmt.Println("Bottom")

}

func simulate(world *State, r *RenderState) {
	//time.Sleep(5 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	go func() {
		people := 0
		steps := 0

		// Add random people

		groups := make([]*Group, 2000)
		for i, _ := range groups {
			groups[i] = &Group{
				Individuals:  make([]*Individual, 0),
			}
		}
		for i := 0; i < 10000; i++ {
			indiv := world.AddRandom()
			groupIndex := rand.Intn(2000)
			groups[groupIndex].Individuals = append(groups[groupIndex].Individuals, indiv)
			people++
		}

		InitFlowFields()

		world.GenerateFlowField(Destination{
			X: 20,
			Y: 180,
		})

		world.GenerateFlowField(Destination{
			X: 180,
			Y: 20,
		})

		world.GenerateFlowField(Destination{
			X: 180,
			Y: 180,
		})

		world.GenerateFlowField(Destination{
			X: 20,
			Y: 20,
		})

		world.GenerateFlowField(Destination{
			X: 100,
			Y: 100,
		})

		world.GenerateFlowField(Destination{
			X: 20,
			Y: 130,
		})
		//world.PrintDistances(Destination{
		//	X: 820,
		//	Y: 520,
		//})
		//world.PrintDirections(Destination{
		//	X: 820,
		//	Y: 520,
		//})

		// Set up parallel processing channels
		channels := make([]chan map[*Individual]utils.OptionalFloat64, 0)
		for i := 0; i < len(groups); i++ {
			channel := make(chan map[*Individual]utils.OptionalFloat64)
			channels = append(channels, channel)
		}

		var avg float64 = -1
		for t := range ticker.C {

			//fmt.Println(steps, "Tick at", t)

			// world.MoveAll()

			// Get the desired positions for each of the individuals in each group in parallel
			for i, group := range groups {
				go group.Next(channels[i], world)
			}

			// Process each group as they come
			for _, channel := range channels {
				// Process them as they come in
				result := <-channel
				processMovementsForGroup(world, result)
			}

			if steps%1 == 0 {
				r.SendEvent(UpdateEvent{world})
			}
			//fmt.Println("people: ", people)
			steps++
			geometry.FlipTick()
			dt := time.Since(t).Nanoseconds()
			if avg < 0 {
				avg = float64(dt)
			} else {
				avg = 0.9*avg + 0.1*float64(dt)
			}

			if steps%50 == 0 {
				fmt.Println("average tick time: ", avg/1000000000)
			}

		}
	}()
	time.Sleep(600 * time.Second)
	ticker.Stop()
	fmt.Println("Ticker stopped")
}

func processMovementsForGroup(world *State, movements map[*Individual]utils.OptionalFloat64) {
	for individual, direction := range movements {
		theta, ok := direction.Value()
		if !ok {
			// If we aren't meant to move... don't
			world.MoveIndividual(individual, 0, 0)
			continue
		}

		world.MoveIndividual(individual, theta, individual.StepSize)
	}
}
