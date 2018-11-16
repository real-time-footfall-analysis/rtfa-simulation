package main

import (
	"fmt"
	"time"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/group"
	"github.com/real-time-footfall-analysis/rtfa-simulation/individual"
	"github.com/real-time-footfall-analysis/rtfa-simulation/render"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
	"github.com/real-time-footfall-analysis/rtfa-simulation/world"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
)

func main() {
	w := world.LoadFromImage("test5.png")
	w.LoadRegions("testRegions.json", 53.867225, -1.380985)
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
		r := render.SetupRender(s, w.GetImage())
		defer r.Release()

		go simulate(&w, &r)

		for r.Step() {
		}
		fmt.Println("EOL")

	})

	fmt.Println("Bottom")

}

func simulate(world *world.State, r *render.RenderState) {
	//time.Sleep(5 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	go func() {
		people := 0
		steps := 0

		// Add random people
		for i := 0; i < 6; i++ {
			world.AddRandom()
			people++

		}

		// Add them to groups
		// TODO:
		groups := make([]*group.Group, 0)

		// Set up parallel processing channels
		channels := make([]chan map[*individual.Individual]utils.OptionalFloat64, 0)
		for i := 0; i < len(groups); i++ {
			channel := make(chan map[*individual.Individual]utils.OptionalFloat64)
			channels = append(channels, channel)
		}

		var avg float64 = -1
		for t := range ticker.C {

			//fmt.Println(steps, "Tick at", t)

			// world.MoveAll()

			// Get the desired positions for each of the individuals in each group in parallel
			for i, group := range groups {
				go group.Next(channels[i])
			}

			// Process each group as they come
			for _, channel := range channels {
				// Process them as they come in
				result := <-channel
				processMovementsForGroup(world, result)
			}

			if steps%1 == 0 {
				r.SendEvent(render.UpdateEvent{world})
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

func processMovementsForGroup(world *world.State, movements map[*individual.Individual]utils.OptionalFloat64) {
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
