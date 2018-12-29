package main

import (
	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/colornames"
	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"image"
	"log"
)

type pannel struct {
	s    screen.Screen
	w    screen.Window
	root *widget.Sheet
	tile *Tile
}

type NewTileEvent struct {
	tile *Tile
}

type UpdatePannel struct {
}

func stretch(n node.Node, alongWeight int) node.Node {
	return widget.WithLayoutData(n, widget.FlowLayoutData{
		AlongWeight:  alongWeight,
		ExpandAlong:  true,
		ShrinkAlong:  true,
		ExpandAcross: true,
		ShrinkAcross: true,
	})
}
func (p *pannel) start(s screen.Screen) {

	vf := widget.NewFlow(
		widget.AxisVertical,
		widget.NewLabel("TODO: status"),
		widget.NewSpace(),
		widget.NewSheet(widget.NewText("hello")),
		widget.NewText("hello2"),
	)
	p.root = widget.NewSheet(vf)

	p.s = s
	var nwo *screen.NewWindowOptions
	w, err := s.NewWindow(nwo)
	if err != nil {
		log.Fatalln("error: ", err)
	}
	p.w = w

	go func() {
		//widget.RunWindow(p.s, p.root, nil)
		err := p.runWindow()
		if err != nil {
			log.Fatalln("error: ", err)
		}
	}()
}

func (p *pannel) runWindow() error {
	var (
		t *theme.Theme
	)
	// paintPending batches up multiple NeedsPaint observations so that we
	// paint only once (which can be relatively expensive) even when there are
	// multiple input events in the queue, such as from a rapidly moving mouse
	// or from the user typing many keys.
	//
	// TODO: determine somehow if there's an external paint event in the queue,
	// not just internal paint events?
	//
	// TODO: if every package that uses package screen should basically
	// throttle like this, should it be provided at a lower level?
	paintPending := false

	gef := gesture.EventFilter{EventDeque: p.w}
	for {
		e := p.w.NextEvent()

		if e = gef.Filter(e); e == nil {
			continue
		}

		switch e := e.(type) {
		case lifecycle.Event:
			p.root.OnLifecycleEvent(e)
			if e.To == lifecycle.StageDead {
				return nil
			}

		case gesture.Event, mouse.Event:
			log.Println("mouse")
			if p.tile != nil {
				p.root = p.recompileWidget()
			}
			p.root.OnInputEvent(e, image.Point{})

		case paint.Event:
			log.Println("paint! ")
			p.w.Fill(image.Rect(100, 100, 200, 200), colornames.Wheat, screen.Src)
			ctx := &node.PaintContext{
				Theme:  t,
				Screen: p.s,
				Drawer: p.w,
				Src2Dst: f64.Aff3{
					1, 0, 0,
					0, 1, 0,
				},
			}
			if err := p.root.Paint(ctx, image.Point{}); err != nil {
				return err
			}
			p.w.Publish()
			paintPending = false

		case size.Event:
			if dpi := float64(e.PixelsPerPt) * unit.PointsPerInch; dpi != t.GetDPI() {
				newT := new(theme.Theme)
				if t != nil {
					*newT = *t
				}
				newT.DPI = dpi
				t = newT
			}

			size := e.Size()
			p.root.Measure(t, size.X, size.Y)
			p.root.Wrappee().Rect = e.Bounds()
			p.root.Layout(t)
			// TODO: call Mark(node.MarkNeedsPaint)?

		case NewTileEvent:
			log.Println("new tile")
			p.tile = e.tile
			p.w.Send(UpdatePannel{})

		case UpdatePannel:
			log.Println("update pannel")
			p.root = p.recompileWidget()
			p.root.Mark(node.MarkNeedsPaintBase | node.MarkNeedsPaint)

		case error:
			return e
		}

		if !paintPending && p.root.Wrappee().Marks.NeedsPaint() {
			paintPending = true
			p.w.Send(paint.Event{})
		}
	}
}

func (p *pannel) recompileWidget() *widget.Sheet {
	str := "coords: " + string(p.tile.X) + string(p.tile.Y)
	log.Println(str)
	vf := widget.NewFlow(
		widget.AxisVertical,
		widget.NewText(str),
		widget.NewText("hello6"),
	)
	return widget.NewSheet(vf)
}
