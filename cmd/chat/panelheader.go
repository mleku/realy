package main

import (
	"golang.org/x/exp/shiny/materialdesign/icons"
	"gioui.org/widget"
	"realy.lol/gui/gel"
	"gioui.org/text"
	col "realy.lol/gui/color"
)

var MenuIcon = func() *Icon {
	icon, _ := widget.NewIcon(icons.NavigationMenu)
	return icon
}()

type PanelHeader struct {
	r             *Root
	Active        st
	searchField   gel.TextField
	menuClickable Clickable
	menuButton    *IconButtonStyle
}

func (ph *PanelHeader) Init(r *Root) *PanelHeader {
	ph.r = r
	size := Dp(ph.r.th.TextSize * 2)
	ph.menuButton = &IconButtonStyle{
		Background: NRGBA{},
		Button:     &ph.menuClickable,
		Size:       size,
		Icon:       MenuIcon,
		Color:      ph.r.GetColor(col.PanelText).NRGBA(),
		Inset:      UniformInset(Dp(ph.r.th.TextSize * 2 / 3)),
	}
	ph.searchField.SingleLine = true
	ph.searchField.WrapPolicy = text.WrapWords
	ph.searchField.Submit = true
	return ph
}

func (ph *PanelHeader) Layout(g Gx) (d Dim) {
	Flex{Axis: Horizontal, Alignment: Start}.Layout(g,
		Rigid(func(g Gx) Dim {
			return ph.menuButton.Layout(g)
		}),
		Flexed(1, func(g Gx) Dim {
			return ph.searchField.Layout(g, ph.r.th, ph.r.Palette, "search")
		}),
	)
	return
}
