package gui

import (
	"image"
	"gioui.org/op/paint"
	"gioui.org/op/clip"
)

// Filler fills the space inside a widget with an optional round corners.
type Filler struct {
	Color        NRGBA
	CornerRadius Dp
}

func (f Filler) Layout(g Gx, w Widget) (d Dim) {
	d = GetDim(g, w)
	sz := d.Size
	rr := g.Dp(f.CornerRadius)
	r := image.Rectangle{Max: sz}
	paint.FillShape(g.Ops,
		f.Color,
		clip.UniformRRect(r, rr).Op(g.Ops),
	)
	w(g)
	return
}
