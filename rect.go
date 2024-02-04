package main

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func NewRect(x int, y int, width int, height int) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  x + width,
		Height: y + height,
	}
}

func (r *Rect) Center() (int, int) {
	centerX := (r.X + r.Width) / 2
	centerY := (r.Y + r.Height) / 2
	return centerX, centerY
}

func (r *Rect) Intersect(other Rect) bool {
	return (r.X <= other.Width &&
		r.Width >= other.X &&
		r.Y <= other.Height &&
		r.Height >= other.Y)
}
