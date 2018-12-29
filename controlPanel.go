package main

import (
	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/colornames"
	"image"
	"image/draw"
	"log"
	"time"
)

type ControlPanel struct {
	s     screen.Screen
	root  *widget.Sheet
	world *State
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

type Button struct {
	node.ShellEmbed
	icon    []byte
	onClick func()
	z       iconvg.Rasterizer
	uniform *widget.Uniform
	label   *widget.Label
}

func NewButton(text string, icon []byte, onClick func()) *Button {
	w := &Button{
		icon: icon,
	}
	fn := func() {
		onClick()
		w.uniform.ThemeColor = theme.StaticColor(colornames.Forestgreen)
		w.uniform.Mark(node.MarkNeedsPaintBase)
	}
	w.onClick = fn
	w.Wrapper = w
	flow := widget.NewFlow(widget.AxisHorizontal)
	flow.Insert(widget.NewSizer(unit.Ems(0.5), unit.Value{}, nil), nil)

	w.label = widget.NewLabel(text)
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

func (p *ControlPanel) start(s screen.Screen) {

	vf := widget.NewFlow(widget.AxisVertical)

	p.root = widget.NewSheet(
		widget.NewUniform(theme.StaticColor(colornames.White),
			widget.NewPadder(widget.AxisBoth, unit.Ems(1), vf)))

	vf.Insert(NewButton("Generate Flow Fields", icons.MapsMap, func() {
		log.Println("TODO:: Generate Flow Fields")
		time.Sleep(1 * time.Second)
	}), nil)

	vf.Insert(NewButton("Run Simulation", icons.MapsMap, func() {
		log.Println("TODO:: Starting Simulation")
	}), nil)

	p.s = s

	newtheme := theme.Theme{}

	go func() {
		//widget.RunWindow(p.s, p.root, nil)
		err := widget.RunWindow(s, p.root, &widget.RunWindowOptions{
			NewWindowOptions: screen.NewWindowOptions{
				Title: "Simulation control",
			},
			Theme: newtheme})

		if err != nil {
			log.Fatalln("error: ", err)
		}
	}()

}
