package main

import (
	"gioui.org/layout"
	"realy.lol/gui/color"
	"realy.lol/gui/component"
	"time"
)

type Root struct {
	th *Theme
	*color.Palette
	Size Point
	*component.ModalLayer
	*component.ModalNavDrawer
	NavAnim component.VisibilityAnimation
	*Panel
	*Chat
}

func (r *Root) Init() *Root {
	r.ModalLayer = component.NewModal()
	nav := component.NewNav("Navigation Drawer", "This is an example.")
	r.ModalNavDrawer = component.ModalNavFrom(&nav, r.ModalLayer)
	r.NavAnim = component.VisibilityAnimation{
		State:    component.Invisible,
		Duration: time.Millisecond * 250,
	}
	r.Chat = new(Chat).Init(r)
	r.Panel = new(Panel).Init(r)
	return r
}

func (r *Root) Layout(g Gx) layout.Dimensions {
	Fill(g.Ops, r.Palette.GetColor(color.DocBg).NRGBA())
	flex := Flex{Axis: Horizontal}
	// return r.Panel.Layout(g)
	if r.Size.X < 720 {
		// at small sizes, only render panel
		r.Panel.Small = true
		r.Panel.Layout(g)
		return r.ModalLayer.Layout(g, r.th)
	}
	// // at larger sizes, rigid render panel at 360px wide
	r.Panel.Small = false
	flex.Layout(g, Rigid(r.Panel.Layout), Rigid(r.Chat.Layout))
	return r.ModalLayer.Layout(g, r.th)
}
