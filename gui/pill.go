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
	v, h := is*2/3, is*4/3
	i := Inset{v, v, h, h}
	return f.Layout(g, func(g Gx) Dim {
		d := i.Layout(g, p.Widget)
		return d
	})
}
