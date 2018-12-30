package main

import (
	"fmt"
	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
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
	"image/draw"
	"log"
)

type ControlPanel struct {
	s     screen.Screen
	root  *widget.Sheet
	world *State
	w     screen.Window
}

type panelUpdate struct {
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

type Icon struct {
	node.LeafEmbed
	icon []byte
	z    iconvg.Rasterizer
}

func NewIcon(icon []byte) *Icon {
	w := &Icon{
		icon: icon,
	}
	w.Wrapper = w
	return w
}

func (w *Icon) Measure(t *theme.Theme, widthHint, heightHint int) {
	px := t.Pixels(unit.Ems(2)).Ceil()
	w.MeasuredSize = image.Point{px, px}
}

func (w *Icon) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	w.Marks.UnmarkNeedsPaintBase()
	w.z.SetDstImage(ctx.Dst, w.Rect.Add(origin), draw.Over)
	return iconvg.Decode(&w.z, w.icon, nil)
}

type Ticker struct {
	node.ShellEmbed
	tick  func() string
	label *widget.Label
}

func (p *ControlPanel) NewTicker(text string, tick func() string) *Ticker {
	w := &Ticker{
		tick: tick,
	}
	w.Wrapper = w
	flow := widget.NewFlow(widget.AxisHorizontal)
	flow.Insert(widget.NewSizer(unit.Ems(0.5), unit.Value{}, nil), nil)

	flow.Insert(widget.NewLabel(fmt.Sprintf("%-30s", text)), nil)
	flow.Insert(widget.NewSizer(unit.Ems(0.5), unit.Value{}, nil), nil)
	w.label = widget.NewLabel(fmt.Sprintf("%30s", ""))
	flow.Insert(w.label, nil)

	uniform := widget.NewUniform(theme.StaticColor(colornames.Aqua), widget.NewPadder(widget.AxisBoth, unit.Ems(0.5), flow))
	padding := widget.NewPadder(widget.AxisBoth, unit.Ems(0.5), uniform)
	w.Insert(padding, nil)

	go func() {
		for {
			newString := w.tick()
			w.label.Text = fmt.Sprintf("%-30s", newString)
			w.label.Mark(node.MarkNeedsPaintBase)
			p.w.Send(update{})
		}
	}()

	return w
}

type Button struct {
	node.ShellEmbed
	icon    []byte
	onClick func()
	z       iconvg.Rasterizer
	uniform *widget.Uniform
	label   *widget.Label
	pressed bool
}

func (p *ControlPanel) NewButton(text string, icon []byte, toggle bool, onClick func() string) *Button {
	w := &Button{
		icon: icon,
	}
	fn := func() {
		w.pressed = !w.pressed
		w.label.Text = fmt.Sprintf("%-30s", onClick())
		w.label.Mark(node.MarkNeedsPaintBase)
		if w.pressed || !toggle {
			w.uniform.ThemeColor = theme.StaticColor(colornames.Forestgreen)
		} else {
			w.uniform.ThemeColor = theme.StaticColor(colornames.Lightcoral)
		}
		w.uniform.Mark(node.MarkNeedsPaintBase)
		p.w.Send(panelUpdate{})

	}
	w.onClick = fn
	w.Wrapper = w
	flow := widget.NewFlow(widget.AxisHorizontal)
	flow.Insert(widget.NewSizer(unit.Ems(0.5), unit.Value{}, nil), nil)

	w.label = widget.NewLabel(fmt.Sprintf("%-30s", text))
	flow.Insert(w.label, nil)
	flow.Insert(widget.NewSizer(unit.Ems(0.5), unit.Value{}, nil), nil)
	flow.Insert(NewIcon(icon), nil)

	w.uniform = widget.NewUniform(theme.StaticColor(colornames.Lightcoral), flow)
	padding := widget.NewPadder(widget.AxisBoth, unit.Ems(0.5), w.uniform)
	w.Insert(padding, nil)
	return w
}

