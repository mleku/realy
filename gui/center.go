package gui

type Centered struct {
}

func (c Centered) Layout(g Gx, w Widget) (d Dim) {
	gd := GetDim(g, w)
	left := (g.Constraints.Max.X - gd.Size.X) / 2
	top := (g.Constraints.Max.Y - gd.Size.Y) / 2
	d = Inset{
		Left: Dp(left),
		Top:  Dp(top),
	}.Layout(g, w)
	return
}
