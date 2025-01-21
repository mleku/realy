package chat

import (
	"realy.lol/gui/color"
	"gioui.org/op/clip"
	"gioui.org/layout"
)

type Panel struct {
	r *Root
	// when in small mode, top bar becomes navigation
	Small bo
	*PanelHeader
}

func (p *Panel) Init(r *Root) *Panel {
	p.r = r
	p.PanelHeader = new(PanelHeader).Init(r)
	return p
}

func (p *Panel) Layout(g Gx) Dim {
	if !p.Small {
		g.Constraints.Max.X = 360
		g.Constraints.Min.X = 360
	}
	l := Body1(p.r.th, "panel")
	FillShape(g.Ops, p.r.Palette.GetColor(color.Primary).NRGBA(),
		clip.Rect(Rectangle{Max: g.Constraints.Max}).Op())
	return Flex{}.Layout(g, Flexed(1, func(g Gx) Dim {
		layout.UniformInset(Dp(l.TextSize)/2).Layout(g,
			func(g Gx) Dim {
				return Flex{Axis: Vertical}.Layout(g,
					Rigid(func(g Gx) Dim {
						return p.PanelHeader.Layout(g)
					}),
					Flexed(1, func(g Gx) Dim {
						return Dim{Size: g.Constraints.Max}
					}),
				)
			},
		)
		return Dim{Size: g.Constraints.Max}
	}))
	// return Dim{Size: g.Constraints.Max}
}
