package main

import (
	"fmt"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/colornames"
	draw2 "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"math/rand"
)

type RenderState struct {
	s               screen.Screen
	w               screen.Window
	b               screen.Buffer
	t               screen.Texture
	bb              screen.Buffer
	bt              screen.Texture
	rb              screen.Buffer
	rt              screen.Texture
	sz              size.Event
	windowScale     float64
	backgroundScale int
	i               image.Image
	mousePressed    bool
	world           *State
	info            ControlPanel

	highlight highlight
}

type highlight struct {
	infoSet    bool
	colour     color.Color
	oldX, oldY int
}

type UpdateEvent struct {
	World *State
}

func SetupRender(s screen.Screen, original image.Image, regions *[]Region, state *State) RenderState {
	r := RenderState{s: s}
	r.world = state

	r.backgroundScale = 1
	window, err := s.NewWindow(nil)
	if err != nil {
		log.Fatal(err)
	}
	r.w = window

	nx, ny := r.loadOriginal(original)

	size0 := image.Point{nx, ny}

	b, err := s.NewBuffer(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.b = b

	bb, err := s.NewBuffer(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.bb = bb

	rb, err := s.NewBuffer(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.rb = rb

	t, err := s.NewTexture(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.t = t

	bt, err := s.NewTexture(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.bt = bt

	rt, err := s.NewTexture(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.rt = rt

	bufferImage(r.bb, r.i)
	r.bt.Upload(image.Point{}, r.bb, r.bb.Bounds())

	for _, region := range *regions {
		fmt.Println("region: ", region.Name, region.X, region.Y)
		red, green, blue := color.YCbCrToRGB(uint8(50), uint8(rand.Intn(256)), uint8(rand.Intn(256)))
		c := color.RGBA{red, green, blue, 150}
		drawRegionInBuffer(&r, region.X, region.Y, c, region.Radius)
	}

	r.rt.Upload(image.Point{}, r.rb, r.rb.Bounds())

	r.windowScale = float64(1)

	r.info.start(s)

	return r
}

func (r *RenderState) loadOriginal(original image.Image) (int, int) {
	nx := original.Bounds().Dx() * r.backgroundScale
	ny := original.Bounds().Dy() * r.backgroundScale
	newImage := image.NewRGBA(image.Rect(0, 0, nx, ny))
	draw2.NearestNeighbor.Scale(newImage, newImage.Bounds(), original, original.Bounds(), draw2.Src, nil)
	r.i = newImage
	return nx, ny
}

func (r *RenderState) resetPeopleBuffer() {
	size := r.b.Size()
	r.b.Release()
	b, err := r.s.NewBuffer(size)
	if err != nil {
		log.Fatal(err)
	}
	r.b = b
}

func (r *RenderState) SendEvent(e interface{}) {
	r.w.Send(e)
}

func (r *RenderState) Step() bool {
	e := r.w.NextEvent()

	/* This print message is to help programmers learn what events this
	// example program generates. A real program shouldn't print such
	// messages; they're not important to end users.
	format := "got %#v\n"
	if _, ok := e.(fmt.Stringer); ok {
		format = "got %v\n"
	}
	fmt.Printf(format, e)
	*/

	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {

			fmt.Println("returning")
			return false
		}
	case mouse.Event:
		if e.Button == mouse.ButtonLeft {
			if e.Direction == mouse.DirPress {
				r.mousePressed = true
				if r.highlight.infoSet {
					log.Println("test")
					px, py := r.GetPixelPos(r.highlight.oldX, r.highlight.oldY)
					r.SetTileColour(px, py, r.highlight.colour)
				}
				r.highlight.infoSet = true
				r.highlight.oldX, r.highlight.oldY = r.GetWorldPos(e)
				px, py := r.GetPixelPos(r.highlight.oldX, r.highlight.oldY)
				r.highlight.colour = r.GetTileColour(px, py)
				r.SetTileColour(px, py, colornames.Darkslategrey)

				if r.world != nil {
					log.Println("send event")
					tile := r.world.GetTile(r.highlight.oldX, r.highlight.oldY)
					log.Println("tile: " + string(tile.X) + string(tile.Y))
				}

			} else if e.Direction == mouse.DirRelease {
				r.mousePressed = false
			}
			r.Redraw()
		}

	case size.Event:
		r.sz = e
		xs := float64(r.sz.WidthPx) / float64(r.b.Bounds().Dx())
		ys := float64(r.sz.HeightPx) / float64(r.b.Bounds().Dy())
		r.windowScale = math.Min(xs, ys)
	case paint.Event:
		r.Redraw()
	case UpdateEvent:
		r.resetPeopleBuffer()
		r.world = e.World
		for x := 0; x < e.World.GetWidth(); x++ {
			for y := 0; y < e.World.GetHeight(); y++ {
				tile := e.World.GetTile(x, y)
				for _, p := range tile.People {
					//fmt.Print(p)
					x, y := p.Loc.GetLatestXY()
					drawPersonInBuffer(r, x, y, p.Colour)
				}
			}
		}

		r.Redraw()
		// run through and draw people to buffer
	}
	return true
}

func drawPersonInBuffer(r *RenderState, x, y float64, c color.Color) {
	ix := int(x * float64(r.backgroundScale))
	iy := int(y * float64(r.backgroundScale))

	p := 3
	pr := r.backgroundScale / p
	if pr == 0 {
		pr = 1
	}
	for px := -pr; px < pr; px++ {
		for py := -pr; py < pr; py++ {
			if px*px+py*py < pr*pr {
				r.b.RGBA().Set(ix+px, iy+py, c)
			}
		}
	}

	//r.b.RGBA().Set(ix, iy, c)
}

func drawRegionInBuffer(r *RenderState, x, y float64, c color.Color, rad int) {
	ix := int(x * float64(r.backgroundScale))
	iy := int(y * float64(r.backgroundScale))
	pr := r.backgroundScale * rad
	if pr < 2 {
		pr = 2
	}

	for px := -pr; px < pr; px++ {
		for py := -pr; py < pr; py++ {
			if px*px+py*py < pr*pr {
				r.rb.RGBA().Set(ix+px, iy+py, c)
			}
		}
	}
}

func (r *RenderState) Redraw() {
	// Set background
	//r.w.Fill(r.sz.Bounds(), color.Transparent, screen.Src)

	// Upload buffer to texture
	r.t.Upload(image.Point{}, r.b, r.b.Bounds())

	// Create transformation matrix
	src2dst := f64.Aff3{
		r.windowScale, 0, 0,
		0, r.windowScale, 0,
	}

	// Draw texture to window
	r.w.Draw(src2dst, r.bt, r.bt.Bounds(), screen.Over, nil)
	r.w.Draw(src2dst, r.rt, r.rt.Bounds(), screen.Over, nil)
	r.w.Draw(src2dst, r.t, r.t.Bounds(), screen.Over, nil)
}

func (r *RenderState) SetTileColour(px, py int, colour color.Color) {
	for xi := 0; xi < r.backgroundScale; xi++ {
		for yi := 0; yi < r.backgroundScale; yi++ {
			r.bb.RGBA().Set(px+xi, py+yi, colour)
		}
	}
	r.bt.Upload(image.Point{}, r.bb, r.bb.Bounds())

}

func (r *RenderState) GetTileColour(px, py int) color.Color {
	return r.bb.RGBA().At(px, py)
}

func (r *RenderState) GetPixelPos(x, y int) (int, int) {
	px := x * r.backgroundScale
	py := y * r.backgroundScale
	return px, py
}

func (r *RenderState) GetWorldPos(e mouse.Event) (int, int) {
	px := int(float64(e.X) / r.windowScale)
	if px < 0 || px >= r.i.Bounds().Dx() {
		return -1, -1
	}
	px /= r.backgroundScale
	py := int(float64(e.Y) / r.windowScale)
	if py < 0 || py >= r.i.Bounds().Dy() {
		return -1, -1
	}
	py /= r.backgroundScale
	return px, py
}

func (r *RenderState) Release() {
	r.t.Release()
	r.b.Release()
	r.bb.Release()
	r.bt.Release()
	r.rb.Release()
	r.rt.Release()
	r.w.Release()
}

func bufferImage(buffer screen.Buffer, i image.Image) {

	dst := buffer.RGBA()
	draw.Draw(dst, dst.Bounds(), i, image.Point{}, screen.Src)
}
