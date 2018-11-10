package render

import (
	"fmt"
	"github.com/real-time-footfall-analysis/rtfa-simulation/world"
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
	"log"
	"math"
)

type RenderState struct {
	s               screen.Screen
	w               screen.Window
	b               screen.Buffer
	t               screen.Texture
	sz              size.Event
	windowScale     float64
	backgroundScale int
	i               image.Image
	mousePressed    bool
}

type UpdateEvent struct {
	World *world.State
}

func SetupRender(s screen.Screen, original image.Image) RenderState {
	r := RenderState{s: s}

	r.backgroundScale = 10
	window, err := s.NewWindow(nil)
	if err != nil {
		log.Fatal(err)
	}
	r.w = window

	nx, ny := r.loadOriginal(original)

	size0 := image.Point{nx, ny}
	//size0 := image.Point{400, 300}
	b, err := s.NewBuffer(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.b = b

	t, err := s.NewTexture(size0)
	if err != nil {
		log.Fatal(err)
	}
	r.t = t

	bufferImage(r.b, r.i)
	r.t.Upload(image.Point{}, r.b, r.b.Bounds())

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

func (r *RenderState) resetImage() {
	bufferImage(r.b, r.i)
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
		r.loadOriginal(e.World.GetImage())
		r.resetImage()
		r.Redraw()

		// run through and draw people to buffer
	}
	return true
}

func (r *RenderState) Redraw() {
	// Set background
	r.w.Fill(r.sz.Bounds(), color.Transparent, screen.Src)

	// Upload buffer to texture
	r.t.Upload(image.Point{}, r.b, r.b.Bounds())

	// Create transformation matrix
	src2dst := f64.Aff3{
		r.windowScale, 0, 0,
		0, r.windowScale, 0,
	}

	// Draw texture to window
	r.w.Draw(src2dst, r.t, r.t.Bounds(), screen.Over, nil)
}

func (r *RenderState) SetTileColour(px, py int, colour color.Color) {
	for xi := 0; xi < r.backgroundScale; xi++ {
		for yi := 0; yi < r.backgroundScale; yi++ {
			r.b.RGBA().Set(px+xi, py+yi, colour)
		}
	}
}

func (r *RenderState) GetPixelPos(px, py int) (int, int) {
	px = px * r.backgroundScale
	py = py * r.backgroundScale
	return px, py
}

func (r *RenderState) GetWorldPos(e mouse.Event) (int, int) {
	px := int(float64(e.X)/r.windowScale) / r.backgroundScale
	py := int(float64(e.Y)/r.windowScale) / r.backgroundScale
	return px, py
}

func (r *RenderState) Release() {
	r.t.Release()
	r.b.Release()
	r.w.Release()
}

func bufferImage(buffer screen.Buffer, i image.Image) {

	dst := buffer.RGBA()
	draw.Draw(dst, dst.Bounds(), i, image.Point{}, screen.Src)
}
