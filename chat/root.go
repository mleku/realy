package chat

import (
	"realy.lol/atomic"
	"realy.lol/gui/color"
	"sync"
	"realy.lol/gui/component"
	"time"
)

// Root is a widget that has a sidebar menu that becomes the whole display when
// on a small display, on large, it has a second panel that becomes the main
// display, when in small, it replaces the menu with the view widget and a top
// bar that contains a back button added to the left of the view widget's
// header.
type Root struct {
	sync.Mutex
	th *Theme
	*color.Palette
	Size Point
	*Panel
	*Chat
	Pages  map[st]Widget
	Active atomic.String
	*component.ModalLayer
	*component.ModalNavDrawer
	NavAnim component.VisibilityAnimation
}

func (r *Root) Init(th *Theme, col *color.Palette) *Root {
	r.th = th
	r.Palette = col
	r.Panel = new(Panel).Init(r)
	r.Chat = new(Chat).Init(r)
	r.ModalLayer = component.NewModal()
	nav := component.NewNav("Navigation Drawer", "This is an example.")
	r.ModalNavDrawer = component.ModalNavFrom(&nav, r.ModalLayer)
	r.NavAnim = component.VisibilityAnimation{
		State:    component.Invisible,
		Duration: time.Millisecond * 250,
	}
	return r
}

func (r *Root) Layout(g Gx) Dim {
	Fill(g.Ops, r.GetColor(color.DocBg).NRGBA())
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
		return r.ModalLayer.Layout(g, r.th)
	}
	// // at larger sizes, rigid render panel at 360px wide
	r.Panel.Small = false
	flex.Layout(g, Rigid(r.Panel.Layout), Rigid(r.Chat.Layout))
	return r.ModalLayer.Layout(g, r.th)

}
