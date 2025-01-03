package chat

import (
	"realy.lol/gui/color"
	"gioui.org/op/clip"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/widget"
)

type S []Relay

type Panel struct {
	r *Root
	// when in small mode, top bar becomes navigation
	Small bo
	*PanelHeader
	widget.List
	S
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
	FillShape(g.Ops, p.r.Palette.GetColor(color.PanelBg).NRGBA(),
		clip.Rect(Rectangle{Max: g.Constraints.Max}).Op())
	return Flex{}.Layout(g, Flexed(1, func(g Gx) Dim {
		layout.UniformInset(Dp(l.TextSize)/2).Layout(g,
			func(g Gx) Dim {
				return Flex{Axis: Vertical}.Layout(g,
					Rigid(func(g Gx) Dim {
						return p.PanelHeader.Layout(g)
					}),
					Flexed(1, func(g Gx) Dim {
						if p.S == nil {
							return Dim{}
						}
						return material.List(p.r.th, &p.List).Layout(g, 1, func(g Gx, item int) Dim {
							return Dim{}
							// return p.S[item]
						})
					}),
				)
			},
		)
		return Dim{Size: g.Constraints.Max}
	}))
}
