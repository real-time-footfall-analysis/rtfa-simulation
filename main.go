package main

import (
	"fmt"
	"time"

	"github.com/real-time-footfall-analysis/rtfa-simulation/directions"
	"github.com/real-time-footfall-analysis/rtfa-simulation/group"
	"github.com/real-time-footfall-analysis/rtfa-simulation/individual"
	"github.com/real-time-footfall-analysis/rtfa-simulation/geometry"
	"github.com/real-time-footfall-analysis/rtfa-simulation/render"
	"github.com/real-time-footfall-analysis/rtfa-simulation/world"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
)

func main() {
	w := world.LoadFromImage("test4.png")
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

	for y := 0; y < w.GetHeight(); y++ {
		for x := 0; x < w.GetWidth(); x++ {
			fmt.Print(int(w.GetTile(x, y).HitCount), ",")
		}
		fmt.Println()
	}
	fmt.Println()

	fmt.Println("Bottom")

}

func simulate(world *world.State, r *render.RenderState) {
	//time.Sleep(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		people := 0
		steps := 0
		for t := range ticker.C {
			fmt.Println(steps, "Tick at", t)

			if steps%2 == 0 && steps < 8 {
				for i := 0; i < 1; i++ {
					world.AddRandom()
					people++

				}
			} else {
				for i := 0; i < 1; i++ {
					fmt.Println("move")
					world.MoveRandom()
				}
			}
			r.SendEvent(render.UpdateEvent{world})
			fmt.Println("people: ", people)
			steps++
			geometry.FlipTick()
		}
	}()
	time.Sleep(600 * time.Second)
	ticker.Stop()
	fmt.Println("Ticker stopped")
}

func processTick(groups []*group.Group) {

	channels := make([]chan map[*individual.Individual]directions.Direction, 0)
	for i := 0; i < len(groups); i++ {
		channel := make(chan map[*individual.Individual]directions.Direction)
		channels = append(channels, channel)
	}

	// Get the desired positions for each of the individuals in each group in parallel
	for i, group := range groups {
		go group.Next(channels[i])
	}

	for _, channel := range channels {
		// Process them as they come in
		result := <-channel
		movePeopleIfPossible(result)
	}
}

func movePeopleIfPossible(map[*individual.Individual]directions.Direction) {
	// TODO:
}
