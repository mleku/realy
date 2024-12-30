package gui

type Pill struct {
	Color    NRGBA
	TextSize Sp
	Widget
}

func (p Pill) Layout(g Gx) Dim {
	ts := Dp(p.TextSize)
	f := Filler{p.Color, ts * 3 / 2}
	is := ts
	i := Inset{is * 2 / 3, is * 2 / 3, is * 4 / 3, is * 4 / 3}
	return f.Layout(g, func(g Gx) Dim {
		d := i.Layout(g, p.Widget)
		return d
	})
}
