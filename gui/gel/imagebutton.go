package gel

import (
	"image"
	widget "widget.mleku.dev"
)

type ImageButtonStyle struct {
	Size        Dp
	Button      *Clickable
	Description string
	image.Image
}

func ImageButton(th *Theme, button *Clickable, img image.Image, description string) *ImageButtonStyle {
	return &ImageButtonStyle{
		Image:       img,
		Size:        Dp(th.TextSize) * 3,
		Button:      button,
		Description: description,
	}
}

func (b ImageButtonStyle) Layout(g Gx) Dim {
	g.Constraints.Min = Point{int(b.Size), int(b.Size)}
	return widget.ClipCircle{}.Layout(g, func(g Gx) Dim {
		Fill(g.Ops, NRGBA{255, 0, 0, 255})

		return Dim{Size: Point{X: int(b.Size), Y: int(b.Size)}}
	})
}
