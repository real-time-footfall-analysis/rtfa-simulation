package geometry

var tick = true
var step = 1

// Here we use TickTock Positions so the state of the world is consistent to the actors as the last state while the
// collision detection is able to write a new state, FlipTick should be called once and only once at the end of each
// time step
type Point struct {
	ax   float64
	ay   float64
	bx   float64
	by   float64
	last bool
}

func (p Point) GetXY() (float64, float64) {
	if tick {
		return p.ax, p.ay
	} else {
		return p.bx, p.by
	}
}

func (p Point) GetLatestXY() (float64, float64) {
	if p.last {
		return p.ax, p.ay
	} else {
		return p.bx, p.by
	}
}

func (p *Point) SetXY(x, y float64) {
	if tick {
		p.bx = x
		p.by = y
		p.last = false
	} else {
		p.ax = x
		p.ay = y
		p.last = true
	}
}

func NewPoint(x, y float64) Point {
	return Point{ax: x, ay: y, bx: x, by: y, last: true}
}

type Value struct {
	ax   float64
	bx   float64
	last bool
}

func (p Value) Get() float64 {
	if tick {
		return p.ax
	} else {
		return p.bx
	}
}

func (p Value) GetLatest() float64 {
	if p.last {
		return p.ax
	} else {
		return p.bx
	}
}

func (p *Value) Set(x float64) {
	if tick {
		p.bx = x
		p.last = false
	} else {
		p.ax = x
		p.last = true
	}
}

func NewValue(x float64) Value {
	return Value{ax: x, bx: x, last: true}
}

func FlipTick() {
	tick = !tick
	step++
}
