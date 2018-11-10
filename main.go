package main

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/render"
	"github.com/real-time-footfall-analysis/rtfa-simulation/world"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"time"
)

func main() {
	w := world.LoadFromImage("test.png")
	fmt.Println(w)

	fmt.Println("state size: %i %i", w.GetWidth(), w.GetHeight())
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
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		steps := 0
		for t := range ticker.C {
			fmt.Println(steps, "Tick at", t)

			if steps%2 == 0 && steps < 10 {
				for i := 0; i < 100; i++ {
					world.AddRandom()
				}
			} else {
				for i := 0; i < 100*steps; i++ {
					world.MoveRandom()
				}
			}
			r.SendEvent(render.UpdateEvent{world})
			steps++
		}
	}()
	time.Sleep(60 * time.Second)
	ticker.Stop()
	fmt.Println("Ticker stopped")
}
