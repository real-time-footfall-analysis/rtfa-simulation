package main

import (
	"bufio"
	"fmt"
	"golang.org/x/exp/shiny/driver/gldriver"
	"golang.org/x/exp/shiny/screen"
	draw2 "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
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
}

type UpdateEvent struct {
	World *State
}

func SetupRender(s screen.Screen, original image.Image, regions *[]Region) RenderState {
	r := RenderState{s: s}

	context, err := gldriver.NewContext()
	if err != nil {
		log.Fatalln("cannot make gl context", err)
	}
	log.Println("GL error:", context.GetError())

	r.backgroundScale = 1
	window, err := s.NewWindow(nil)
	if err != nil {
		log.Fatal(err)
	}
	r.w = window
	//rectangle := image.Rect(0,0,100,100)
	//r.w.Fill(rectangle, color.Black, screen.Src)
	r.w.Publish()

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
		red, green, blue := color.YCbCrToRGB(uint8(100), uint8(rand.Intn(256)), uint8(rand.Intn(256)))
		c := color.RGBA{red, green, blue, 100}
		drawRegionInBuffer(&r, region.X, region.Y, c, region.Radius)
	}

	r.rt.Upload(image.Point{}, r.rb, r.rb.Bounds())

	r.windowScale = float64(1)

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
			} else if e.Direction == mouse.DirRelease {
				r.mousePressed = false
			}
		}
		if r.mousePressed {
			px, py := r.GetPixelPos(r.GetWorldPos(e))
			r.SetTileColour(px, py, color.Black)
			if r.world != nil {
				tx, ty := r.GetWorldPos(e)
				if tx >= 0 {
					r.world.GetTile(tx, ty).SetWalkable(false)
				}
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

var tick int = 0

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
	r.w.Publish()

	tick++
	if tick%500 != 0 {
		return
	}
	log.Println("redraw")
	f, err := os.Create("/tmp/dat3.png")
	if err != nil {
		log.Println("error0", err)
		err = os.Remove("/tmp/dat3.png")
		if err != nil {
			log.Fatalln("error1", err)
		}
		f, err = os.Create("/tmp/dat3.png")
		if err != nil {
			log.Fatalln("error2", err)
		}

	}
	defer f.Close()

	w := bufio.NewWriter(f)
	bi := r.bb.RGBA().SubImage(r.bb.Bounds())
	i := r.b.RGBA().SubImage(r.b.Bounds())
	ri := r.rb.RGBA().SubImage(r.rb.Bounds())
	newImage := image.NewRGBA(image.Rect(0, 0, r.b.Bounds().Dx(), r.b.Bounds().Dy()))
	draw2.NearestNeighbor.Scale(newImage, newImage.Bounds(), bi, bi.Bounds(), draw2.Src, nil)
	draw2.NearestNeighbor.Scale(newImage, newImage.Bounds(), ri, ri.Bounds(), draw2.Over, nil)
	draw2.NearestNeighbor.Scale(newImage, newImage.Bounds(), i, i.Bounds(), draw2.Over, nil)
	err = png.Encode(w, newImage)
	if err != nil {
		log.Fatalln("error3", err)
	}
	err = w.Flush()
	if err != nil {
		log.Fatalln("error4", err)
	}

}

func (r *RenderState) SetTileColour(px, py int, colour color.Color) {
	for xi := 0; xi < r.backgroundScale; xi++ {
		for yi := 0; yi < r.backgroundScale; yi++ {
			r.bb.RGBA().Set(px+xi, py+yi, colour)
		}
	}
	r.bt.Upload(image.Point{}, r.bb, r.bb.Bounds())

}

func (r *RenderState) GetPixelPos(px, py int) (int, int) {
	px = px * r.backgroundScale
	py = py * r.backgroundScale
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
