package gel

import (
	"image"
	"image/color"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Rect struct {
	Color color.NRGBA
	Size  image.Point
	Radii int
}

func (r Rect) Layout(g Gx) Dim {
	paint.FillShape(
		g.Ops,
		r.Color,
		clip.UniformRRect(
			image.Rectangle{
				Max: r.Size,
			},
			r.Radii,
		).Op(g.Ops))
	return Dim{Size: r.Size}
}
