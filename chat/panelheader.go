package chat

import (
	"golang.org/x/exp/shiny/materialdesign/icons"
	"gioui.org/widget"
	"realy.lol/gui/component"
	"realy.lol/gui"
	"realy.lol/gui/color"
)

var MenuIcon = func() *Icon {
	icon, _ := widget.NewIcon(icons.NavigationMenu)
	return icon
}()

var CloseIcon = func() *Icon {
	icon, _ := widget.NewIcon(icons.NavigationClose)
	return icon
}()

type PanelHeader struct {
	r             *Root
	Active        st
	searchField   component.TextField
	menuClickable Clickable
	menuButton    *ButtonLayoutStyle
}

func (ph *PanelHeader) Init(r *Root) *PanelHeader {
	ph.r = r
	size := Dp(ph.r.th.TextSize)
	ph.menuButton = &ButtonLayoutStyle{
		Background:   NRGBA{},
		Button:       &ph.menuClickable,
		CornerRadius: size / 2,
	}
	return ph
}

func (ph *PanelHeader) Layout(g Gx) (d Dim) {
	dims := gui.GetDim(g, func(Gx) Dim { return ph.searchField.Layout(g, ph.r.th, "search") })
	Flex{Spacing: SpaceAround}.Layout(g,
		Rigid(func(g Gx) Dim {
			g.Constraints.Min.Y = dims.Size.Y * 8 / 7
			g.Constraints.Max.Y = g.Constraints.Min.Y
			g.Constraints.Min.X = g.Constraints.Min.Y
			g.Constraints.Max.X = g.Constraints.Min.Y
			ph.menuButton.Layout(g, func(g Gx) Dim {
				return MenuIcon.Layout(g, ph.r.GetColor(color.PanelText).NRGBA())
			})
			// material.IconButton(ph.r.th, &ph.menuClickable, MenuIcon, "close").Layout(g)
			return Dim{Size: g.Constraints.Min}
		}),
		Flexed(1, func(g Gx) Dim {
			g.Constraints.Max.Y = dims.Size.Y
			// g.Constraints.Max.Y = g.Constraints.Min.Y
			h := Dp(ph.r.th.TextSize) / 4
			return Inset{0, 0, h, h}.Layout(g, func(g Gx) Dim {
				return ph.searchField.Layout(g, ph.r.th, "search")
			})
		}),
	)
	return
}