func (w *Button) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	switch e := e.(type) {
	case gesture.Event:
		if e.Type != gesture.TypeTap {
			break
		}
		if w.onClick != nil {
			w.uniform.ThemeColor = theme.StaticColor(colornames.Orange)
			w.uniform.Mark(node.MarkNeedsPaintBase)
			go w.onClick()
		}
		return node.Handled
	}
	return node.NotHandled
}

func (p *ControlPanel) start(s screen.Screen, world *State) {

	p.world = world
	vf := widget.NewFlow(widget.AxisVertical)

	p.root = widget.NewSheet(
		widget.NewUniform(theme.StaticColor(colornames.White),
			widget.NewPadder(widget.AxisBoth, unit.Ems(1), vf)))

	vf.Insert(p.NewGenrateFlowFieldsButton(), nil)

	vf.Insert(p.NewStartSimulationButton(), nil)

	vf.Insert(p.NewTicker("Total People:", func() string { return fmt.Sprintf("%d", <-p.world.peopleCurrentChan) }), nil)
	vf.Insert(p.NewTicker("Total People Added:", func() string { return fmt.Sprintf("%d", <-p.world.peopleAddedChan) }), nil)
	vf.Insert(p.NewTicker("Simulation Time:", func() string { return (<-p.world.simulationTimeChan).String() }), nil)

	for i, _ := range p.world.scenario.Destinations {
		dest := &p.world.scenario.Destinations[i]
		button := p.NewButton(fmt.Sprintf("Close %s", dest.Name), icons.NavigationClose, true, func() string {
			if dest.isClosed() {
				dest.Open()
				return fmt.Sprintf("Close %s", dest.Name)
			} else {
				dest.Close()
				return fmt.Sprintf("Reopen %s", dest.Name)
			}
		})

		vf.Insert(button, nil)
	}

	p.s = s

	newtheme := theme.Theme{}

	go func() {
		//widget.RunWindow(p.s, p.root, nil)
		err := p.RunWindow(&widget.RunWindowOptions{
			NewWindowOptions: screen.NewWindowOptions{
				Title: "Simulation control",
			},
			Theme: newtheme})

		if err != nil {
			log.Fatalln("error: ", err)
		}
	}()

}

func (p *ControlPanel) NewGenrateFlowFieldsButton() *Button {
	pressed := false

	return p.NewButton("Generate Flow Fields", icons.MapsMap, false, func() string {
		if pressed {
			return "Generate Flow Fields"
		}
		pressed = true
		log.Println("Generate Flow Fields")

		InitFlowFields()
		for _, dest := range p.world.scenario.Destinations {
			log.Println("Flow field for", dest.Name, "starting")

			err := p.world.GenerateFlowField(dest.ID)
			log.Println("Flow field for", dest.Name, "done")
			if err != nil {
				log.Fatal("cannot make flow field for", dest)
			}
		}
		log.Println("Flow fields done")
		return "Generate Flow Fields"

	})
}

func (p *ControlPanel) NewStartSimulationButton() *Button {
	pressed := false
	return p.NewButton("Run Simulation", icons.ActionBuild, false, func() string {
		if pressed {
			return "Run Simulation"
		}
		pressed = true
		log.Println("Starting Simulation")
		p.world.startWaiter <- true
		return "Run Simulation"
	})
}

func (p *ControlPanel) RunWindow(opts *widget.RunWindowOptions) error {
	var (
		nwo *screen.NewWindowOptions
		t   *theme.Theme
	)
	if opts != nil {
		nwo = &opts.NewWindowOptions
		t = &opts.Theme
	}
	var err error
	p.w, err = p.s.NewWindow(nwo)
	if err != nil {
		return err
	}
	defer p.w.Release()

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
			p.root.OnInputEvent(e, image.Point{})

		case paint.Event:
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

		case panelUpdate:

		case error:
			return e
		}

		if !paintPending && p.root.Wrappee().Marks.NeedsPaint() {
			paintPending = true
			p.w.Send(paint.Event{})
		}
	}
}
