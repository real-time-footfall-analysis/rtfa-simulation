package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/utils"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
)

func main() {
	w := LoadFromImage("hyde_park_quater1.png")
	w.BulkSend = true
	w.LoadRegions("testRegions.json", 51.506478, -0.172219)
	w.LoadScenario("scenario1.json")
	log.Println("loaded scenario")

	/*fmt.Println("state size:", w.GetWidth(), w.GetHeight())
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
	fmt.Println()*/

	driver.Main(func(s screen.Screen) {
		r := SetupRender(s, w.GetImage(), &w.Regions)
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
	log.Println("simulation starting")

	steps := 0

	world.groups = make([]*Group, world.scenario.TotalGroups)
	for i, _ := range world.groups {
		world.groups[i] = &Group{
			Individuals: make([]*Individual, 0),
		}
	}

	log.Println("Flow fields starting")
	InitFlowFields()
	for _, dest := range world.scenario.Destinations {
		log.Println("Flow field for", dest.X, ",", dest.Y, "starting")

		err := world.GenerateFlowField(Destination{
			X: dest.X,
			Y: dest.Y,
			R: dest.Radius,
		})
		log.Println("Flow field for", dest.X, ",", dest.Y, "done")
		if err != nil {
			log.Fatal("cannot make flow field for", dest)
		}
	}
	log.Println("Flow fields done")

	// Set up parallel processing channels
	channels := make([]chan map[*Individual]utils.OptionalFloat64, 0)
	for i := 0; i < len(world.groups); i++ {
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
			world.groups[groupIndex].Individuals = append(world.groups[groupIndex].Individuals, indiv)
			world.peopleAdded++
		}

		//fmt.Println(steps, "Tick at", t)

		// world.MoveAll()

		// Get the desired positions for each of the individuals in each group in parallel
		for i, group := range world.groups {
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
			SendBulk()
		}

	}

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

		UpdateServer(&world.Regions, individual, world.time, world.BulkSend)
	}
}
