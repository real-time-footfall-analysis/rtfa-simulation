package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
)

func main() {
	w := LoadFromImage("test5.png")
	w.LoadRegions("testRegions.json", 51.506478, -0.172219)
	w.BulkSend = true
	w.LoadScenario("scenario1.json")

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

	simulate(&w)

}

func simulate(world *State) {
	//time.Sleep(5 * time.Second)

	steps := 0

	groups := make([]*Group, world.scenario.TotalGroups)
	for i, _ := range groups {
		groups[i] = &Group{
			Individuals: make([]*Individual, 0),
		}
	}

	InitFlowFields()

	for _, dest := range world.scenario.Destinations {
		err := world.GenerateFlowField(Destination{
			X: dest.X,
			Y: dest.Y,
		})
		if err != nil {
			log.Fatal("cannot make flow field for", dest)
		}
	}

	// Set up parallel processing channels
	channels := make([]chan map[*Individual]utils.OptionalFloat64, 0)
	for i := 0; i < len(groups); i++ {
		channel := make(chan map[*Individual]utils.OptionalFloat64)
		channels = append(channels, channel)
	}

	var avg float64 = -1
	for world.time.Before(world.scenario.End) {
		t := time.Now()

		// add more people until someone doesn't fit
		for i := world.peopleAdded; i < world.scenario.TotalPeople; i++ {
			indiv := world.AddRandom()
			if indiv == nil {
				break
			}
			groupIndex := rand.Intn(world.scenario.TotalGroups)
			groups[groupIndex].Individuals = append(groups[groupIndex].Individuals, indiv)
			world.peopleAdded++
		}

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
		//fmt.Println("people: ", people)
		steps++
		world.TickTime()
		geometry.FlipTick()

		// calculate simulation speed analytics
		dt := time.Since(t).Nanoseconds()
		if avg < 0 {
			avg = float64(dt)
		} else {
			avg = 0.9*avg + 0.1*float64(dt)
		}

		if steps%500 == 0 {
			fmt.Println("average tick time: ", avg/1000000000)
			fmt.Println("sim time: ", world.time)
		}

	}

	fmt.Println("Ticker stopped")
	SendBulk()
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

		UpdateServer(&world.Regions, individual, world.time, world.BulkSend)
	}
}
