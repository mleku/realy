package main

import (
	"time"

	"realy.lol/gui/color"
	"realy.lol/gui/gel"
)

type Root struct {
	th *Theme
	*color.Palette
	Size Point
	*gel.ModalLayer
	*gel.ModalNavDrawer
	NavAnim gel.VisibilityAnimation
	*Panel
	*Chat
}

func (r *Root) Init() *Root {
	r.ModalLayer = gel.NewModal()
	nav := gel.NewNav("username", "status text")
	r.ModalNavDrawer = gel.ModalNavFrom(&nav, r.ModalLayer)
	r.NavAnim = gel.VisibilityAnimation{
		State:    gel.Invisible,
		Duration: time.Millisecond * 200,
	}
	r.Chat = new(Chat).Init(r)
	r.Panel = new(Panel).Init(r)
	return r
}

func (r *Root) Layout(g Gx) Dim {
	Fill(g.Ops, r.Palette.GetColor(color.DocBg).NRGBA())
	flex := Flex{Axis: Horizontal}
	if r.Panel.PanelHeader.menuClickable.Clicked(g) {
		log.I.F("clicked menu button")
		if !r.NavAnim.Visible() {
			r.ModalNavDrawer.Appear(g.Now)
			r.NavAnim.Disappear(g.Now)
		} else {
			r.NavAnim.Appear(g.Now)
			r.ModalNavDrawer.Disappear(g.Now)
		}
	}
	if r.Size.X < 720 {
		// at small sizes, only render panel
		r.Panel.Small = true
		r.Panel.Layout(g)
		return r.ModalLayer.Layout(g, r.th, r.Palette)
	}
	// // at larger sizes, rigid render panel at 360px wide
	r.Panel.Small = false
	flex.Layout(g, Rigid(r.Panel.Layout), Rigid(r.Chat.Layout))
	return r.ModalLayer.Layout(g, r.th, r.Palette)
}
