package chat

import (
	"realy.lol/gui/color"
	"gioui.org/op/clip"
	"gioui.org/layout"
	"realy.lol/gui"
)

type Panel struct {
	r *Root
	// when in small mode, top bar becomes navigation
	Small bo
}

func (p *Panel) Init(r *Root) *Panel {
	p.r = r
	return p
}

func (p *Panel) Layout(g Gx) Dim {
	if !p.Small {
		g.Constraints.Max.X = 360
		g.Constraints.Min.X = 360
	}
	l := Body1(p.r.th, "panel")
	FillShape(g.Ops, p.r.Palette.GetColor(color.PanelBg).NRGBA(),
		clip.Rect(Rectangle{Max: g.Constraints.Max}).Op())
	Flex{}.Layout(g, Flexed(1, func(g Gx) Dim {
		layout.UniformInset(Dp(l.TextSize)).Layout(g,
			func(g Gx) Dim {
				return gui.Centered{}.Layout(g, l.Layout)
			})
		return Dim{Size: g.Constraints.Max}
	}))
	return Dim{Size: g.Constraints.Max}
}
